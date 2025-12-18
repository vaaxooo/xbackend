package events

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vaaxooo/xbackend/internal/test/contracts"
)

func TestUserRegisteredSchemaMatchesStruct(t *testing.T) {
	schema, err := contracts.BuildSchemaForStruct(UserRegistered{})
	if err != nil {
		t.Fatalf("build schema: %v", err)
	}

	canonical, err := contracts.Canonicalize(schema)
	if err != nil {
		t.Fatalf("canonicalize: %v", err)
	}

	raw, err := os.ReadFile(eventSchemaPath())
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}
	fileSchema, err := contracts.LoadSchema(raw)
	if err != nil {
		t.Fatalf("decode schema: %v", err)
	}
	fileCanonical, err := contracts.Canonicalize(fileSchema)
	if err != nil {
		t.Fatalf("canonicalize file: %v", err)
	}

	if string(canonical) != string(fileCanonical) {
		t.Fatalf("schema mismatch\nexpected:\n%s\nactual:\n%s", string(canonical), string(fileCanonical))
	}
}

func TestAsyncAPILinksSchema(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("user_registered.asyncapi.yaml"))
	if err != nil {
		t.Fatalf("read asyncapi: %v", err)
	}
	if !strings.Contains(string(raw), "./user_registered.schema.json") {
		t.Fatalf("asyncapi does not reference schema file")
	}
}

func eventSchemaPath() string {
	return filepath.Join("user_registered.schema.json")
}
