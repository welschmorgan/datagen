package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/welschmorgan/datagen/internal/generators"
	"github.com/welschmorgan/datagen/internal/models"
)

const DB_FILE string = "resources.db"

func GeneratorForResource(res *models.Resource, reg *generators.Registry) (generators.Generator, error) {
	if res.GeneratorName != nil {
		typeName := res.GeneratorName
		gen_alloc, err := reg.GetType(*typeName)
		if err != nil {
			return nil, err
		}
		return gen_alloc(res.Template)
	}
	return nil, nil
}

func allocateGeneratorIntRange(params ...any) (generators.Generator, error) {
	if len(params) != 1 {
		return nil, fmt.Errorf("invalid arguments: %v", params)
	}
	var expr string
	switch t := params[0].(type) {
	case string:
		expr = params[0].(string)
	case *string:
		expr = *params[0].(*string)
	default:
		return nil, fmt.Errorf("invalid argument 0, expected string but got %T", t)
	}
	parts := strings.Split(expr, "..")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid argument 0, expected 'min..max' but got '%s'", params[0])
	}
	min, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, err
	}
	max, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, err
	}
	return generators.NewIntRangeGenerator(min, max), nil
}

type App struct {
	db  *sql.DB
	reg *generators.Registry
}

func NewApp() *App {
	return &App{
		nil,
		nil,
	}
}

func (a *App) Init() error {
	var err error
	a.db, err = sql.Open("sqlite3", DB_FILE)
	if err != nil {
		log.Fatalf("failed to open resources DB: %s\n", err)
	}

	a.reg = generators.NewRegistry()
	a.reg.AddType(generators.INT_RANGE_GENERATOR_NAME, allocateGeneratorIntRange)

	resources := models.LoadResources(a.db)
	for _, r := range resources {
		msg := fmt.Sprint(r)
		if r.GeneratorName != nil {
			if g, err := GeneratorForResource(r, a.reg); err != nil {
				msg += fmt.Sprintf(" - %s", err)
			} else {
				msg += fmt.Sprintf(" - %s", g.GetName())
			}
		}
		fmt.Println(msg)
	}

	return nil
}

func (a *App) Shutdown() error {
	if err := a.db.Close(); err != nil {
		log.Printf("Failed to close DB, %s", err)
	}
	return nil
}

func main() {
	a := NewApp()
	if err := a.Init(); err != nil {
		panic(err)
	}
	defer a.Shutdown()
}
