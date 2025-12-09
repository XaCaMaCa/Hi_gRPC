package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"

	// Драйвер для чтения миграций из файловой системы
	_ "github.com/golang-migrate/migrate/v4/source/file"
	// Драйвер для SQLite (чистый Go, без CGO)
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
)

func main() {
	var storagePath, migrationsPath, migrationsTable string
	var forceVersion int

	flag.StringVar(&storagePath, "storage-path", "", "path to storage file")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations file")
	flag.StringVar(&migrationsTable, "migrations-table", "", "name of migrations table")
	flag.IntVar(&forceVersion, "force", -1, "force set version (use to fix dirty state)")
	flag.Parse()

	if storagePath == "" {
		panic("storage path is required")
	}
	if migrationsPath == "" {
		panic("migrations path is required")
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf("sqlite://%s?x-migrations-table=%s", storagePath, migrationsTable),
	)
	if err != nil {
		panic(err)
	}

	// Принудительно установить версию (для исправления dirty state)
	if forceVersion != -1 {
		if err := m.Force(forceVersion); err != nil {
			panic(fmt.Sprintf("failed to force version: %v", err))
		}
		log.Printf("forced version to %d", forceVersion)
		return // После force выходим
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		panic(err)
	}
	log.Println("migrations applied successfully")
}
