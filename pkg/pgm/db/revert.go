package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/quadgod/seo/pkg/pgm"
)

// RevertOne откатывает миграцию
func RevertOne(
	ctx context.Context,
	tx pgx.Tx,
	migTbl string,
	migName string,
) (*pgm.MigrationResult, error) {
	downSqlQuery := fmt.Sprintf(`SELECT down_sql FROM %s WHERE migration_name = $1 LIMIT 1;`, migTbl)
	row := tx.QueryRow(ctx, downSqlQuery, migName)

	var downSql string
	if err := row.Scan(&downSql); err != nil {
		return nil, fmt.Errorf("revert %s migration errors - down sql was not found. %v", migName, err)
	}

	if _, err := tx.Exec(ctx, downSql); err != nil {
		return nil, fmt.Errorf("revert %s migration errors - can't execute down sql. %v", migName, err)
	}

	delMigSql := fmt.Sprintf(`DELETE FROM %s WHERE migration_name = $1;`, migTbl)
	if _, err := tx.Exec(ctx, delMigSql, migName); err != nil {
		return nil, fmt.Errorf("revert %s migration errors - can't delete record from migrations table. %v", migName, err)
	}

	result := new(pgm.MigrationResult)
	result.MigrationName = migName
	result.Status = pgm.REVERTED

	return result, nil
}

// RevertMany откатывает набор миграций в обратном порядке
func RevertMany(
	ctx context.Context,
	tx pgx.Tx,
	migTbl string,
	migNames []string,
) ([]pgm.MigrationResult, error) {
	results := make([]pgm.MigrationResult, 0)

	for i := len(migNames) - 1; i >= 0; i-- {
		result, err := RevertOne(ctx, tx, migTbl, migNames[i])
		if err != nil {
			return nil, err
		}
		results = append(results, *result)
	}

	return results, nil
}
