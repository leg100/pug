package table

// BulkInsertMsg performs a bulk insertion of entities into a table
type BulkInsertMsg[T any] []T

// EnableFilterMsg enables the filter widget
type EnableFilterMsg struct{}

// ExitFilterMsg disables the filter widget
type ExitFilterMsg struct{}
