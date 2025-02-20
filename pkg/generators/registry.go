package generators

import (
	"fmt"

	"github.com/welschmorgan/datagen/pkg/generator"
)

type Registry struct {
	types     map[string]GeneratorAllocator
	instances map[string]generator.Generator
}

func NewRegistry() *Registry {
	return &Registry{
		types:     map[string]GeneratorAllocator{},
		instances: map[string]generator.Generator{},
	}
}

func (r *Registry) GetInstance(k string) (generator.Generator, error) {
	t, ok := r.instances[k]
	if !ok {
		return nil, fmt.Errorf("unknown generator type '%s'", k)
	}
	return t, nil
}

func (r *Registry) FindInstance(k string) generator.Generator {
	return r.instances[k]
}

func (r *Registry) ContainsInstance(k string) bool {
	_, ok := r.instances[k]
	return ok
}

func (r *Registry) AddInstance(g generator.Generator) error {
	if r.ContainsInstance(g.GetName()) {
		return fmt.Errorf("Generator instance '%s' already registered", g.GetName())
	}
	r.instances[g.GetName()] = g
	return nil
}

func (r *Registry) GetType(k string) (GeneratorAllocator, error) {
	t, ok := r.types[k]
	if !ok {
		return nil, fmt.Errorf("unknown generator type '%s'", k)
	}
	return t, nil
}

func (r *Registry) FindType(k string) GeneratorAllocator {
	return r.types[k]
}

func (r *Registry) ContainsType(k string) bool {
	_, ok := r.types[k]
	return ok
}

func (r *Registry) AddType(k string, g GeneratorAllocator) error {
	if r.ContainsType(k) {
		return fmt.Errorf("Generator type '%s' already registered", k)
	}
	r.types[k] = g
	return nil
}
