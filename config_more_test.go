package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadJSONAndGetters(t *testing.T) {
	cfg := New()
	src := `{"app":{"port":8080,"debug":true,"rate":1.5,"timeout":"5s","tags":["a","b"],"nested":{"x":1}}}`
	if err := cfg.LoadReader(strings.NewReader(src), "json"); err != nil {
		t.Fatalf("LoadReader: %v", err)
	}
	if cfg.GetInt("app.port") != 8080 {
		t.Fatalf("GetInt: %d", cfg.GetInt("app.port"))
	}
	if !cfg.GetBool("app.debug") {
		t.Fatal("GetBool")
	}
	if cfg.GetFloat("app.rate") != 1.5 {
		t.Fatalf("GetFloat: %v", cfg.GetFloat("app.rate"))
	}
	if cfg.GetDuration("app.timeout") != 5*time.Second {
		t.Fatalf("GetDuration: %v", cfg.GetDuration("app.timeout"))
	}
	tags := cfg.GetStringSlice("app.tags")
	if len(tags) != 2 || tags[0] != "a" {
		t.Fatalf("GetStringSlice: %#v", tags)
	}
	if cfg.GetMap("app.nested")["x"].(float64) != 1 {
		t.Fatalf("GetMap: %#v", cfg.GetMap("app.nested"))
	}
	if cfg.GetString("missing", "fb") != "fb" {
		t.Fatal("fallback")
	}
}

func TestLoadYAMLFileAndMerge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	if err := os.WriteFile(path, []byte("db:\n  host: file-host\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := New()
	cfg.SetDefault("db.host", "default")
	cfg.SetDefault("db.port", 5432)
	if err := cfg.LoadFile(path, ""); err != nil {
		t.Fatal(err)
	}
	other := New()
	other.Set("db.host", "merged")
	cfg.Merge(other)
	if cfg.GetString("db.host") != "merged" {
		t.Fatalf("merge: %s", cfg.GetString("db.host"))
	}
	if cfg.GetInt("db.port") != 5432 {
		t.Fatalf("default preserved: %d", cfg.GetInt("db.port"))
	}
}

func TestLoadTOMLAndENV(t *testing.T) {
	cfg := New()
	if err := cfg.LoadReader(strings.NewReader("name = \"onyx\"\n"), "toml"); err != nil {
		t.Fatal(err)
	}
	if cfg.GetString("name") != "onyx" {
		t.Fatalf("toml: %s", cfg.GetString("name"))
	}
	cfg2 := New()
	if err := cfg2.LoadReader(strings.NewReader("# c\nFOO=bar\n"), "env"); err != nil {
		t.Fatal(err)
	}
	if cfg2.GetString("FOO") != "bar" {
		t.Fatalf("env: %s", cfg2.GetString("FOO"))
	}
}

func TestUnsupportedFormat(t *testing.T) {
	cfg := New()
	if err := cfg.LoadReader(strings.NewReader("{}"), "ini"); err == nil {
		t.Fatal("expected unsupported format error")
	}
}

func TestSetOverwritesNested(t *testing.T) {
	cfg := New()
	cfg.Set("a.b", "x")
	cfg.Set("a.b", "y")
	if cfg.GetString("a.b") != "y" {
		t.Fatal(cfg.GetString("a.b"))
	}
	all := cfg.All()
	if all["a"] == nil {
		t.Fatal("All missing nested map")
	}
}
