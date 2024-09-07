package app

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/MatusOllah/slogcolor"
	"github.com/fatih/color"
	_ "github.com/mattn/go-sqlite3"

	"github.com/welschmorgan/datagen/pkg/cache"
	"github.com/welschmorgan/datagen/pkg/config"
	"github.com/welschmorgan/datagen/pkg/generators"
	"github.com/welschmorgan/datagen/pkg/models"
	"github.com/welschmorgan/datagen/pkg/seed"
)

func DBPath() string {
	return fmt.Sprintf("%s/%s", cache.RootCacheDir(), "resources.db")
}

func GeneratorForResource(options *generators.GeneratorOptions, res *models.Resource, reg *generators.Registry) (generators.Generator, error) {
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

func allocateGeneratorIntRange(options *generators.GeneratorOptions, params ...any) (generators.Generator, error) {
	min, max, err := generators.ParseRange(params...)
	if err != nil {
		return nil, err
	}
	return generators.NewIntRangeGenerator(options, min, max), nil
}

func allocateGeneratorRandomDB(db *sql.DB) generators.GeneratorAllocator {
	return func(options *generators.GeneratorOptions, params ...any) (generators.Generator, error) {
		expectedArgs := 3
		expectedArgNames := "'random_row', table, filter"
		args, err := generators.ParseStrings(expectedArgs, params...)
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
		return generators.NewRandomDBRowGenerator(options, db, tableName, tableFilterKey, tableFilterValue)
	}
}

type App struct {
	db        *sql.DB
	reg       *generators.Registry
	options   *Options
	config    *config.Config
	resources []*models.Resource
	locales   []*models.Locale
}

func New(opts *Options) *App {
	return &App{
		db:      nil,
		reg:     nil,
		options: opts,
		config:  config.Default(),
	}
}

func (a *App) Init() error {
	var err error
	if err = a.initLogging(); err != nil {
		return err
	}
	if a.options.resetConfig {
		if err = a.config.Reset(a.options.configPath); err != nil {
			return err
		}
	} else if err = a.config.Init(a.options.configPath); err != nil {
		return err
	}
	slog.Debug("Command-line options", "value", a.options)
	slog.Debug("User configuration", "path", a.options.configPath)
	slog.Debug("Data directory", "path", cache.RootCacheDir())
	dbPath := DBPath()
	slog.Debug("Database", "path", dbPath)

	dbDir := filepath.Dir(dbPath)
	if _, err := os.Stat(dbDir); errors.Is(err, os.ErrNotExist) {
		if err = os.MkdirAll(dbDir, 0755); err != nil {
			return err
		}
	}
	_, existErr := os.Stat(dbPath)
	a.db, err = sql.Open("sqlite3", dbPath)
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
		tpl := ""
		if r.Template != nil {
			tpl = *r.Template
		}
		if r.GeneratorName != nil {
			if g, err := GeneratorForResource(&a.options.generator, r, a.reg); err != nil {
				slog.Error(fmt.Sprintf("Invalid resource #%d '%s'", r.Id, r.Name), "err", err, "generator", *r.GeneratorName, "template", tpl)
			} else {
				r.Generator = g
				a.resources = append(a.resources, r)
				slog.Debug(fmt.Sprintf("Found resource #%d '%s'", r.Id, r.Name), "generator", *r.GeneratorName, "template", tpl)
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
	seeder, err := seed.NewSeederFromConfig(a.db, a.config)
	if err != nil {
		return err
	}
	return seeder.Seed()
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
