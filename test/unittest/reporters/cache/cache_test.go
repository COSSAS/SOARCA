package cache_test

import (
	"soarca/internal/reporters"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestAddElement(t *testing.T) {
	cache := reporters.ReporterCache{Size: 3}
	cache.Add(reporters.CacheEntry{Name: "test", Data: "asd"})
	assert.Equal(t, cache.Get()[0].Data, "asd")
}

func TestAddOverflowElement(t *testing.T) {
	cache := reporters.ReporterCache{Size: 1}
	cache.Add(reporters.CacheEntry{Name: "test", Data: "asd"})
	cache.Add(reporters.CacheEntry{Name: "test", Data: "lol"})
	assert.Equal(t, cache.Get()[0].Data, "lol")
}
