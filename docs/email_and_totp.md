# Email delivery and TOTP testing

## SMTP configuration

Set the following environment variables to enable real email sending via SMTP:

- `SMTP_HOST` – SMTP server hostname
- `SMTP_PORT` – port (defaults to 587)
- `SMTP_USERNAME` / `SMTP_PASSWORD` – credentials (optional if your relay is open or uses IP allow-list)
- `SMTP_FROM` – address that will appear in the From header (required to send)
- `SMTP_USE_TLS` – `true` to dial with TLS and fall back to STARTTLS when possible (default), `false` to use plaintext/STARTTLS only
- `SMTP_TIMEOUT` – dial timeout (default `10s`)

If `SMTP_HOST` or `SMTP_FROM` is empty, the outbox worker will only log email events instead of sending them.

The application continues to emit email confirmation and password reset events into the outbox. The worker (`outbox.Worker`) now fans out those events to the SMTP mailer and the logger. Run the app normally (`go run cmd/app/main.go`) to keep the worker running alongside the HTTP server.

HTML and plain-text templates for confirmation and password reset live under `internal/modules/users/infrastructure/events/templates`. You can edit them and rebuild the app; the templates are embedded at compile time so deployment does not need to ship separate template files.

## Testing email flows locally

1. Start the app with SMTP variables pointing to a test SMTP server (e.g., MailHog, Mailtrap, or your provider’s sandbox account).
2. Register a user. If `AUTH_REQUIRE_EMAIL_CONFIRMATION=true`, the app will enqueue a confirmation code; the worker will send it via SMTP.
3. Use `/auth/confirm` with the received code to activate the account.
4. Request a password reset through `/auth/password/reset` and confirm with `/auth/password/confirm` using the emailed code.

## Testing OTP / Google Authenticator

1. Authenticate and call `POST /auth/2fa/setup` (bearer token required). The response contains `secret` and an `uri` field suitable for QR import.
2. In Google Authenticator (or any TOTP app), add an account using the `uri` (scan the QR or paste the URI manually) or enter the `secret` manually.
3. Confirm setup by calling `POST /auth/2fa/confirm` with JSON `{ "code": "123456" }` using the code shown in your app.
4. Future logins will require the `otp_code` field in the `/auth/login` body. Codes rotate every 30 seconds; you can validate them locally with your authenticator.
5. To disable, call `POST /auth/2fa/disable` with the current code.

Routes are defined under `internal/platform/http/users/routes.go`, and handler shapes are in `internal/platform/http/users/dto/auth.go` and `internal/platform/http/users/handler.go`.

