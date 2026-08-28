package mongodb

import (
	"testing"
)

func TestConfigWithURI(t *testing.T) {
	tests := []struct {
		name string
		uri  string
	}{
		{
			name: "Valid URI with database",
			uri:  "mongodb://localhost:27017/mydb",
		},
		{
			name: "Valid URI with credentials",
			uri:  "mongodb://user:pass@localhost:27017/mydb",
		},
		{
			name: "Valid URI with timeout options",
			uri:  "mongodb://localhost:27017/mydb?connectTimeoutMS=5000&serverSelectionTimeoutMS=10000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{URI: tt.uri}
			if cfg.URI != tt.uri {
				t.Errorf("Config.URI = %q, want %q", cfg.URI, tt.uri)
			}
		})
	}
}
