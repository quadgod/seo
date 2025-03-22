package fs

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func Test_CreateMigration(t *testing.T) {
	migrationsDir := path.Join(t.TempDir(), "/filesystem_create_test")

	t.Cleanup(func() {
		err := os.RemoveAll(migrationsDir)
		if err != nil {
			t.Fatalf("unable to cleanup migrations dir %s", err)
		}
	})

	t.Run("should create migrations files", func(t *testing.T) {
		migrationDirEntries1, err := os.ReadDir(migrationsDir)
		assert.Nil(t, migrationDirEntries1)
		assert.True(t, os.IsNotExist(err))

		firstMigration, err := CreateMigration(migrationsDir, "first_migration")
		if !assert.Nil(t, err) {
			t.FailNow()
		}

		assert.Contains(t, firstMigration.Up, ".up.sql")
		assert.Contains(t, firstMigration.Down, ".down.sql")

		secondMigration, err := CreateMigration(migrationsDir, "second_migration")
		if !assert.Nil(t, err) {
			t.FailNow()
		}

		assert.Contains(t, secondMigration.Up, ".up.sql")
		assert.Contains(t, secondMigration.Down, ".down.sql")

		migrationDirEntries2, err := os.ReadDir(migrationsDir)
		if !assert.Nil(t, err) {
			t.FailNow()
		}
		assert.Len(t, migrationDirEntries2, 4)
	})
}
