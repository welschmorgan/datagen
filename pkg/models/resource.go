package models

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/welschmorgan/datagen/pkg/generator"
)

type Resource struct {
	Id            int64
	Name          string
	Template      *string
	GeneratorName *string
	Generator     generator.Generator
}

func NewResource(id int64, name string, template *string, generator *string) *Resource {
	return &Resource{
		Id:            id,
		Name:          name,
		Template:      template,
		GeneratorName: generator,
		Generator:     nil,
	}
}

func (r *Resource) String() string {
	return fmt.Sprintf("Resource #%d: %s = %s", r.Id, r.Name, r.FullGeneratorName())
}

func (r *Resource) FullGeneratorName() string {
	generator := "<nil>"
	if r.GeneratorName != nil {
		generator = *r.GeneratorName
	}
	template := "<nil>"
	if r.Template != nil {
		template = *r.Template
	}
	return fmt.Sprintf("%s['%s']", generator, template)
}

func LoadResources(db *sql.DB) []*Resource {
	ret := []*Resource{}
	resources, err := db.Query("select * from resource")
	if err != nil {
		slog.Error("failed to load resources", "err", err)
		panic("Fatal error")
	}
	defer resources.Close()
	for resources.Next() {
		var id int64 = 0
		name := ""
		var template *string
		var generator *string
		if err := resources.Scan(&id, &name, &generator, &template); err != nil {
			slog.Error("failed to scan resource", "err", err)
			panic("Fatal error")
		}
		ret = append(ret, NewResource(id, name, template, generator))
	}
	return ret
}
