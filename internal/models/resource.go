package models

import (
	"database/sql"
	"fmt"
	"log"
)

type Resource struct {
	id        int64
	name      string
	template  *string
	generator *string
}

func NewResource(id int64, name string, template *string, generator *string) *Resource {
	return &Resource{
		id:        id,
		name:      name,
		template:  template,
		generator: generator,
	}
}

func (r *Resource) String() string {
	generator := "<nil>"
	if r.generator != nil {
		generator = *r.generator
	}
	template := "<nil>"
	if r.template != nil {
		template = *r.template
	}
	return fmt.Sprintf("Resource #%d: %s = %v[\"%v\"]", r.id, r.name, generator, template)
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
