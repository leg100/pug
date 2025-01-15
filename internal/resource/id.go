package resource

// ID uniquely identifies a Pug resource.
type ID any

// Identifiable is a Pug resource with an identity.
type Identifiable interface {
	GetID() ID
}
