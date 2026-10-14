package config

import (
	"encoding/json"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/joho/godotenv"
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

func TestLoadAcceptsExistingGauthSettings(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("OIDC_ISSUER", "https://authentik.example/application/o/soarca/")
	t.Setenv("OIDC_CLIENT_ID", "soarca")

	config, err := Load()
	assert.Equal(t, err, nil)
	assert.Equal(t, config.Static.Auth.OidcEnabled(), true)
	assert.Equal(t, config.Static.Auth.EntraEnabled(), false)
	assert.Equal(t, config.Static.Auth.AdminGroup, "soarca_admin")
}

func TestLoadReadsEntraSettings(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_ENTRA_TENANT_IDS", " tenant-a, tenant-b,, ")
	t.Setenv("AUTH_ENTRA_API_CLIENT_ID", "api")
	t.Setenv("AUTH_ENTRA_REQUIRED_USER_SCOPE", "")

	config, err := Load()
	assert.Equal(t, err, nil)
	entra := config.Static.Auth.Entra
	assert.Equal(t, entra.TenantIds, []string{"tenant-a", "tenant-b"})
	assert.Equal(t, entra.RequiredUserScope, "access_as_user")
	assert.Equal(t, entra.Authority, "https://login.microsoftonline.com")
}

func TestLoadAcceptsApiKeysAsOnlyProvider(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_API_KEYS_FILE", "/etc/soarca/keys.json")

	_, err := Load()
	assert.Equal(t, err, nil)
}

func TestLoadRejectsHalfConfiguredProviders(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("OIDC_ISSUER", "https://authentik.example/application/o/soarca/")
	_, err := Load()
	assert.NotEqual(t, err, nil)

	t.Setenv("OIDC_ISSUER", "")
	t.Setenv("AUTH_ENTRA_TENANT_IDS", "tenant-a")
	_, err = Load()
	assert.NotEqual(t, err, nil)
}

func TestLoadIgnoresAuthSettingsWhenAuthDisabled(t *testing.T) {
	// The shipped docker-compose.yaml sets OIDC_CLIENT_ID with auth disabled.
	t.Setenv("AUTH_ENABLED", "false")
	t.Setenv("OIDC_CLIENT_ID", "some client ID")

	_, err := Load()
	assert.Equal(t, err, nil)
}

// envSettings returns every setting SOARCA reads, keyed by its env tag.
func envSettings() map[string]reflect.StructTag {
	settings := map[string]reflect.StructTag{}
	var collect func(typ reflect.Type)
	collect = func(typ reflect.Type) {
		for index := 0; index < typ.NumField(); index++ {
			field := typ.Field(index)
			if field.Type.Kind() == reflect.Struct {
				collect(field.Type)
			} else if name := field.Tag.Get("env"); name != "" {
				settings[name] = field.Tag
			}
		}
	}
	collect(reflect.TypeOf(Static{}))
	collect(reflect.TypeOf(Dynamic{}))
	return settings
}

// envExample returns every variable in .env.example, commented out or not,
// with its value unquoted.
func envExample(t *testing.T) map[string]string {
	content, err := os.ReadFile("../../.env.example")
	assert.Equal(t, err, nil)
	listed := map[string]string{}
	for _, match := range regexp.MustCompile(`(?m)^#?\s*([A-Z][A-Z0-9_]+):\s*(.*)$`).FindAllStringSubmatch(string(content), -1) {
		listed[match[1]] = strings.Trim(strings.TrimSpace(match[2]), `"`)
	}
	return listed
}

// TestEnvExampleListsEverySetting keeps .env.example in step with the
// structs above: every env tag must appear in it, commented out or not,
// and it must not list variables SOARCA doesn't read.
func TestEnvExampleListsEverySetting(t *testing.T) {
	listed := envExample(t)
	known := envSettings()

	for name := range known {
		if _, ok := listed[name]; !ok {
			t.Errorf("%s is read by SOARCA but missing from .env.example", name)
		}
	}
	for name := range listed {
		if _, ok := known[name]; !ok {
			t.Errorf("%s is in .env.example but SOARCA doesn't read it", name)
		}
	}
}

// TestEnvExampleShowsDefaults makes sure .env.example doesn't suggest a
// default that differs from the code. Settings without a default may hold
// any example value.
func TestEnvExampleShowsDefaults(t *testing.T) {
	listed := envExample(t)
	for name, tag := range envSettings() {
		fallback, hasDefault := tag.Lookup("envDefault")
		value, ok := listed[name]
		if hasDefault && ok && value != fallback {
			t.Errorf("%s is %q in .env.example but defaults to %q", name, value, fallback)
		}
	}
}

// TestEnvExampleLoads checks that the example, as shipped, is valid.
func TestEnvExampleLoads(t *testing.T) {
	values, err := godotenv.Read("../../.env.example")
	assert.Equal(t, err, nil)
	for name, value := range values {
		t.Setenv(name, value)
	}

	config, err := Load()
	assert.Equal(t, err, nil)
	assert.Equal(t, config.Static.Auth.Issuer, "https://authentikuri:9443/application/o/soarca/")
}
