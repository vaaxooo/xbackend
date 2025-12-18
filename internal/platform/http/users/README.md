# Users HTTP contract quickstart

This folder contains the OpenAPI 3.1 contract (`openapi.yaml`) and the JSON Schemas used by the users HTTP handlers. Use the steps below to preview the docs or validate responses.

## Preview the OpenAPI spec in a browser

From the repository root:

- ReDoc (fastest single command):

  ```bash
  npx -y @redocly/cli preview-docs internal/platform/http/users/openapi.yaml --host 0.0.0.0 --port 8080
  ```

  Then open http://localhost:8080.

- Swagger UI (served from Docker):

  ```bash
  docker run --rm -p 8080:8080 -v "$(pwd)/internal/platform/http/users/openapi.yaml:/usr/share/nginx/html/openapi.yaml:ro" swaggerapi/swagger-ui \
    sh -c 'sed -i "s|https://petstore.swagger.io/v2/swagger.json|/openapi.yaml|" /usr/share/nginx/html/swagger-initializer.js && /usr/share/nginx/html/start.sh'
  ```

  Then open http://localhost:8080 to view the Swagger UI.

## Validate schemas and handler contracts

Contract tests already exercise the handlers against the JSON Schemas:

```bash
go test ./internal/platform/http/users -count=1
```

The test suite will fail if a DTO or HTTP response stops matching `openapi.yaml` or the schemas in `internal/platform/http/users/schemas`.

## Where the schemas live
- `openapi.yaml` — HTTP contract for `/api/v1` users endpoints.
- `schemas/*.schema.json` — JSON Schemas reused by the OpenAPI paths and handler tests.

These files act as the single source of truth for clients and for generated SDKs.
