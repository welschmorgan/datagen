package main

import (
	"github.com/welschmorgan/datagen/pkg/app"
)

func main() {
	a := app.New(app.ParseOptions())
	if err := a.Init(); err != nil {
		panic(err)
	}
	if err := a.Run(); err != nil {
		panic(err)
	}
	defer a.Shutdown()
}
