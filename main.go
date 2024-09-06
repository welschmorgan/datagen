package main

import (
	"github.com/welschmorgan/datagen/pkg/app"
	"github.com/welschmorgan/datagen/pkg/seed"

	_ "embed"
)

//go:embed assets/seed.sql
var DBSeedScript string

func main() {
	seed.DEFAULT_SEED_SCHEMA = &DBSeedScript
	a := app.New(app.ParseOptions())
	if err := a.Init(); err != nil {
		panic(err)
	}
	if err := a.Generate(); err != nil {
		panic(err)
	}
	defer a.Shutdown()
}
