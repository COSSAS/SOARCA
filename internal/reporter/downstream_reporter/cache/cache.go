package cache

import (
	"soarca/internal/guid"
	"soarca/internal/reporter/downstream_reporter"
)

// The cache reporter would offer functionalities to query cache results
// The cache can interface the database reporter
// We can expose the cache reporter via Reporting APIs. The cache reporter would handle retrieval
// The cache could then include a database-specific downstreamreporter that handles functions

type CacheReporter struct {
	Size  int
	cache []CacheEntry
}

type CacheEntry struct {
	ExecutionId guid.IGuid
	PlaybookId  string
	// Key is StepID
	WorkflowResults map[string]downstream_reporter.WorkflowEntry
	StepResults     map[string]downstream_reporter.StepEntry
	PlaybookResult  error
	Name            string
	Data            interface{}
}

func (cacheReporter *CacheReporter) Add(result CacheEntry) {
	if len(cacheReporter.cache) < cacheReporter.Size {
		cacheReporter.cache = append(cacheReporter.cache, result)
	} else {
		// FIFO logic
		cacheReporter.cache = cacheReporter.cache[1:]
		cacheReporter.cache = append(cacheReporter.cache, result)
	}
}

func (cacheReporter *CacheReporter) Get() []CacheEntry {
	return cacheReporter.cache
}
