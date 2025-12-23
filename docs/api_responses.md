# API responses and error contracts

Документ описывает стабильные ответы HTTP для `api/v1` пользователя. Формат ошибок и успеха единообразный, чтобы фронтенд мог
стабильно обрабатывать stateless ответы.

## Общий формат ошибок
Все ошибки возвращаются с кодом статуса из таблицы ниже и телом:

```json
{
  "error": {
    "code": "<machine_code>",
    "message": "<human_readable>"
  }
}
```

Частые коды и статусы:

| Код | HTTP статус | Комментарий |
| --- | --- | --- |
| `validation_error` | `400` | Некорректный JSON/валидация email/пароля/аватара и т.п. |
| `email_already_used` | `409` | Email уже зарегистрирован. |
| `identity_already_linked` | `409` | Внешний провайдер уже привязан. |
| `invalid_credentials` | `401` | Неверный логин/пароль. |
| `refresh_token_invalid` | `401` | Истёкший/отозванный refresh токен. |
| `email_not_verified` | `403` | Требуется подтверждение email. |
| `two_factor_required` | `401` | Нужна верификация 2FA перед выдачей токенов. |
| `invalid_two_factor` | `401` | Неверный код 2FA. |
| `two_factor_already_enabled` | `409` | Попытка повторно включить 2FA. |
| `too_many_requests` | `429` | Сработал rate limiting. |
| `unauthorized` | `401` | Нет или истёкший access-токен. |
| `internal_error` | `500` | Непредвиденная ошибка сервера. |

## Формат успешного ответа без данных
Для действий без отдельного payload возвращается единый формат:

```json
{ "status": "ok", "message": "<описание действия>" }
```

## Auth маршруты
- `POST /api/v1/auth/register` → `201` + профиль и токены.
- `POST /api/v1/auth/login` → `200` + профиль и токены либо challenge.
- `POST /api/v1/auth/refresh` → `200` + новые токены.
- `POST /api/v1/auth/confirm` → `200` + профиль и токены после кода из email.
- `POST /api/v1/auth/confirm/request` → `202` + `{status,message}` о запросе письма.
- `POST /api/v1/auth/password/reset` → `202` + `{status,message}` о запросе письма.
- `POST /api/v1/auth/password/confirm` → `200` + `{status,message}` о смене пароля.
- `POST /api/v1/auth/password/change` → `200` + `{status,message}` при успешной смене.
- `POST /api/v1/auth/challenge/status` → `200` + профиль/токены или challenge.
- `POST /api/v1/auth/challenge/verify-totp` → `200` + профиль/токены.
- `POST /api/v1/auth/challenge/resend-email` → `200` + профиль/токены.
- `POST /api/v1/auth/challenge/confirm-email` → `200` + профиль/токены.
- `POST /api/v1/auth/telegram|google|apple` → `200` + профиль/токены.
- `POST /api/v1/auth/link` → `200` + `{ linked: bool }` (JWT обязателен).
- `POST /api/v1/auth/2fa/setup` → `200` + секрет и QR.
- `POST /api/v1/auth/2fa/confirm` → `200` + `{status,message}` о включении 2FA.
- `POST /api/v1/auth/2fa/disable` → `200` + `{status,message}` о выключении 2FA.
- `GET /api/v1/auth/sessions` → `200` + список сессий.
- `POST /api/v1/auth/sessions/revoke` → `200` + `{status,message}` о ревоке конкретной сессии.
- `POST /api/v1/auth/sessions/revoke-others` → `200` + `{status,message}` о ревоке остальных сессий.

## Профиль
- `GET /api/v1/me` → `200` + профиль.
- `PATCH /api/v1/me` → `200` + обновлённый профиль.

Маршруты профиля и защищённых auth endpoint'ов всегда требуют валидный access-токен и при его отсутствии отвечают `401 unauthorized` в указанном формате ошибки.
