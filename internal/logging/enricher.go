package logging

// enricher enriches a log record with further meaningful attributes that aren't
// readily available to the caller.
type enricher struct {
	enrichers []Enricher
}

// Enricher implementations update a log record with further info
type Enricher interface {
	EnrichLogRecord(args ...any) []any
}

func (e *enricher) AddEnricher(enricher Enricher) {
	e.enrichers = append(e.enrichers, enricher)
}

func (e *enricher) enrich(args ...any) []any {
	for _, en := range e.enrichers {
		args = en.EnrichLogRecord(args...)
	}
	return args
}
