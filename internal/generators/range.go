package generators

import (
	"fmt"
	"math/rand/v2"
)

const INT_RANGE_GENERATOR_NAME = "int_range"

type IntRangeGenerator struct {
	Generator

	min  int64
	max  int64
	name string
}

func NewIntRangeGenerator(min, max int64) *IntRangeGenerator {
	return &IntRangeGenerator{
		min:  min,
		max:  max,
		name: INT_RANGE_GENERATOR_NAME,
	}
}

func (r *IntRangeGenerator) GetName() string {
	return r.name
}

func (r *IntRangeGenerator) SetName(v string) {
	r.name = v
}

func (r *IntRangeGenerator) Next() string {
	return fmt.Sprintf("%d", r.min+rand.Int64N(r.max-r.min))
}
