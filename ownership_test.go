package config

import (
	"errors"
	"strings"
	"testing"
)

func TestAllAndGetMapReturnCopies(t *testing.T) {
	cfg := New()
	cfg.Set("db.host", "localhost")
	cfg.Set("db.tags", []string{"a", "b"})

	all := cfg.All()
	all["db"].(map[string]interface{})["host"] = "mutated"
	if cfg.GetString("db.host") != "localhost" {
		t.Fatal("All() must return a deep copy")
	}

	m := cfg.GetMap("db")
	m["host"] = "mutated2"
	if cfg.GetString("db.host") != "localhost" {
		t.Fatal("GetMap() must return a deep copy")
	}

	slice := cfg.GetStringSlice("db.tags")
	slice[0] = "z"
	if cfg.GetStringSlice("db.tags")[0] != "a" {
		t.Fatal("GetStringSlice must copy []string")
	}
}

func TestLookupHelpers(t *testing.T) {
	cfg := New()
	cfg.Set("port", 8080)
	cfg.Set("debug", true)
	cfg.Set("name", "onyx")
	cfg.Set("bad", []int{1})

	if v, err := cfg.LookupInt("port"); err != nil || v != 8080 {
		t.Fatalf("LookupInt: %v %v", v, err)
	}
	if _, err := cfg.LookupInt("missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("LookupInt missing: %v", err)
	}
	if _, err := cfg.LookupInt("bad"); !errors.Is(err, ErrTypeMismatch) {
		t.Fatalf("LookupInt type: %v", err)
	}
	if v, err := cfg.LookupBool("debug"); err != nil || !v {
		t.Fatalf("LookupBool: %v %v", v, err)
	}
	if v, err := cfg.LookupString("name"); err != nil || v != "onyx" {
		t.Fatalf("LookupString: %v %v", v, err)
	}
}

func TestMergeNil(t *testing.T) {
	cfg := New()
	cfg.Set("a", 1)
	cfg.Merge(nil)
	if cfg.GetInt("a") != 1 {
		t.Fatal("Merge(nil) must be no-op")
	}
}

func TestXMLFlatLimitation(t *testing.T) {
	cfg := New()
	xml := `<app port="8080"><name>onyx</name></app>`
	if err := cfg.LoadReader(strings.NewReader(xml), "xml"); err != nil {
		t.Fatal(err)
	}
	// Flat keys: attributes and text nodes as dotted paths, not nested maps.
	if cfg.GetString("app.port") != "8080" {
		t.Fatalf("attr: %v", cfg.Get("app.port"))
	}
	if cfg.GetString("app.name") != "onyx" {
		t.Fatalf("text: %v", cfg.Get("app.name"))
	}
	if cfg.GetMap("app") != nil {
		t.Fatal("flat XML must not create nested maps for elements")
	}
}
