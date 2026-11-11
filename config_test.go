package config

import "testing"

// TestBindEnvOverridesDefault asserts the FIX-04 behaviour: an environment
// variable must override a value registered with SetDefault.
func TestBindEnvOverridesDefault(t *testing.T) {
	t.Setenv("APP_APP_PORT", "9999")

	cfg := New()
	cfg.SetDefault("app.port", 8080)
	cfg.BindEnv("APP_")

	if got := cfg.GetString("app.port"); got != "9999" {
		t.Fatalf("BindEnv must override defaults: want %q, got %q", "9999", got)
	}
}

// TestBindEnvOverridesUnderscoreKey asserts that an env var can override a
// configuration key whose name contains an underscore (e.g. auth.jwt_secret),
// which is required to rotate secrets through the environment.
func TestBindEnvOverridesUnderscoreKey(t *testing.T) {
	t.Setenv("APP_AUTH_JWT_SECRET", "env-secret")

	cfg := New()
	cfg.SetDefault("auth.jwt_secret", "change-me-in-production")
	cfg.BindEnv("APP_")

	if got := cfg.GetString("auth.jwt_secret"); got != "env-secret" {
		t.Fatalf("BindEnv must override underscore keys: want %q, got %q", "env-secret", got)
	}
}

// TestBindEnvUnknownKeyFallsBackToDottedKey asserts the historical behaviour is
// preserved for keys that do not exist in the configuration yet.
func TestBindEnvUnknownKeyFallsBackToDottedKey(t *testing.T) {
	t.Setenv("APP_AUTH_LOGIN_RATE_LIMIT", "5")

	cfg := New()
	cfg.BindEnv("APP_")

	if got := cfg.GetString("auth.login.rate.limit"); got != "5" {
		t.Fatalf("BindEnv must fall back to the dotted key: want %q, got %q", "5", got)
	}
}

// TestBindEnvOverridesFileValue asserts env vars also override values loaded
// from a file.
func TestBindEnvOverridesFileValue(t *testing.T) {
	t.Setenv("APP_DB_HOST", "env-host")

	cfg := New()
	cfg.Set("db.host", "file-host")
	cfg.BindEnv("APP_")

	if got := cfg.GetString("db.host"); got != "env-host" {
		t.Fatalf("BindEnv must override file values: want %q, got %q", "env-host", got)
	}
}

// TestBindEnvRespectsPrefix asserts variables without the prefix are ignored.
func TestBindEnvRespectsPrefix(t *testing.T) {
	t.Setenv("OTHER_DB_HOST", "ignored")

	cfg := New()
	cfg.SetDefault("db.host", "default-host")
	cfg.BindEnv("APP_")

	if got := cfg.GetString("db.host"); got != "default-host" {
		t.Fatalf("non-prefixed env var must be ignored: want %q, got %q", "default-host", got)
	}
}
