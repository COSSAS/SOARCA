// Package registry provides a small, generic, in-memory store for
// pending work items that are registered under a two-level key, looked
// up by callers that later resolve or expire them, and automatically
// pruned once their outer key has no remaining inner entries.
//
// It captures the shape used by SOARCA's manual-command interaction queue
// (keyed by execution id -> step execution id): register a pending item,
// let something external resolve it asynchronously (e.g. by pushing a
// result onto a channel stored alongside it), and clean up stale entries.
//
// The Fin job queue (pkg/core/capability/fin/queue) looks superficially
// similar - it is also a two-level, register/resolve/clean-up store keyed
// by capability type -> job id - but it is deliberately its own,
// independent implementation rather than an instantiation of Registry: it
// needs capability-type-filtered claiming by any number of competing
// pollers, per-job lease expiry with a background sweep-and-requeue loop,
// and broadcast wakeups for long-polling claimants, none of which fit this
// package's single-value-per-key, synchronous get/remove model.
package registry

import (
	"fmt"
	"sync"
)

// ErrOuterKeyNotFound indicates no entries are registered under outerKey
// at all.
type ErrOuterKeyNotFound struct {
	OuterKey string
}

func (e ErrOuterKeyNotFound) Error() string {
	return fmt.Sprintf("no entries found for key %q", e.OuterKey)
}

// ErrInnerKeyNotFound indicates outerKey has registered entries, but none
// under innerKey.
type ErrInnerKeyNotFound struct {
	OuterKey string
	InnerKey string
}

func (e ErrInnerKeyNotFound) Error() string {
	return fmt.Sprintf("no entry found for key %q -> %q", e.OuterKey, e.InnerKey)
}

// ErrAlreadyRegistered indicates an entry already exists under the given
// outerKey/innerKey pair.
type ErrAlreadyRegistered struct {
	OuterKey string
	InnerKey string
}

func (e ErrAlreadyRegistered) Error() string {
	return fmt.Sprintf("an entry is already registered for key %q -> %q", e.OuterKey, e.InnerKey)
}

// Registry is a generic, two-level (outerKey -> innerKey -> value)
// in-memory store, safe for concurrent use.
type Registry[V any] struct {
	mu      sync.Mutex
	entries map[string]map[string]V
}

// New creates an empty Registry.
func New[V any]() *Registry[V] {
	return &Registry[V]{entries: map[string]map[string]V{}}
}

// Register adds value under outerKey/innerKey. It fails if an entry is
// already registered under that exact pair.
func (r *Registry[V]) Register(outerKey string, innerKey string, value V) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	inner, ok := r.entries[outerKey]
	if !ok {
		r.entries[outerKey] = map[string]V{innerKey: value}
		return nil
	}
	if _, exists := inner[innerKey]; exists {
		return ErrAlreadyRegistered{OuterKey: outerKey, InnerKey: innerKey}
	}
	inner[innerKey] = value
	return nil
}

// Get retrieves the value registered under outerKey/innerKey.
func (r *Registry[V]) Get(outerKey string, innerKey string) (V, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var zero V
	inner, ok := r.entries[outerKey]
	if !ok {
		return zero, ErrOuterKeyNotFound{OuterKey: outerKey}
	}
	value, ok := inner[innerKey]
	if !ok {
		return zero, ErrInnerKeyNotFound{OuterKey: outerKey, InnerKey: innerKey}
	}
	return value, nil
}

// Remove deletes the value registered under outerKey/innerKey. If this
// was the last remaining entry for outerKey, the outer key itself is
// pruned too, keeping the registry clean.
func (r *Registry[V]) Remove(outerKey string, innerKey string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	inner, ok := r.entries[outerKey]
	if !ok {
		return ErrOuterKeyNotFound{OuterKey: outerKey}
	}
	if _, ok := inner[innerKey]; !ok {
		return ErrInnerKeyNotFound{OuterKey: outerKey, InnerKey: innerKey}
	}
	delete(inner, innerKey)
	if len(inner) == 0 {
		delete(r.entries, outerKey)
	}
	return nil
}

// List returns all currently registered values, in no particular order.
func (r *Registry[V]) List() []V {
	r.mu.Lock()
	defer r.mu.Unlock()

	values := make([]V, 0)
	for _, inner := range r.entries {
		for _, value := range inner {
			values = append(values, value)
		}
	}
	return values
}
