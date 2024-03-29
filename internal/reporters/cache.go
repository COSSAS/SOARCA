package reporters

type ReporterCache struct {
	size  int
	cache []interface{}
}

func (reportsCache *ReporterCache) addToCache(result interface{}) {
	if len(reportsCache.cache) < reportsCache.size {
		reportsCache.cache = append(reportsCache.cache, result)
	} else {
		// FIFO logic
		reportsCache.cache = reportsCache.cache[1:]
		reportsCache.cache = append(reportsCache.cache, result)
	}
}
