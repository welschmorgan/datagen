package generators

type Generator interface {
	GetName() string
	SetName(string)

	Next() string
}
