package generators

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/welschmorgan/datagen/pkg/generator"
	"github.com/welschmorgan/datagen/pkg/models"
)

type GeneratorAllocator func(*generator.GeneratorOptions, ...any) (generator.Generator, error)

func GeneratorForResource(options *generator.GeneratorOptions, res *models.Resource, reg *Registry) (generator.Generator, error) {
	parts := []interface{}{}
	if res.GeneratorName != nil {
		parts = append(parts, *res.GeneratorName)
	}
	if res.Template != nil {
		for _, arg := range strings.Split(*res.Template, ":") {
			parts = append(parts, arg)
		}
	}
	if res.GeneratorName != nil {
		typeName := parts[0]
		gen_alloc, err := reg.GetType(typeName.(string))
		if err != nil {
			return nil, err
		}
		return gen_alloc(options, parts...)
	}
	return nil, nil
}

func AllocateGeneratorPattern(options *generator.GeneratorOptions, params ...any) (generator.Generator, error) {
	pattern, err := ParsePattern(params...)
	if err != nil {
		return nil, err
	}
	return NewPatternGenerator(options, pattern), nil
}

func AllocateGeneratorUnion(db *sql.DB, resGetter func(name string) generator.Generator) GeneratorAllocator {
	return func(options *generator.GeneratorOptions, params ...any) (generator.Generator, error) {
		variants, err := ParseUnion(params...)
		if err != nil {
			return nil, err
		}
		return NewUnionGenerator(db, options, variants, resGetter), nil
	}
}

func AllocateGeneratorIntRange(options *generator.GeneratorOptions, params ...any) (generator.Generator, error) {
	r, err := ParseRangeArgs(params...)
	if err != nil {
		return nil, err
	}
	return NewIntRangeGenerator(options, r), nil
}

func AllocateGeneratorRandomDB(db *sql.DB) GeneratorAllocator {
	return func(options *generator.GeneratorOptions, params ...any) (generator.Generator, error) {
		expectedArgs := 3
		expectedArgNames := "'random_row', table, filter"
		args, err := ParseStrings(expectedArgs, params...)
		if len(args) != expectedArgs {
			return nil, fmt.Errorf("invalid arguments to RandomDBRowGenerator, expected %d args (%s) but got %d\n%s", expectedArgs, expectedArgNames, len(args), err)
		}
		if err != nil {
			return nil, err
		}
		tableName := args[1]
		parts := strings.Split(args[2], "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid arguments to RandomDBRowGenerator, tableFilter is invalid. Expected 'column=value' but got '%s'", args[1])
		}
		tableFilterKey := parts[0]
		tableFilterValue := parts[1]
		return NewRandomDBRowGenerator(options, db, tableName, tableFilterKey, tableFilterValue)
	}
}
