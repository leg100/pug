package resource

// ID uniquely identifies a Pug resource.
type ID any

type Identifiable interface {
	GetID() ID
}
