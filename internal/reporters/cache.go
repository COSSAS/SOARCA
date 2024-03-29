package reporters

type ReporterCache struct {
	Size  int
	cache []CacheEntry
}

type CacheEntry struct {
	Name string
	Data interface{}
}

func (reportsCache *ReporterCache) Add(result CacheEntry) {
	if len(reportsCache.cache) < reportsCache.Size {
		reportsCache.cache = append(reportsCache.cache, result)
	} else {
		// FIFO logic
		reportsCache.cache = reportsCache.cache[1:]
		reportsCache.cache = append(reportsCache.cache, result)
	}
}

func (reportsCache *ReporterCache) Get() []CacheEntry {
	return reportsCache.cache
}
