package config

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-playground/assert/v2"
)

func patch(config *Config, body string) error {
	return config.Update(func(dynamic *Dynamic) error {
		decoder := json.NewDecoder(strings.NewReader(body))
		decoder.DisallowUnknownFields()
		return decoder.Decode(dynamic)
	})
}

func TestLoadReadsEnvironment(t *testing.T) {
	t.Setenv("SOARCA_ALLOWED_ORIGINS", "https://a.example, https://b.example")
	t.Setenv("MQTT_PORT", "1884")

	config, err := Load("does-not-exist.env")
	assert.Equal(t, err, nil)
	assert.Equal(t, config.Static.Fins.MqttPort, 1884)
	assert.Equal(t, config.Static.Server.Port, 8080)
	assert.Equal(t, config.Dynamic().AllowedOrigins, []string{"https://a.example", "https://b.example"})
}

func TestUpdatePatchesOnlyGivenFields(t *testing.T) {
	config, err := Load("does-not-exist.env")
	assert.Equal(t, err, nil)

	changed := ""
	config.OnChange(func(previous Dynamic, current Dynamic) {
		changed = previous.LogLevel + "->" + current.LogLevel
	})

	err = patch(config, `{"log_level":"debug"}`)
	assert.Equal(t, err, nil)
	assert.Equal(t, config.Dynamic().LogLevel, "debug")
	assert.Equal(t, config.Dynamic().AllowedOrigins, []string{"*"})
	assert.Equal(t, changed, "info->debug")
}

func TestUpdateRejectsStaticFields(t *testing.T) {
	config, err := Load("does-not-exist.env")
	assert.Equal(t, err, nil)

	err = patch(config, `{"port":9000}`)
	assert.NotEqual(t, err, nil)
	assert.Equal(t, config.Static.Server.Port, 8080)
}

func TestUpdateRejectsInvalidValues(t *testing.T) {
	config, err := Load("does-not-exist.env")
	assert.Equal(t, err, nil)

	err = patch(config, `{"log_level":"loud"}`)
	assert.NotEqual(t, err, nil)
	assert.Equal(t, config.Dynamic().LogLevel, "info")

	err = patch(config, `{"validation_schema_url":"ftp://example.com/schema.json"}`)
	assert.NotEqual(t, err, nil)
	assert.Equal(t, config.Dynamic().ValidationSchemaUrl, "")
}

func TestDynamicReturnsCopy(t *testing.T) {
	config, err := Load("does-not-exist.env")
	assert.Equal(t, err, nil)

	dynamic := config.Dynamic()
	dynamic.AllowedOrigins[0] = "https://evil.example"
	assert.Equal(t, config.Dynamic().AllowedOrigins, []string{"*"})
}

func TestLoadRequiresMongoSettingsWhenDatabaseEnabled(t *testing.T) {
	t.Setenv("DATABASE", "true")

	_, err := Load("does-not-exist.env")
	assert.NotEqual(t, err, nil)
}

func TestLoadRequiresOidcSettingsWhenAuthEnabled(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")

	_, err := Load("does-not-exist.env")
	assert.NotEqual(t, err, nil)
}
