package generators

import (
	"database/sql"
	"fmt"
	"math/rand/v2"

	"github.com/welschmorgan/datagen/pkg/generator"
)

const RANDOM_DB_ROW_GENERATOR_NAME = "random_row"

type RandomDBRowGenerator struct {
	*CacheGenerator

	options *generator.GeneratorOptions
	db      *sql.DB

	name string

	tableName        string
	tableFilterKey   string
	tableFilterValue string

	values []string
	seen   []string
}

func NewRandomDBRowGenerator(options *generator.GeneratorOptions, db *sql.DB, tableName, tableFilterKey, tableFilterValue string) (*RandomDBRowGenerator, error) {
	ret := &RandomDBRowGenerator{
		options: options,
		db:      db,

		name: RANDOM_DB_ROW_GENERATOR_NAME,

		tableName:        tableName,
		tableFilterKey:   tableFilterKey,
		tableFilterValue: tableFilterValue,
	}
	ret.CacheGenerator = NewCacheGenerator(options, RANDOM_DB_ROW_GENERATOR_NAME, ret.next)
	return ret, nil
}

func (g *RandomDBRowGenerator) next() (string, error) {
	if g.values == nil {
		rawQuery := fmt.Sprintf("SELECT value FROM %s WHERE %s = ?", g.tableName, g.tableFilterKey)
		query, err := g.db.Prepare(rawQuery)
		if err != nil {
			return "", err
		}

		rows, err := query.Query(g.tableFilterValue)
		if err != nil {
			return "", err
		}
		for rows.Next() {
			var value string
			if err := rows.Scan(&value); err != nil {
				return "", fmt.Errorf("failed to scan rows: %s", err)
			}
			g.values = append(g.values, value)
		}
		if len(g.values) == 0 {
			return "", fmt.Errorf("invalid random_row generator, filter matches nothing: '%s' (params=['%s'])", rawQuery, g.tableFilterValue)
		}
	}
	value_id := rand.Int() % len(g.values)
	return g.values[value_id], nil
}
