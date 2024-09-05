package app

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"

	"github.com/welschmorgan/datagen/pkg/generators"
	"github.com/welschmorgan/datagen/pkg/models"
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
	db        *sql.DB
	reg       *generators.Registry
	options   *Options
	resources []*models.Resource
}

func New(opts *Options) *App {
	return &App{
		db:      nil,
		reg:     nil,
		options: opts,
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
				r.Generator = g
				a.resources = append(a.resources, r)
				msg += fmt.Sprintf(" - %s", g.GetName())
			}
		}
		log.Println(msg)
	}

	return nil
}

func (a *App) GetResource(name string) (*models.Resource, error) {
	for _, app_res := range a.resources {
		if strings.EqualFold(app_res.Name, name) {
			return app_res, nil
		}
	}
	return nil, fmt.Errorf("failed to find resource '%s'", name)
}

func (a *App) Run() error {
	type Result struct {
		g     generators.Generator
		round int
		value string
	}

	value_chan := make(chan Result)
	var wgGen sync.WaitGroup
	var wgOut sync.WaitGroup

	nTasks := a.options.count * len(a.options.resources)
	wgGen.Add(nTasks)
	wgOut.Add(nTasks)

	for _, user_res := range a.options.resources {
		app_res, err := a.GetResource(user_res)
		if err != nil {
			return err
		}
		gen := app_res.Generator

		go func() {
			for i := range a.options.count {
				value_chan <- Result{g: gen, round: i, value: gen.Next()}
				wgGen.Done()
			}
		}()
		go func() {
			for range a.options.count {
				res := <-value_chan
				output := a.options.output.fmt(res.g, res.round, res.value)
				fmt.Println(output)
				wgOut.Done()
			}
		}()
	}

	wgGen.Wait()
	wgOut.Wait()

	return nil
}

func (a *App) Shutdown() error {
	if err := a.db.Close(); err != nil {
		log.Printf("Failed to close DB, %s", err)
	}
	return nil
}
