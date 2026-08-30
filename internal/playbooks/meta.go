package playbooks
package playbooks

import "time"

// Meta is the lightweight summary used when listing stored playbooks.
type Meta struct {
	ID          string
	Name        string
	Description string
	ValidFrom   time.Time
	ValidUntil  time.Time
	Labels      []string
}