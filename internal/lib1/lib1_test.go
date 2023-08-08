package lib1_test

import (
	"soarca/internal/lib1"
	"testing"
)

func TestAbs(t *testing.T) {
	got := lib1.Somestruct{"test"}
	if got.Name != "test" {
		t.Errorf("Somtestruct(test) = ; want test")
	}
}
