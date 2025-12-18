package contracts

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

// BuildSchemaForStruct builds a simple JSON schema representation for the
// provided struct value. The generated schema intentionally covers only the
// primitives used by the DTOs and events in tests (objects with string, boolean
// and date-time properties).
func BuildSchemaForStruct(value any) (map[string]any, error) {
	t := reflect.TypeOf(value)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", t.Kind())
	}

	props := make(map[string]any)
	var required []string

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" { // unexported
			continue
		}

		tag := f.Tag.Get("json")
		name, opts := parseJSONTag(tag, f.Name)
		if name == "-" {
			continue
		}

		// Flatten embedded structs to mirror Go's JSON encoding rules.
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			sub, err := BuildSchemaForStruct(reflect.New(f.Type).Elem().Interface())
			if err != nil {
				return nil, err
			}
			mergeSchemas(props, &required, sub)
			continue
		}

		propSchema, err := typeSchema(f.Type)
		if err != nil {
			return nil, err
		}
		if propSchema == nil {
			continue
		}

		props[name] = propSchema
		if !opts.contains("omitempty") && f.Type.Kind() != reflect.Ptr {
			required = append(required, name)
		}
	}

	schema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties":           props,
	}
	if len(required) > 0 {
		sort.Strings(required)
		schema["required"] = required
	}
	return schema, nil
}

// ValidateSchema validates a decoded JSON document against the given schema.
// The validator intentionally supports a small subset of JSON Schema required
// by the tests: type checking for objects, strings, booleans, required
// properties, nullable markers, format=date-time and additionalProperties.
func ValidateSchema(schema map[string]any, document any) error {
	return validateNode(schema, schema, document, "$")
}

// LoadSchema parses the JSON schema file contents.
func LoadSchema(data []byte) (map[string]any, error) {
	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("schema decode: %w", err)
	}
	return schema, nil
}

// Canonicalize renders the schema map into stable, indented JSON suitable for
// golden files.
func Canonicalize(schema map[string]any) ([]byte, error) {
	return json.MarshalIndent(schema, "", "  ")
}

func typeSchema(t reflect.Type) (map[string]any, error) {
	nullable := false
	for t.Kind() == reflect.Ptr {
		nullable = true
		t = t.Elem()
	}

	var schema map[string]any
	switch {
	case t.PkgPath() == "time" && t.Name() == "Time":
		schema = map[string]any{"type": "string", "format": "date-time"}
	case t.Kind() == reflect.Struct:
		nested, err := BuildSchemaForStruct(reflect.New(t).Elem().Interface())
		if err != nil {
			return nil, err
		}
		schema = nested
	case t.Kind() == reflect.String:
		schema = map[string]any{"type": "string"}
	case t.Kind() == reflect.Bool:
		schema = map[string]any{"type": "boolean"}
	default:
		return nil, fmt.Errorf("unsupported field type %s", t.Kind())
	}

	if nullable {
		schema = copyMap(schema)
		schema["nullable"] = true
	}
	return schema, nil
}

func mergeSchemas(props map[string]any, required *[]string, schema map[string]any) {
	nestedProps, _ := schema["properties"].(map[string]any)
	for k, v := range nestedProps {
		props[k] = v
	}

	switch req := schema["required"].(type) {
	case []any:
		for _, r := range req {
			if name, ok := r.(string); ok {
				*required = append(*required, name)
			}
		}
	case []string:
		*required = append(*required, req...)
	}
}

func validateNode(node, root map[string]any, value any, path string) error {
	if ref, ok := node["$ref"].(string); ok {
		target, err := resolveRef(ref, root)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		return validateNode(target, root, value, path)
	}

	if nullable, _ := node["nullable"].(bool); nullable && value == nil {
		return nil
	}

	typ, _ := node["type"].(string)
	switch typ {
	case "object":
		obj, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("%s: expected object", path)
		}

		props, _ := node["properties"].(map[string]any)
		required := stringSlice(node["required"])
		addProps := true
		if ap, ok := node["additionalProperties"].(bool); ok {
			addProps = ap
		}

		for _, key := range required {
			if _, ok := obj[key]; !ok {
				return fmt.Errorf("%s: missing required property %q", path, key)
			}
		}

		for key, v := range obj {
			if prop, ok := props[key]; ok {
				propNode, _ := prop.(map[string]any)
				if err := validateNode(propNode, root, v, path+"."+key); err != nil {
					return err
				}
				continue
			}
			if !addProps {
				return fmt.Errorf("%s: unexpected property %q", path, key)
			}
		}
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%s: expected string", path)
		}
		if fmtVal, ok := node["format"].(string); ok && fmtVal == "date-time" {
			if _, err := time.Parse(time.RFC3339, value.(string)); err != nil {
				return fmt.Errorf("%s: invalid date-time", path)
			}
		}
		if enums, ok := node["enum"].([]any); ok {
			if !contains(enums, value) {
				return fmt.Errorf("%s: value %v not in enum", path, value)
			}
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("%s: expected boolean", path)
		}
	default:
		return fmt.Errorf("%s: unsupported type %q", path, typ)
	}

	return nil
}

func resolveRef(ref string, root map[string]any) (map[string]any, error) {
	if !strings.HasPrefix(ref, "#/") {
		return nil, fmt.Errorf("only local refs are supported: %s", ref)
	}

	parts := strings.Split(strings.TrimPrefix(ref, "#/"), "/")
	var current any = root
	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, errors.New("invalid ref path")
		}
		next, ok := m[part]
		if !ok {
			return nil, fmt.Errorf("reference not found: %s", ref)
		}
		current = next
	}

	node, ok := current.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("reference target is not an object: %s", ref)
	}
	return node, nil
}

func parseJSONTag(tag, fallback string) (name string, opts tagOptions) {
	if tag == "" {
		return fallback, nil
	}
	parts := strings.Split(tag, ",")
	return parts[0], parts[1:]
}

type tagOptions []string

func (o tagOptions) contains(opt string) bool {
	for _, v := range o {
		if v == opt {
			return true
		}
	}
	return false
}

func stringSlice(v any) []string {
	raw, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

func contains(list []any, value any) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func copyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
