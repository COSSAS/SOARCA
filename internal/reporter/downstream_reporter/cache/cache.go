package cache

import (
	"soarca/internal/guid"
	"soarca/models/cacao"
)

// The cache could be a downstream reporter which is always initiated
// The reporter aggregates all info of execution, and gives it to downstream reporters
// The task of a downstream reporter e.g. DB is taking exec info and putting into the DB
// Do not make the cache a property, but a reporter specific for reporting into the cache
// The cache reporter would offer functionalities to query cache results
// The cache can interface the database reporter
// We can expose the cache reporter via Reporting APIs. The cache reporter would handle retrieval
// The cache could then include a database-specific downstreamreporter that handles functions

type ReporterCache struct {
	Size  int
	cache []ReportEntry
}

// Give reportEntry to all sub-reporters

// This is now cache-specific, and the cache will be implemented as
// ad-hoc downstreamreporter

type ReportEntry struct {
	ExecutionId guid.IGuid
	PlaybookId  string
	// Key is StepID
	WorkflowResults map[string]StepResults
	PlaybookResult  error
	Name            string
	Data            interface{}
}

type StepResults struct {
	Variables cacao.Variables
	Error     error
}

func (reportsCache *ReporterCache) Add(result ReportEntry) {
	if len(reportsCache.cache) < reportsCache.Size {
		reportsCache.cache = append(reportsCache.cache, result)
	} else {
		// FIFO logic
		reportsCache.cache = reportsCache.cache[1:]
		reportsCache.cache = append(reportsCache.cache, result)
	}
}

func (reportsCache *ReporterCache) Get() []ReportEntry {
	return reportsCache.cache
}
