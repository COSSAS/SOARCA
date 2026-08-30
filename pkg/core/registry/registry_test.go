package registry

import (
	"errors"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestRegisterAndGet(t *testing.T) {
	r := New[string]()

	err := r.Register("exec1", "step1", "value1")
	assert.Equal(t, err, nil)

	value, err := r.Get("exec1", "step1")
	assert.Equal(t, err, nil)
	assert.Equal(t, value, "value1")
}

func TestRegisterDuplicateFails(t *testing.T) {
	r := New[string]()

	err := r.Register("exec1", "step1", "value1")
	assert.Equal(t, err, nil)

	err = r.Register("exec1", "step1", "value2")
	assert.Equal(t, err, ErrAlreadyRegistered{OuterKey: "exec1", InnerKey: "step1"})
}

func TestRegisterSameOuterDifferentInner(t *testing.T) {
	r := New[string]()

	assert.Equal(t, r.Register("exec1", "step1", "value1"), nil)
	assert.Equal(t, r.Register("exec1", "step2", "value2"), nil)

	value1, err := r.Get("exec1", "step1")
	assert.Equal(t, err, nil)
	assert.Equal(t, value1, "value1")

	value2, err := r.Get("exec1", "step2")
	assert.Equal(t, err, nil)
	assert.Equal(t, value2, "value2")
}

func TestGetMissingOuterKey(t *testing.T) {
	r := New[string]()

	_, err := r.Get("missing", "step1")
	assert.Equal(t, err, ErrOuterKeyNotFound{OuterKey: "missing"})

	var target ErrOuterKeyNotFound
	assert.Equal(t, errors.As(err, &target), true)
}

func TestGetMissingInnerKey(t *testing.T) {
	r := New[string]()
	assert.Equal(t, r.Register("exec1", "step1", "value1"), nil)

	_, err := r.Get("exec1", "missing")
	assert.Equal(t, err, ErrInnerKeyNotFound{OuterKey: "exec1", InnerKey: "missing"})
}

func TestRemove(t *testing.T) {
	r := New[string]()
	assert.Equal(t, r.Register("exec1", "step1", "value1"), nil)

	err := r.Remove("exec1", "step1")
	assert.Equal(t, err, nil)

	_, err = r.Get("exec1", "step1")
	assert.Equal(t, err, ErrOuterKeyNotFound{OuterKey: "exec1"})
}

func TestRemovePrunesOuterKeyOnlyWhenEmpty(t *testing.T) {
	r := New[string]()
	assert.Equal(t, r.Register("exec1", "step1", "value1"), nil)
	assert.Equal(t, r.Register("exec1", "step2", "value2"), nil)

	err := r.Remove("exec1", "step1")
	assert.Equal(t, err, nil)

	// exec1 should still exist because step2 remains registered.
	value, err := r.Get("exec1", "step2")
	assert.Equal(t, err, nil)
	assert.Equal(t, value, "value2")

	err = r.Remove("exec1", "step2")
	assert.Equal(t, err, nil)

	_, err = r.Get("exec1", "step2")
	assert.Equal(t, err, ErrOuterKeyNotFound{OuterKey: "exec1"})
}

func TestRemoveMissingOuterKey(t *testing.T) {
	r := New[string]()

	err := r.Remove("missing", "step1")
	assert.Equal(t, err, ErrOuterKeyNotFound{OuterKey: "missing"})
}

func TestRemoveMissingInnerKey(t *testing.T) {
	r := New[string]()
	assert.Equal(t, r.Register("exec1", "step1", "value1"), nil)

	err := r.Remove("exec1", "missing")
	assert.Equal(t, err, ErrInnerKeyNotFound{OuterKey: "exec1", InnerKey: "missing"})
}

func TestList(t *testing.T) {
	r := New[string]()

	assert.Equal(t, len(r.List()), 0)

	assert.Equal(t, r.Register("exec1", "step1", "value1"), nil)
	assert.Equal(t, r.Register("exec1", "step2", "value2"), nil)
	assert.Equal(t, r.Register("exec2", "step1", "value3"), nil)

	values := r.List()
	assert.Equal(t, len(values), 3)

	seen := map[string]bool{}
	for _, v := range values {
		seen[v] = true
	}
	assert.Equal(t, seen["value1"], true)
	assert.Equal(t, seen["value2"], true)
	assert.Equal(t, seen["value3"], true)
}
