# Auth routes overview

This document summarizes the user-facing `/auth` routes, what each endpoint does, and when to call it during login/verification flows.

## Quick reference

| Path | Method | Purpose |
| --- | --- | --- |
| `/auth/register` | POST | Create a new user with email/password. |
| `/auth/login` | POST | Start a session; may return tokens or an auth challenge. |
| `/auth/refresh` | POST | Exchange a refresh token for new access/refresh tokens. |
| `/auth/confirm/request` | POST | Send an email confirmation code to a specific address. |
| `/auth/confirm` | POST | Confirm email with `email + code` (non-challenge flow). |
| `/auth/challenge/status` | POST | Fetch the latest state of an auth challenge. |
| `/auth/challenge/verify-totp` | POST | Submit a TOTP code for an auth challenge. |
| `/auth/challenge/resend-email` | POST | Resend a challenge email verification token. |
| `/auth/challenge/confirm-email` | POST | Confirm email for an auth challenge using `challenge_id + token`. |
| `/auth/password/reset` | POST | Request a password reset email. |
| `/auth/password/confirm` | POST | Set a new password using a reset code. |
| `/auth/telegram` | POST | Log in via Telegram login data. |
| `/auth/link` | POST | Link an external provider to the signed-in account. |
| `/auth/2fa/setup` | POST | Begin TOTP enrollment (requires JWT). |
| `/auth/2fa/confirm` | POST | Confirm TOTP enrollment with a code. |
| `/auth/2fa/disable` | POST | Disable TOTP for the signed-in user. |
| `/me` | GET | Fetch the current profile (requires JWT). |
| `/me` | PATCH | Update profile fields (requires JWT). |

## Login and challenges

`POST /auth/login` accepts `{ "email", "password", "otp_code"? }`. If the account is clear to sign in, the response contains `access_token` and `refresh_token`. When the account needs extra steps (blocked, email not verified, or TOTP enabled), the response has `status: "challenge_required"` plus a `challenge` block describing required steps, attempts left, and expiration. Tokens are only issued once all required steps are completed via the challenge endpoints.

Use the `challenge_id` from this response with:

- `POST /auth/challenge/status` – poll for updated state if the client is unsure which steps remain.
- `POST /auth/challenge/verify-totp` with `{ "challenge_id", "otp_code" }` – submit a TOTP code when `totp` is required.
- `POST /auth/challenge/resend-email` with `{ "challenge_id" }` – trigger another email if `email_verification` is required.
- `POST /auth/challenge/confirm-email` with `{ "challenge_id", "token" }` – confirm the emailed token; when successful, the challenge completes and tokens are returned.

## Email confirmation: regular vs challenge

There are two email confirmation flows:

- `/auth/confirm/request` and `/auth/confirm` work by email address. They send and validate a short code for cases like initial account activation or manual re-sends outside of an active login challenge. The confirm request body is `{ "email", "code" }`.
- `/auth/challenge/resend-email` and `/auth/challenge/confirm-email` belong to the login challenge flow. They are keyed by `challenge_id` (not raw email) to avoid leaking addresses and to tie verification to the pending sign-in. The confirm body is `{ "challenge_id", "token" }`.

## Password reset

`POST /auth/password/reset` sends a reset code to the given email. `POST /auth/password/confirm` accepts `{ "email", "code", "password" }` to set a new password.

## Refresh and logout

`POST /auth/refresh` exchanges a refresh token for new tokens. The project does not expose a logout endpoint; clients should discard tokens locally and can rotate refresh tokens via `/auth/refresh`.

## Telegram login

`POST /auth/telegram` takes the Telegram `init_data` payload and signs the user in if the Telegram identity is valid and linked.

## Two-factor lifecycle

All TOTP management endpoints require a bearer token:

- `POST /auth/2fa/setup` returns a secret and otpauth URI for enrollment.
- `POST /auth/2fa/confirm` expects `{ "code" }` from the authenticator app to finalize enrollment.
- `POST /auth/2fa/disable` expects `{ "code" }` to remove TOTP from the account.

## Profile

Authenticated users can fetch or update their profile via `GET /me` and `PATCH /me`. Profile fields include names and avatar URL.
