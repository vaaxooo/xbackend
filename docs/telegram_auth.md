# Telegram Web App Authentication Usage

The backend exposes a dedicated endpoint for Telegram Web App login/registration.

## Endpoint
- **Path:** `POST /api/v1/auth/telegram`
- **Body:** JSON with a single field `init_data` containing the raw `initData` string provided by Telegram Web Apps (the query-string payload passed to your web app inside Telegram).

```json
{
  "init_data": "query_id=AAE...&user=%7B%22id%22%3A12345%2C...%7D&auth_date=1716500000&hash=..."
}
```

The handler validates the HMAC signature (`hash`), checks `auth_date` against `TELEGRAM_INIT_DATA_TTL`, and then either logs in the linked user or auto-registers a new account with the Telegram profile data.

## Response
On success the endpoint returns the same payload as the email/password login flow:

```json
{
  "user_id": "<uuid>",
  "first_name": "...",
  "last_name": "...",
  "middle_name": "...",
  "display_name": "...",
  "avatar_url": "https://...",
  "access_token": "<jwt>",
  "refresh_token": "<opaque-token>"
}
```

HTTP status codes mirror other auth endpoints: `200` on success, `400/401` for invalid data, and `500` for server errors.

## Example cURL
```bash
curl -X POST https://<host>/api/v1/auth/telegram \
  -H 'Content-Type: application/json' \
  -d '{"init_data":"<raw initData from Telegram>"}'
```

## Configuration
Telegram verification uses the following environment variables (see `config`):
- `TELEGRAM_BOT_TOKEN` – bot token used to derive the HMAC secret for `init_data` validation (required).
- `TELEGRAM_INIT_DATA_TTL` – allowed age for `auth_date` (default `24h`).

> The application will fail to start if `TELEGRAM_BOT_TOKEN` is missing because the init data signature cannot be verified.
