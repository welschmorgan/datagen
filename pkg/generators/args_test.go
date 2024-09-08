package generators_test

import (
	"slices"
	"testing"

	"github.com/welschmorgan/datagen/pkg/generators"
)

func TestParseIntRange(t *testing.T) {
	expr := "0..10"
	rng, err := generators.ParseRange(expr)
	if err != nil {
		t.Errorf("failed to parse IntRange from '%s', %s", expr, err)
	}
	min, max := rng.Bounds()
	if min != 0 && max != 10 {
		t.Errorf("invalid bounds, expected '%d..%d' but got '%d..%d'", 0, 10, min, max)
	}
}

func TestParseIntRangeWithExclusions(t *testing.T) {
	expr := "0..10!2|3"
	rng, err := generators.ParseRange(expr)
	if err != nil {
		t.Errorf("failed to parse IntRange from '%s', %s", expr, err)
	}
	min, max := rng.Bounds()
	if min != 0 && max != 10 {
		t.Errorf("invalid bounds, expected '%d..%d' but got '%d..%d'", 0, 10, min, max)
	}
	excl := rng.Exclusions()
	expected := []int64{2, 3}
	if slices.Compare(excl, expected) != 0 {
		t.Errorf("invalid exclusions for IntRange, expected %v but got %v", expected, excl)
	}
}

func TestParseDiscreteValues(t *testing.T) {
	expr := "1|2|3"
	rng, err := generators.ParseRange(expr)
	if err != nil {
		t.Errorf("failed to parse DiscreteValues from '%s', %s", expr, err)
	}
	min, max := rng.Bounds()
	if min != 1 && max != 3 {
		t.Errorf("invalid bounds, expected '%d..%d' but got '%d..%d'", 1, 3, min, max)
	}
}
