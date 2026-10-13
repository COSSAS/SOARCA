package config

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-playground/assert/v2"
)

// patch mirrors what an API handler does with a request body.
func patch(config *Config, body string) error {
	var dynamicPatch DynamicPatch
	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&dynamicPatch); err != nil {
		return err
	}
	_, err := config.Update(dynamicPatch)
	return err
}

func TestLoadReadsEnvironment(t *testing.T) {
	t.Setenv("SOARCA_ALLOWED_ORIGINS", "https://a.example, https://b.example")
	t.Setenv("MQTT_PORT", "1884")

	config, err := Load()
	assert.Equal(t, err, nil)
	assert.Equal(t, config.Static.Fins.MqttPort, 1884)
	assert.Equal(t, config.Static.Server.Port, 8080)
	assert.Equal(t, config.Dynamic().AllowedOrigins, []string{"https://a.example", "https://b.example"})
}

func TestUpdatePatchesOnlyGivenFields(t *testing.T) {
	config, err := Load()
	assert.Equal(t, err, nil)

	changed := ""
	config.OnChange(func(before Dynamic, after Dynamic) {
		changed = before.LogLevel + "->" + after.LogLevel
	})

	err = patch(config, `{"log_level":"debug"}`)
	assert.Equal(t, err, nil)
	assert.Equal(t, config.Dynamic().LogLevel, "debug")
	assert.Equal(t, config.Dynamic().AllowedOrigins, []string{"*"})
	assert.Equal(t, changed, "info->debug")
}

func TestUpdateRejectsStaticFields(t *testing.T) {
	config, err := Load()
	assert.Equal(t, err, nil)

	err = patch(config, `{"port":9000}`)
	assert.NotEqual(t, err, nil)
	assert.Equal(t, config.Static.Server.Port, 8080)
}

func TestUpdateRejectsInvalidValues(t *testing.T) {
	config, err := Load()
	assert.Equal(t, err, nil)

	err = patch(config, `{"log_level":"loud"}`)
	assert.NotEqual(t, err, nil)
	assert.Equal(t, config.Dynamic().LogLevel, "info")

	err = patch(config, `{"validation_schema_url":"ftp://example.com/schema.json"}`)
	assert.NotEqual(t, err, nil)
	assert.Equal(t, config.Dynamic().ValidationSchemaUrl, "")
}

func TestDynamicReturnsCopy(t *testing.T) {
	config, err := Load()
	assert.Equal(t, err, nil)

	dynamic := config.Dynamic()
	dynamic.AllowedOrigins[0] = "https://evil.example"
	assert.Equal(t, config.Dynamic().AllowedOrigins, []string{"*"})
}

func TestLoadRequiresMongoSettingsWhenDatabaseEnabled(t *testing.T) {
	t.Setenv("DATABASE", "true")

	_, err := Load()
	assert.NotEqual(t, err, nil)
}

func TestLoadRequiresOidcSettingsWhenAuthEnabled(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")

	_, err := Load()
	assert.NotEqual(t, err, nil)
}

func TestUpdateRejectsInvalidOrigins(t *testing.T) {
	config, err := Load()
	assert.Equal(t, err, nil)

	err = patch(config, `{"allowed_origins":["example.com"]}`)
	assert.NotEqual(t, err, nil)

	err = patch(config, `{"allowed_origins":["*","https://example.com"]}`)
	assert.NotEqual(t, err, nil)

	err = patch(config, `{"allowed_origins":["https://example.com"]}`)
	assert.Equal(t, err, nil)
	assert.Equal(t, config.Dynamic().AllowedOrigins, []string{"https://example.com"})
}

func TestFailedUpdateLeavesCurrentUntouched(t *testing.T) {
	config, err := Load()
	assert.Equal(t, err, nil)

	level := "debug"
	origins := []string{"not-a-url"}
	_, err = config.Update(DynamicPatch{LogLevel: &level, AllowedOrigins: &origins})
	assert.NotEqual(t, err, nil)
	assert.Equal(t, config.Dynamic().LogLevel, "info")
	assert.Equal(t, config.Dynamic().AllowedOrigins, []string{"*"})
}

func TestUpdateDoesNotKeepCallerSlice(t *testing.T) {
	config, err := Load()
	assert.Equal(t, err, nil)

	origins := []string{"https://a.example"}
	updated, err := config.Update(DynamicPatch{AllowedOrigins: &origins})
	assert.Equal(t, err, nil)
	assert.Equal(t, updated.AllowedOrigins, []string{"https://a.example"})

	origins[0] = "https://evil.example"
	assert.Equal(t, config.Dynamic().AllowedOrigins, []string{"https://a.example"})
}

func TestConcurrentUpdatesToDifferentFieldsBothSurvive(t *testing.T) {
	config, err := Load()
	assert.Equal(t, err, nil)

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = patch(config, `{"log_level":"debug"}`)
	}()
	err = patch(config, `{"allowed_origins":["https://gui.example"]}`)
	assert.Equal(t, err, nil)
	<-done

	assert.Equal(t, config.Dynamic().LogLevel, "debug")
	assert.Equal(t, config.Dynamic().AllowedOrigins, []string{"https://gui.example"})
}
