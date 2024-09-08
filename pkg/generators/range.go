package generators

import "github.com/welschmorgan/datagen/pkg/generator"

const INT_RANGE_GENERATOR_NAME = "int_range"

type IntRangeGenerator struct {
	*CacheGenerator
}

func NewIntRangeGenerator(options *generator.GeneratorOptions, range_ Range[int64]) *IntRangeGenerator {
	return &IntRangeGenerator{
		CacheGenerator: NewCacheGenerator(options, INT_RANGE_GENERATOR_NAME, func() (string, error) {
			return range_.RandPadded(), nil
		}),
	}
}
