package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	var (
		migrationsPath string
		dbDSN          string
		table          string

		direction string
		steps     int
		force     int

		showStatus  bool
		showVersion bool
	)

	flag.StringVar(&migrationsPath, "path", "", "Path to migrations folder (e.g. ./migrations/users)")
	flag.StringVar(&dbDSN, "dsn", "", "Database DSN (defaults to DB_DSN env)")
	flag.StringVar(&table, "table", "schema_migrations", "Migrations table name")

	flag.StringVar(&direction, "direction", "up", "Migration direction: up or down (ignored if -steps is set)")
	flag.IntVar(&steps, "steps", 0, "Run N migration steps (positive=up, negative=down). Overrides -direction")
	flag.IntVar(&force, "force", -1, "Force set migration version (use carefully)")

	flag.BoolVar(&showStatus, "status", false, "Print current migration version and dirty state")
	flag.BoolVar(&showVersion, "version", false, "Print current migration version only")
	flag.Parse()

	if migrationsPath == "" {
		log.Fatal("migrations path is required")
	}

	if dbDSN == "" {
		dbDSN = os.Getenv("DB_DSN")
		if dbDSN == "" {
			log.Fatal("DB_DSN is required")
		}
	}

	m := mustMigrate(migrationsPath, dbDSN, table)

	// Status/version are read-only operations.
	if showStatus || showVersion {
		v, dirty, err := m.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				if showStatus {
					fmt.Println("version: none, dirty: false")
				} else {
					fmt.Println("version: none")
				}
				return
			}
			log.Fatal(err)
		}

		if showStatus {
			fmt.Printf("version: %d, dirty: %t\n", v, dirty)
		} else {
			fmt.Printf("version: %d\n", v)
		}
		return
	}

	// Force is a destructive operation. It sets the version without running migrations.
	if force >= 0 {
		if err := m.Force(force); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("forced version to %d\n", force)
		return
	}

	// Steps override direction.
	if steps != 0 {
		if err := m.Steps(steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}
		fmt.Printf("steps applied: %d\n", steps)
		return
	}

	switch direction {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}
		fmt.Println("migrations applied (up)")
	case "down":
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}
		fmt.Println("migrations rolled back (down)")
	default:
		log.Fatal("direction must be 'up' or 'down'")
	}
}

func mustMigrate(migrationsPath, dbDSN, table string) *migrate.Migrate {
	db, err := sql.Open("pgx", dbDSN)
	if err != nil {
		log.Fatal(err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: table,
	})
	if err != nil {
		_ = db.Close()
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		_ = db.Close()
		log.Fatal(err)
	}

	// Note: migrate.Migrate holds the database instance; it will be closed on m.Close().
	// We rely on process exit here. If you want explicit cleanup, we can add defer m.Close().
	return m
}
