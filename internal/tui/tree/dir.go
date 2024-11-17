package tree

type dir struct {
	name string
}

func (d dir) String() string {
	return d.name
}
