package models

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/welschmorgan/datagen/internal/generators"
)

type Resource struct {
	Id            int64
	Name          string
	Template      *string
	GeneratorName *string
	Generator     generators.Generator
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
		log.Fatalf("failed to load resources: %s\n", err)
	}
	defer resources.Close()
	for resources.Next() {
		var id int64 = 0
		name := ""
		var template *string
		var generator *string
		if err := resources.Scan(&id, &name, &template, &generator); err != nil {
			log.Fatalf("failed to scan resource: %s\n", err)
		}
		ret = append(ret, NewResource(id, name, template, generator))
	}
	return ret
}
