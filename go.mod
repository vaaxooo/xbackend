module github.com/vaaxooo/xbackend

go 1.24.0

toolchain go1.24.11

require (
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/go-chi/chi/v5 v5.2.3
	github.com/golang-migrate/migrate/v4 v4.19.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.6
        golang.org/x/crypto v0.45.0
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/lib/pq v1.10.9 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/text v0.31.0 // indirect
)

replace github.com/DATA-DOG/go-sqlmock => ./internal/test/sqlmock
