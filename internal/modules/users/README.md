# Контекст Users

Документ описывает доменные элементы и публичные сценарии работы в контексте `users`. Терминология синхронизирована с текущей моделью, чтобы новым разработчикам было проще ориентироваться.

## Агрегаты
- **User** (`domain.User`): содержит идентификатор, ФИО, отображаемое имя, ссылку на аватар и флаг пользовательских настроек профиля. Создаётся через `NewUser`, который очищает `DisplayName`, сбрасывает аватар и помечает профиль как некастомизированный.
- **Identity** (`domain.Identity`): связь пользователя с провайдером аутентификации (email или внешний провайдер). Может включать `SecretHash` для пароля и хранит данные провайдера.
- **RefreshToken** (`domain.RefreshToken`): запись о refresh-токене с хэшем, датой истечения, отметкой об отзыве, user-agent и IP.

## Value Objects
- **Email** (`domain.Email`): нормализует строку (trim/lowercase), проверяет базовую валидность и через `EnsureUnique` требует уникальность в хранилище идентичностей.
- **PasswordHash** (`domain.PasswordHash`): создаётся через `NewPasswordHash`, требует надёжный пароль (>= 8 символов), хранит хэш и сравнивает пароль с помощью `PasswordHasher`.
- **UserID** (`domain.UserID`): генерируется через `NewUserID`, парсится функцией `ParseUserID`, которая отвергает пустые значения и возвращает `ErrUnauthorized`.
- **ProfilePatch** (`domain.ProfilePatch`): DTO-патч профиля. В `User.ApplyPatch` каждое переданное поле очищается от пробелов, пустая строка очищает значение, а флаг `ProfileCustomized` устанавливается в `true`.

## Ключевые инварианты
- Email должен быть валидным и уникальным для провайдера `email` (`NewEmail`, `EnsureUnique`, `EnsureIdentityAvailable`).
- Пароль должен быть достаточно сильным (минимум 8 символов) перед хэшированием (`NewPasswordHash`).
- Идентичность должна быть единственной для пары (userID, provider) и для пары (provider, providerUserID); иначе возвращается `ErrIdentityAlreadyLinked` (`EnsureIdentityAvailable`).
- Аутентификация по email-паролю возможна только при наличии `SecretHash` у идентичности; иначе `ErrInvalidCredentials` (`Identity.Authenticate`).
- Refresh-токен считается действительным, если не отозван и не истёк; истёкшие токены при проверке помечаются отозванными (`RefreshToken.IsValid`, `refresh.UseCase`).
- Пустой `UserID` в запросах профиля или привязки провайдера ведёт к `ErrUnauthorized` (`ParseUserID`).

## Порты
- **UserRepository**: создать пользователя, получить по ID, обновить профиль.
- **IdentityRepository**: создать идентичность, получить по провайдеру, получить по пользователю и провайдеру.
- **RefreshTokenRepository**: создать refresh-запись, получить по хэшу, отозвать по ID.
- **PasswordHasher**: хэширование и проверка пароля.
- **AccessTokenIssuer**: выдача access-токенов с TTL.
- **UnitOfWork**: транзакционная обёртка для сценариев регистрации/логина/обновления refresh-токенов.

## Публичные use-cases и DTO
- **Register** (`register.UseCase`)
  - Вход (`register.Input`): `Email`, `Password`, `DisplayName`.
  - Выход (`login.Output`): `UserID`, `DisplayName`, `AvatarURL`, `AccessToken`, `RefreshToken`.
  - Логика: валидирует и уникализирует email, проверяет силу пароля и хэширует его, создаёт пользователя и email-идентичность, выпускает access/refresh токены в транзакции.
- **Login** (`login.UseCase`)
  - Вход (`login.Input`): `Email`, `Password`.
  - Выход (`login.Output`): `UserID`, ФИО (`FirstName`, `LastName`, `MiddleName`), `DisplayName`, `AvatarURL`, `AccessToken`, `RefreshToken`.
  - Логика: ищет email-идентичность, проверяет пароль, поднимает пользователя, выдаёт новые access/refresh токены и сохраняет refresh-запись.
- **Refresh** (`refresh.UseCase`)
  - Вход (`refresh.Input`): `RefreshToken` (сырой).
  - Выход (`refresh.Output`): новый `AccessToken`, `RefreshToken`.
  - Логика: ищет запись по хэшу токена, проверяет срок/отзыв, ревокирует старый и создаёт новый refresh в транзакции, выдаёт новый access.
- **GetMe** (`profile.GetUseCase`)
  - Вход (`profile.GetInput`): `UserID`.
  - Выход (`profile.Output`): `UserID`, ФИО, `DisplayName`, `AvatarURL`.
  - Логика: валидирует `UserID`, читает пользователя; при отсутствии — `ErrUnauthorized`.
- **UpdateProfile** (`profile.UpdateUseCase`)
  - Вход (`profile.UpdateInput`): `UserID` и опциональные поля `FirstName`, `LastName`, `MiddleName`, `DisplayName`, `AvatarURL` (nil — без изменений, пустая строка — очистить поле).
  - Выход (`profile.Output`): `UserID`, ФИО, `DisplayName`, `AvatarURL`.
  - Логика: валидирует `UserID`, загружает пользователя, применяет patch через `ApplyPatch`, сохраняет и возвращает обновлённые данные.
- **LinkProvider** (`link.UseCase`)
  - Вход (`link.Input`): `UserID`, `Provider`, `ProviderUserID`.
  - Выход (`link.Output`): `Linked` (bool).
  - Логика: валидирует `UserID`, проверяет уникальность идентичности для пользователя и провайдера, создаёт внешнюю идентичность.

## Терминология
- **Identity** — привязка пользователя к способу аутентификации (email/password или внешний провайдер), хранит идентификаторы провайдера и при необходимости секрет.
- **Access Token** — короткоживущий токен авторизации, выдаётся через `AccessTokenIssuer`.
- **Refresh Token** — длинноживущий токен обновления; в хранилище сохраняется только хэш (`HashToken`), а пользователю отдаётся сырой токен (`NewRefreshToken`).
- **Profile Patch** — частичное обновление профиля: отсутствие поля означает «оставить как есть», пустая строка — «очистить значение».
- **Provider** — строковый идентификатор системы аутентификации (например, `email` или внешний OAuth-провайдер).
