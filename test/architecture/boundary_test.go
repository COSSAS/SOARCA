// Package architecture_test enforces the orchestrator/transport boundary.
//
// The orchestrator must stay drivable by any transport (HTTP today, gRPC or a
// CLI later). That only holds if the core never depends on transport concerns,
// so this is asserted mechanically rather than by convention.
package architecture_test

import (
	"os/exec"
	"strings"
	"testing"
)

// corePackages are the orchestrator-side package trees.
var corePackages = []string{
	"soarca/internal/runtime/...",
	"soarca/internal/runs/...",
	"soarca/internal/services/...",
	"soarca/internal/storage/...",
}

// forbiddenInCore are inbound-transport dependencies. Note that net/http is
// deliberately absent: capabilities such as http and openc2 make outbound calls
// and legitimately need it. What the core must never gain is a web framework,
// route registration, or auth middleware.
var forbiddenInCore = []string{
	"github.com/gin-gonic/gin",
	"github.com/COSSAS/gauth",
	"github.com/swaggo/",
	"soarca/internal/transport",
	"soarca/pkg/api",
}

func deps(t *testing.T, packages ...string) []string {
	t.Helper()

	args := append([]string{"list", "-deps"}, packages...)
	cmd := exec.Command("go", args...)
	cmd.Dir = "../.."

	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list -deps failed: %v", err)
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n")
}

func TestCoreDoesNotDependOnTransport(t *testing.T) {
	for _, dep := range deps(t, corePackages...) {
		for _, forbidden := range forbiddenInCore {
			if strings.Contains(dep, forbidden) {
				t.Errorf("orchestrator core depends on transport concern %q (via %q).\n"+
					"The core must stay drivable by gRPC or a CLI; move this to internal/transport.",
					forbidden, dep)
			}
		}
	}
}

// TestDetectorWorks guards the test above: the transport layer genuinely does
// depend on gin, so if this stops finding it the check has silently stopped
// checking anything.
func TestDetectorWorks(t *testing.T) {
	for _, dep := range deps(t, "soarca/internal/transport/...") {
		if strings.Contains(dep, "github.com/gin-gonic/gin") {
			return
		}
	}
	t.Fatal("expected the transport layer to depend on gin; the boundary detector is not working")
}
