# UserRegistered event contract quickstart

The `user_registered.asyncapi.yaml` and `user_registered.schema.json` files describe the `UserRegistered` integration event. Use these commands to inspect or verify the contract.

## Render the AsyncAPI doc to HTML

From the repository root, use the AsyncAPI CLI to generate a preview HTML file:

```bash
npx -y @asyncapi/cli render internal/modules/users/application/events/user_registered.asyncapi.yaml -o /tmp/user_registered.html
xdg-open /tmp/user_registered.html  # or open the file in your browser
```

## Validate the event schema against the Go struct

Run the existing contract test to ensure the schema stays in sync with the event payload struct:

```bash
go test ./internal/modules/users/application/events -count=1
```

## Files to know
- `user_registered.asyncapi.yaml` — AsyncAPI contract referencing the payload schema.
- `user_registered.schema.json` — JSON Schema for the event payload, validated in tests.
