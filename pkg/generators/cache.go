package generators

import (
	"fmt"

	"github.com/welschmorgan/datagen/pkg/generator"
)

type CacheGenFunc func() (string, error)

type CacheGenerator struct {
	generator.Generator

	name     string
	options  *generator.GeneratorOptions
	gen_func CacheGenFunc
	seen     []string
}

func NewCacheGenerator(options *generator.GeneratorOptions, name string, gen_func CacheGenFunc) *CacheGenerator {
	return &CacheGenerator{
		name:     name,
		options:  options,
		gen_func: gen_func,
	}
}

func (g *CacheGenerator) GetName() string {
	return g.name
}

func (g *CacheGenerator) SetName(v string) {
	g.name = v
}

func (g *CacheGenerator) GetOptions() *generator.GeneratorOptions {
	return g.options
}

func (g *CacheGenerator) Next() (string, error) {
	next, err := g.gen_func()
	if err != nil {
		return "", err
	}
	if g.options.OnlyUniqueValues {
		numRetries := 1
		for g.HasSeenValue(next) {
			if numRetries >= g.options.MaximumUniqueRetries {
				return "", fmt.Errorf("not enough items, maximum unique retries reached (%d)", g.options.MaximumUniqueRetries)
			}
			numRetries += 1
			if next, err = g.gen_func(); err != nil {
				return "", err
			}
		}
		g.seen = append(g.seen, next)
	}
	return next, nil
}

func (g *CacheGenerator) HasSeenValue(v string) bool {
	for _, seen := range g.seen {
		if seen == v {
			return true
		}
	}
	return false
}
