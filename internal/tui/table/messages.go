package table

// BulkInsertMsg performs a bulk insertion of entities into a table
type BulkInsertMsg[T any] []T

// DeselectMsg deselects any table rows currently selected.
type DeselectMsg struct{}
