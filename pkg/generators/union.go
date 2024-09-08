package generators

import (
	"database/sql"
	"math/rand/v2"

	"github.com/welschmorgan/datagen/pkg/generator"
)

const UNION_GENERATOR_NAME = "union"

type UnionGenerator struct {
	*CacheGenerator

	union []string
}

func NewUnionGenerator(db *sql.DB, options *generator.GeneratorOptions, union []string, variantGetter func(name string) generator.Generator) *UnionGenerator {
	return &UnionGenerator{
		CacheGenerator: NewCacheGenerator(options, UNION_GENERATOR_NAME, func() (string, error) {
			variantId := rand.IntN(len(union))
			variant := union[variantId]
			return variantGetter(variant).Next()
		}),
		union: union,
	}
}
