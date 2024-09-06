package app

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/MatusOllah/slogcolor"
	"github.com/fatih/color"
	_ "github.com/mattn/go-sqlite3"

	"github.com/welschmorgan/datagen/pkg/generators"
	"github.com/welschmorgan/datagen/pkg/models"
	"github.com/welschmorgan/datagen/pkg/seed"
)

const DB_FILE string = "resources.db"

func GeneratorForResource(options *generators.GeneratorOptions, res *models.Resource, reg *generators.Registry) (generators.Generator, error) {
	if res.GeneratorName != nil {
		parts := strings.Split(*res.GeneratorName, ":")
		typeName := parts[0]
		gen_alloc, err := reg.GetType(typeName)
		if err != nil {
			return nil, err
		}
		params := []interface{}{}
		if res.Template != nil {
			params = append(params, res.Template)
		}
		if len(parts) > 1 {
			for _, part := range parts[1:] {
				params = append(params, part)
			}
		}
		return gen_alloc(options, params...)
	}
	return nil, nil
}

func allocateGeneratorIntRange(options *generators.GeneratorOptions, params ...any) (generators.Generator, error) {
	min, max, err := generators.ParseRange(params...)
	if err != nil {
		return nil, err
	}
	return generators.NewIntRangeGenerator(options, min, max), nil
}

func allocateGeneratorRandomDB(db *sql.DB) generators.GeneratorAllocator {
	return func(options *generators.GeneratorOptions, params ...any) (generators.Generator, error) {
		expectedArgs := 2
		expectedArgNames := "table, filter"
		args, err := generators.ParseStrings(expectedArgs, params...)
		if len(args) != 2 {
			return nil, fmt.Errorf("invalid arguments to RandomDBRowGenerator, expected %d args (%s) but got %d", expectedArgs, expectedArgNames, len(args))
		}
		if err != nil {
			return nil, err
		}
		tableName := args[0]
		parts := strings.Split(args[1], "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid arguments to RandomDBRowGenerator, tableFilter is invalid. Expected 'column=value' but got '%s'", args[1])
		}
		tableFilterKey := parts[0]
		tableFilterValue := parts[1]
		return generators.NewRandomDBRowGenerator(options, db, tableName, tableFilterKey, tableFilterValue)
	}
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
	if err = a.initLogging(); err != nil {
		return err
	}
	slog.Debug(fmt.Sprintf("%+v", a.options))
	_, existErr := os.Stat(DB_FILE)
	a.db, err = sql.Open("sqlite3", DB_FILE)
	if err != nil {
		slog.Error("failed to open DB", "err", err)
		panic("Fatal error")
	}
	if errors.Is(existErr, fs.ErrNotExist) {
		slog.Warn("DB does not exist, creating now ...")
		if err = a.Seed(); err != nil {
			return fmt.Errorf("failed to create DB, %s", err)
		}
	} else if a.options.seed {
		if err = a.Seed(); err != nil {
			return fmt.Errorf("failed to seed DB, %s", err)
		}
	}

	a.reg = generators.NewRegistry()
	a.reg.AddType(generators.INT_RANGE_GENERATOR_NAME, allocateGeneratorIntRange)
	a.reg.AddType(generators.RANDOM_DB_ROW_GENERATOR_NAME, allocateGeneratorRandomDB(a.db))

	resources := models.LoadResources(a.db)
	for _, r := range resources {
		msg := fmt.Sprint(r)
		if r.GeneratorName != nil {
			if g, err := GeneratorForResource(&a.options.generator, r, a.reg); err != nil {
				slog.Error(fmt.Sprintf("%s - %s", msg, err))
			} else {
				r.Generator = g
				a.resources = append(a.resources, r)
				slog.Info(fmt.Sprintf("%s - %s", msg, g.GetName()))
			}
		}
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

func (a *App) Generate() error {
	type Result struct {
		resource  *models.Resource
		generator generators.Generator
		round     int
		value     string
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
				if value, err := gen.Next(); err != nil {
					panic(fmt.Errorf("failed to generate value #%d: %s", i, err))
				} else {
					value_chan <- Result{resource: app_res, generator: gen, round: i, value: value}
					wgGen.Done()
				}
			}
		}()
		go func() {
			for range a.options.count {
				res := <-value_chan
				output := a.options.output.fmt(res.resource, res.generator, res.round, res.value)
				fmt.Println(output)
				wgOut.Done()
			}
		}()
	}

	wgGen.Wait()
	wgOut.Wait()

	return nil
}

func (a *App) Seed() error {
	return seed.NewDefaultSeeder(a.db).Seed()
}

func (a *App) Shutdown() error {
	if err := a.db.Close(); err != nil {
		slog.Warn("Failed to close DB", "err", err)
	}
	return nil
}

func (a *App) initLogging() error {
	level := slog.LevelInfo
	if a.options.verbose {
		level = slog.LevelDebug
	}
	slog.SetLogLoggerLevel(level)
	slog.SetDefault(slog.New(slogcolor.NewHandler(os.Stderr, &slogcolor.Options{
		Level:         level,
		TimeFormat:    time.RFC3339,
		SrcFileMode:   slogcolor.ShortFile,
		SrcFileLength: 0,
		MsgPrefix:     color.HiWhiteString("| "),
		MsgLength:     0,
		MsgColor:      color.New(),
	})))

	return nil
}
