package pgm

import (
	"errors"
	"fmt"
	"regexp"
)

type Flags struct {
	Priority              string
	Command               string
	MigrationName         string
	MigrationsDir         string
	MigrationsTableSchema string
	MigrationsTable       string
	ConnectionString      string
}

func (f *Flags) ToMigratorOptions() MigratorOptions {
	return MigratorOptions{
		Priority:              Priority(f.Priority),
		Command:               Command(f.Command),
		MigrationName:         f.MigrationName,
		MigrationsDir:         f.MigrationsDir,
		MigrationsTableSchema: f.MigrationsTableSchema,
		MigrationsTable:       f.MigrationsTable,
		ConnectionString:      f.ConnectionString,
	}
}

func (f *Flags) Validate() error {
	if f.MigrationsDir == "" {
		return errors.New("migrations dir is required")
	}

	cmd := Command(f.Command)
	priority := Priority(f.Priority)

	switch priority {
	case PriorityDB:
	case PriorityFS:
		break
	default:
		return fmt.Errorf("invalid priority. valid values \"%s\" or \"%s\"", PriorityDB, PriorityFS)
	}

	switch cmd {
	case CommandCreate:
		if match, err := regexp.MatchString(`^[a-zA-Z0-9_]+$`, f.MigrationName); err == nil && !match {
			return fmt.Errorf("migration name might contain only letters, numbers and \"_\" symbol")
		}
	case CommandDown:
	case CommandMigrate:
		if match, err := regexp.MatchString(`^[a-zA-Z0-9_]+$`, f.MigrationsTableSchema); err == nil && !match {
			return fmt.Errorf("migrations table schema might contain only letters, numbers and \"_\" symbol")
		}

		if match, err := regexp.MatchString(`^[a-zA-Z0-9_]+$`, f.MigrationsTable); err == nil && !match {
			return fmt.Errorf("migrations table might contain only letters, numbers and \"_\" symbol")
		}

		if f.ConnectionString == "" {
			return errors.New("connection string is required")
		}
	default:
		return fmt.Errorf("invalid command. valid cli \"%s\", \"%s\"", CommandCreate, CommandMigrate)
	}

	return nil
}
