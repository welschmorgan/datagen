package app

import (
	"database/sql"
	"fmt"
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
	min, max, err := generators.ParseRange(params...)
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

func (a *App) Init() error {
	if err := a.initLogging(); err != nil {
		return err
	}
	slog.Debug("Configuration", "app-options", a.options)
	var err error
	a.db, err = sql.Open("sqlite3", DB_FILE)
	if err != nil {
		slog.Error("failed to open resources DB", "err", err)
		panic("Fatal error")
	}

	a.reg = generators.NewRegistry()
	a.reg.AddType(generators.INT_RANGE_GENERATOR_NAME, allocateGeneratorIntRange)

	resources := models.LoadResources(a.db)
	for _, r := range resources {
		msg := fmt.Sprint(r)
		if r.GeneratorName != nil {
			if g, err := GeneratorForResource(r, a.reg); err != nil {
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

func (a *App) Run() error {
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
				value_chan <- Result{resource: app_res, generator: gen, round: i, value: gen.Next()}
				wgGen.Done()
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

func (a *App) Shutdown() error {
	if err := a.db.Close(); err != nil {
		slog.Warn("Failed to close DB", "err", err)
	}
	return nil
}
