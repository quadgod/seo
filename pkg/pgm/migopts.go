package pgm

import (
	"fmt"
)

type Priority string

const (
	PriorityDB Priority = "db"
	PriorityFS Priority = "fs"
)

type Command string

const (
	CommandCreate  Command = "create"
	CommandMigrate Command = "migrate"
	CommandDown    Command = "down"
)

type MigratorOptions struct {
	Priority              Priority
	Command               Command
	MigrationName         string
	MigrationsDir         string
	MigrationsTableSchema string
	MigrationsTable       string
	ConnectionString      string
}

func (o *MigratorOptions) MigrationsTableNameWithSchema() string {
	return fmt.Sprintf("%s.%s", o.MigrationsTableSchema, o.MigrationsTable)
}
