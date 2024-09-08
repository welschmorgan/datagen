package generators

import (
	"fmt"
	"strings"

	"github.com/welschmorgan/datagen/pkg/generator"
)

const PATTERN_GENERATOR_NAME = "pattern"

type PatternGenerator struct {
	*CacheGenerator

	pattern string
}

func NewPatternGenerator(options *generator.GeneratorOptions, pattern string) *PatternGenerator {
	matchId := 0
	ranges := map[string]Range[int64]{}
	evaledPattern := PatternRange.ReplaceAllStringFunc(pattern, func(match string) string {
		matchId += 1
		key := fmt.Sprintf("{group_%d}", matchId)
		range_, err := ParseRange(match)
		if err != nil {
			panic(err)
		}
		ranges[key] = range_
		return key
	})
	return &PatternGenerator{
		CacheGenerator: NewCacheGenerator(options, PATTERN_GENERATOR_NAME, func() (string, error) {
			pattern := evaledPattern
			for k, v := range ranges {
				pattern = strings.ReplaceAll(pattern, k, v.RandPadded())
			}
			return pattern, nil
		}),
		pattern: pattern,
	}
}
