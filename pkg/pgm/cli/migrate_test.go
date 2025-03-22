package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/quadgod/seo/pkg/pgm"
	"github.com/quadgod/seo/pkg/pgm/db"
	"github.com/quadgod/seo/pkg/pgm/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"path"
	"testing"
	"time"
)

func genUpSql(table string) string {
	return fmt.Sprintf(`
		CREATE SCHEMA IF NOT EXISTS test;
		CREATE TABLE test.%s (
			"id"            SERIAL PRIMARY KEY,
			"active"        BOOLEAN
		);
	`, table)
}

func genDownSql(table string) string {
	return fmt.Sprintf("DROP TABLE test.%s;", table)
}

func genMigration(migrationsDir string, migrationName string, tableName string) (*pgm.Migration, error) {
	migration, err := fs.CreateMigration(migrationsDir, migrationName)

	if err != nil {
		return nil, err
	}

	upSql := genUpSql(tableName)
	if err = os.WriteFile(migration.Up, []byte(upSql), 0755); err != nil {
		return nil, err
	}

	downSql := genDownSql(tableName)
	if err = os.WriteFile(migration.Down, []byte(downSql), 0755); err != nil {
		return nil, err
	}

	return migration, nil
}

func Test_Migrate(t *testing.T) {
	ctx := context.Background()
	migrationsDir := path.Join(t.TempDir(), "/pgm_migrate_test")

	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:alpine3.19"),
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)

	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %s", err)
		}
		if err := os.RemoveAll(migrationsDir); err != nil {
			t.Errorf("failed to remove migrations dir: %s", err)
		}
	})

	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")

	if err != nil {
		t.Fatalf("unable to build connection string")
	}

	t.Run("should apply 2 migrations", func(t *testing.T) {
		opts := new(pgm.MigratorOptions)
		opts.ConnectionString = connStr
		opts.Command = pgm.CommandMigrate
		opts.Priority = pgm.PriorityDB
		opts.MigrationsTableSchema = "detmir_jobs"
		opts.MigrationsTable = "migrations"
		opts.MigrationsDir = migrationsDir

		_, err = genMigration(opts.MigrationsDir, "first_migration", "table1")
		if err != nil {
			t.Fatal(err)
		}

		_, err = genMigration(opts.MigrationsDir, "second_migration", "table2")
		if err != nil {
			t.Fatal(err)
		}

		appliedMigrations1, err := Migrate(ctx, opts)
		require.Nil(t, err)
		require.Len(t, appliedMigrations1, 2)
		require.Contains(t, appliedMigrations1[0].MigrationName, "first_migration")
		require.Equal(t, pgm.APPLIED, appliedMigrations1[0].Status)
		require.Contains(t, appliedMigrations1[1].MigrationName, "second_migration")
		require.Equal(t, pgm.APPLIED, appliedMigrations1[1].Status)

		appliedMigrations2, err := Migrate(ctx, opts)
		require.Nil(t, err)

		require.Len(t, appliedMigrations2, 0)

		pool, err := db.Connect(ctx, connStr)
		require.Nil(t, err)

		defer pool.Close()

		migrationSchemaTables, err := db.Tables(ctx, pool, opts.MigrationsTableSchema)
		require.Nil(t, err)

		require.Len(t, migrationSchemaTables, 1)
		require.Equal(t, opts.MigrationsTableSchema, migrationSchemaTables[0].TableSchema)
		require.Equal(t, opts.MigrationsTable, migrationSchemaTables[0].TableName)

		testSchemaTables, err := db.Tables(ctx, pool, "test")
		if assert.Nil(t, err) {
			require.Len(t, testSchemaTables, 2)
			require.Equal(t, "table1", testSchemaTables[0].TableName)
			require.Equal(t, "table2", testSchemaTables[1].TableName)
		}

		err = db.ResetForTests(ctx, pool, opts.MigrationsTableNameWithSchema())
		if assert.Nil(t, err) {
			testSchemaTablesAfterReset, err := db.Tables(ctx, pool, "test")
			if assert.Nil(t, err) {
				require.Len(t, testSchemaTablesAfterReset, 0)
			}
		}

		err = os.RemoveAll(migrationsDir)
		require.Nil(t, err)
	})

	t.Run("should test fs priority", func(t *testing.T) {
		opts := new(pgm.MigratorOptions)
		opts.ConnectionString = connStr
		opts.Command = pgm.CommandMigrate
		opts.Priority = pgm.PriorityFS
		opts.MigrationsTableSchema = "detmir_jobs"
		opts.MigrationsTable = "migrations"
		opts.MigrationsDir = migrationsDir

		_, err = genMigration(opts.MigrationsDir, "first_migration", "table1")
		require.Nil(t, err)

		migration2, err := genMigration(opts.MigrationsDir, "second_migration", "table2")
		require.Nil(t, err)

		migration3, err := genMigration(opts.MigrationsDir, "third_migration", "table3")
		require.Nil(t, err)

		appliedMigrations1, err := Migrate(ctx, opts)
		require.Nil(t, err)

		require.Len(t, appliedMigrations1, 3)

		require.Contains(t, appliedMigrations1[0].MigrationName, "first_migration")
		require.Equal(t, pgm.APPLIED, appliedMigrations1[0].Status)

		require.Contains(t, appliedMigrations1[1].MigrationName, "second_migration")
		require.Equal(t, pgm.APPLIED, appliedMigrations1[1].Status)

		require.Contains(t, appliedMigrations1[2].MigrationName, "third_migration")
		require.Equal(t, pgm.APPLIED, appliedMigrations1[2].Status)

		// Удалим 2 и 3 миграцию
		removeErr := errors.Join(
			os.Remove(migration2.Up),
			os.Remove(migration2.Down),
			os.Remove(migration3.Up),
			os.Remove(migration3.Down),
		)

		require.Nil(t, removeErr)

		// Создадим 4ю миграцю
		_, err = genMigration(opts.MigrationsDir, "fourth_migration", "table4")
		require.Nil(t, err)

		secondResults, err := Migrate(ctx, opts)
		require.Nil(t, err)
		require.Len(t, secondResults, 3)
		require.Contains(t, secondResults[0].MigrationName, "third")
		require.Equal(t, pgm.REVERTED, secondResults[0].Status)
		require.Contains(t, secondResults[1].MigrationName, "second")
		require.Equal(t, pgm.REVERTED, secondResults[1].Status)
		require.Contains(t, secondResults[2].MigrationName, "fourth")
		require.Equal(t, pgm.APPLIED, secondResults[2].Status)

		pool, err := db.Connect(ctx, connStr)
		require.Nil(t, err)
		defer pool.Close()

		err = db.ResetForTests(ctx, pool, opts.MigrationsTableNameWithSchema())
		if assert.Nil(t, err) {
			testSchemaTablesAfterReset, err := db.Tables(ctx, pool, "test")
			if assert.Nil(t, err) {
				require.Len(t, testSchemaTablesAfterReset, 0)
			}
		}

		err = os.RemoveAll(migrationsDir)
		require.Nil(t, err)
	})
}
