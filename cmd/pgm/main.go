package main

import (
	"context"
	"flag"
	seoLogger "github.com/quadgod/seo/pkg/logger"
	"github.com/quadgod/seo/pkg/pgm"
	"github.com/quadgod/seo/pkg/pgm/cli"
	"log"
	"log/slog"
	"os"
)

func main() {
	logLevel := new(slog.LevelVar)
	logLevel.Set(slog.LevelInfo)
	logger := seoLogger.CreateLogger(logLevel)

	flags := new(pgm.Flags)

	flag.StringVar(&flags.Command, "command", "", "command")
	flag.StringVar(&flags.MigrationsDir, "migrationsDir", "", "path to migrations directory")
	flag.StringVar(&flags.MigrationsTable, "migrationsTable", "migrations", "name of migrations table")
	flag.StringVar(&flags.MigrationsTableSchema, "migrationsTableSchema", "", "migrations table schema")
	flag.StringVar(&flags.MigrationName, "migrationName", "", "migration name")
	flag.StringVar(&flags.ConnectionString, "connectionString", os.Getenv("DATABASE_URL"), "connection string")
	flag.StringVar(&flags.Priority, "priority", string(pgm.PriorityFS), "db or fs migrations priority")

	flag.Parse()

	if err := flags.Validate(); err != nil {
		log.Fatalf("arguments validation errors: %v", err)
	}

	opts := flags.ToMigratorOptions()

	switch opts.Command {
	case pgm.CommandCreate:
		mig, err := cli.Create(&opts)

		if err != nil {
			log.Fatalf("create migration errors: %v", err)
		}

		logger.Info("migration created", "name", mig.Name, "up", mig.Up, "down", mig.Down)
	case pgm.CommandMigrate:
		res, err := cli.Migrate(context.Background(), &opts)
		if err != nil {
			log.Fatalf("errors occurs during migrate command execution: %v", err)
		}

		for _, r := range res {
			logger.Info("migration", r.MigrationName, r.Status)
		}
	case pgm.CommandDown:
		res, err := cli.Down(context.Background(), &opts)
		if err != nil {
			log.Fatalf("errors occurs during down command execution: %v", err)
		}

		logger.Info("migration", res.MigrationName, res.Status)
	}
}
