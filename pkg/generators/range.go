package generators

import (
	"fmt"
	"math/rand/v2"
)

const INT_RANGE_GENERATOR_NAME = "int_range"

type IntRangeGenerator struct {
	*CacheGenerator

	min int64
	max int64
}

func NewIntRangeGenerator(options *GeneratorOptions, min, max int64) *IntRangeGenerator {
	return &IntRangeGenerator{
		CacheGenerator: NewCacheGenerator(options, INT_RANGE_GENERATOR_NAME, func() (string, error) {
			return fmt.Sprintf("%d", min+rand.Int64N(max-min)), nil
		}),
		min: min,
		max: max,
	}
}
