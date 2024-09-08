package generators

import (
	"fmt"
	"math/rand/v2"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func ParsePattern(params ...any) (pattern string, err error) {
	if len(params) != 2 {
		return "", fmt.Errorf("invalid arguments, expected ['generator_name', 'pattern'] but got %v", params)
	}
	return params[1].(string), nil
}

func ParseUnion(params ...any) (pattern []string, err error) {
	if len(params) != 2 {
		return nil, fmt.Errorf("invalid arguments, expected ['generator_name', 'pattern'] but got %v", params)
	}
	return strings.Split(params[1].(string), "|"), nil
}

type Range[T any] interface {
	Bounds() (T, T)
	Min() T
	Max() T

	Exclusions() []T

	Rand() T
	RandPadded() string
}

var PatternRange = regexp.MustCompile(`(\d+\.\.\d+[\!\d\|]*|[\d\|]+)`)

type IntRange struct {
	Range[int64]

	min int64
	max int64

	exclude []int64

	minLen int
	maxLen int
}

func (r *IntRange) Exclusions() []int64 {
	return r.exclude
}

func (r *IntRange) Bounds() (int64, int64) {
	return r.min, r.max
}

func (r *IntRange) Min() int64 {
	min, _ := r.Bounds()
	return min
}

func (r *IntRange) Max() int64 {
	_, max := r.Bounds()
	return max
}

func (r *IntRange) String() string {
	excludes := ""
	if len(r.exclude) > 0 {
		excludes = "!"
		for _, x := range r.exclude {
			excludes = fmt.Sprintf("%s%d", excludes, x)
		}
	}
	return fmt.Sprintf("%d..%d%s", r.min, r.max, excludes)
}
func (r *IntRange) Rand() int64 {
	var val int64
	for {
		val = r.min + rand.Int64N(r.max-r.min)
		if !slices.Contains(r.exclude, val) {
			break
		}
	}
	return val
}
func (r *IntRange) RandPadded() string {
	val := fmt.Sprintf("%d", r.Rand())
	pad := ""
	if len(val) < r.minLen {
		pad = strings.Repeat("0", r.minLen-len(val))
	}
	return fmt.Sprintf("%s%s", pad, val)
}

type DiscreteValues struct {
	Range[int64]

	values []int64
	sizes  []int
}

func (r *DiscreteValues) Exclusions() []int64 {
	return []int64{}
}

func (r *DiscreteValues) Bounds() (int64, int64) {
	var min int64 = 99999
	var max int64 = -99999
	for _, v := range r.values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	return min, max
}

func (r *DiscreteValues) Min() int64 {
	min, _ := r.Bounds()
	return min
}

func (r *DiscreteValues) Max() int64 {
	_, max := r.Bounds()
	return max
}

func (r *DiscreteValues) String() string {
	items := []string{}
	for _, val := range r.values {
		items = append(items, fmt.Sprintf("%d", val))
	}
	return strings.Join(items, "|")
}

func (r *DiscreteValues) Rand() int64 {
	id := rand.IntN(len(r.values))
	return r.values[id]
}

func (r *DiscreteValues) RandPadded() string {
	id := rand.IntN(len(r.values))
	val := fmt.Sprintf("%d", r.values[id])
	pad := ""
	if len(val) < r.sizes[id] {
		pad = strings.Repeat("0", r.sizes[id]-len(val))
	}
	return fmt.Sprintf("%s%s", pad, val)
}

func ParseRangeArgs(params ...any) (r Range[int64], err error) {
	if len(params) != 2 {
		return nil, fmt.Errorf("invalid arguments, expected ['generator_name', 'min..max'] but got %v", params)
	}
	var expr string
	switch t := params[1].(type) {
	case string:
		expr = params[1].(string)
	case *string:
		expr = *params[1].(*string)
	default:
		return nil, fmt.Errorf("invalid argument 1, expected string but got %T", t)
	}
	return ParseRange(expr)
}

func ParseRange(s string) (Range[int64], error) {
	discreteValues := func(s string) (*DiscreteValues, error) {
		parts := strings.Split(s, "|")
		ret := &DiscreteValues{
			values: []int64{},
			sizes:  make([]int, len(parts)),
		}
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
			ret.sizes[i] = len(strings.TrimSpace(parts[i]))
			val, err := strconv.ParseInt(parts[i], 10, 64)
			if err != nil {
				return nil, err
			}
			ret.values = append(ret.values, val)
		}
		return ret, nil
	}
	if strings.Contains(s, "..") {
		parts := strings.Split(s, "..")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid range, expected 'min..max' but got '%s'", s)
		}
		sizes := make([]int, len(parts))
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
			sizes[i] = len(parts[i])
		}
		exclude := []int64{}
		if strings.Contains(parts[1], "!") {
			excludeParts := strings.Split(parts[1], "!")
			parts[1] = excludeParts[0]
			discrete, err := discreteValues(excludeParts[1])
			if err != nil {
				return nil, err
			}
			exclude = append(exclude, discrete.values...)
		}
		min, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, err
		}
		max, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, err
		}
		ret := &IntRange{
			min: min,
			max: max,

			exclude: exclude,

			minLen: sizes[0],
			maxLen: sizes[1],
		}
		return ret, nil
	}
	return discreteValues(s)
}

func ParseStrings(minRequired int, params ...any) (args []string, err error) {
	if len(params) != minRequired {
		return nil, fmt.Errorf("invalid arguments: %v", params)
	}
	ret := []string{}
	for i := range len(params) {
		var expr string
		switch t := params[i].(type) {
		case string:
			expr = params[i].(string)
		case *string:
			expr = *params[i].(*string)
		default:
			return nil, fmt.Errorf("invalid argument #%d, expected string but got %T", i, t)
		}
		ret = append(ret, expr)
	}
	return ret, nil
}
