package generator

type GeneratorOptions struct {
	OnlyUniqueValues     bool
	MaximumUniqueRetries int
}

func NewGeneratorOptions() *GeneratorOptions {
	return &GeneratorOptions{
		OnlyUniqueValues:     false,
		MaximumUniqueRetries: 20,
	}
}

type Generator interface {
	GetName() string
	SetName(string)

	GetOptions() *GeneratorOptions

	Next() (string, error)
}
