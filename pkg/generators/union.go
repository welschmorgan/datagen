package generators

import (
	"database/sql"
	"math/rand/v2"
)

const UNION_GENERATOR_NAME = "union"

type UnionGenerator struct {
	*CacheGenerator

	union []string
}

func NewUnionGenerator(db *sql.DB, options *GeneratorOptions, union []string, variantGetter func(name string) Generator) *UnionGenerator {
	return &UnionGenerator{
		CacheGenerator: NewCacheGenerator(options, UNION_GENERATOR_NAME, func() (string, error) {
			variantId := rand.IntN(len(union))
			variant := union[variantId]
			return variantGetter(variant).Next()
		}),
		union: union,
	}
}
