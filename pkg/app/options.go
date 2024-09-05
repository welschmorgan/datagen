package app

import (
	"flag"
	"fmt"
	"strings"

	"github.com/welschmorgan/datagen/pkg/generators"
	"github.com/welschmorgan/datagen/pkg/models"
)

const DEFAULT_ITEMS_COUNT = 100

type OutputFormatter interface {
	fmt(r *models.Resource, g generators.Generator, round int, value string) string
}

type DefaultOutputFormatter struct {
	OutputFormatter
}

func NewDefaultOutputFormatter() *DefaultOutputFormatter {
	return &DefaultOutputFormatter{}
}

func (f *DefaultOutputFormatter) fmt(r *models.Resource, g generators.Generator, round int, value string) string {
	return fmt.Sprintf("[%s:%s #%d] %s", r.Name, g.GetName(), round, value)
}

type Options struct {
	verbose   bool
	resources ResourceList
	count     int
	output    OutputFormatter
	generator generators.GeneratorOptions
}

type ResourceList []string

func (l *ResourceList) String() string {
	return strings.Join(*l, ",")
}

func (l *ResourceList) Set(value string) error {
	parts := strings.Split(value, ",")
	for _, part := range parts {
		*l = append(*l, part)
	}
	return nil
}

func ParseOptions() *Options {
	opt := Options{
		verbose:   false,
		resources: []string{},
		output:    NewDefaultOutputFormatter(),
		count:     0,
		generator: *generators.NewGeneratorOptions(),
	}
	flag.BoolVar(&opt.verbose, "verbose", opt.verbose, "show additional log messages")
	flag.Var(&opt.resources, "resource", "generate a dataset with the specified type")
	flag.IntVar(&opt.count, "count", DEFAULT_ITEMS_COUNT, "generate this number of items")
	flag.BoolVar(&opt.generator.OnlyUniqueValues, "unique", opt.generator.OnlyUniqueValues, "only generate unique values")
	flag.Parse()
	return &opt
}
