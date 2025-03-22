package cli

import (
	"github.com/quadgod/seo/pkg/pgm"
	"github.com/quadgod/seo/pkg/pgm/fs"
)

func Create(opts *pgm.MigratorOptions) (*pgm.Migration, error) {
	migrations, err := fs.CreateMigration(opts.MigrationsDir, opts.MigrationName)

	if err != nil {
		return nil, err
	}

	return migrations, nil
}
