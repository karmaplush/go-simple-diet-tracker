package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/go-playground/validator"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Config struct {
	StoragePath     string `validate:"required"`
	MigrationsPath  string `validate:"required"`
	MigrationsTable string
}

func main() {
	var cfg Config

	flag.StringVar(&cfg.StoragePath, "storage-path", "", "path to storage")
	flag.StringVar(&cfg.MigrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(
		&cfg.MigrationsTable,
		"migrations-table",
		"migrations",
		"name of migrations table",
	)
	flag.Parse()

	if err := validator.New().Struct(cfg); err != nil {
		panic("-storage-path or/and -migrations-path flag was not provided")
	}

	m, err := migrate.New(
		"file://"+cfg.MigrationsPath,
		fmt.Sprintf("sqlite3://%s?x-migrations-table=%s", cfg.StoragePath, cfg.MigrationsTable),
	)

	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}

		panic(err)
	}

	fmt.Println("migrations was successful")
}
