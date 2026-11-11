// Package config provides a flexible configuration loader for Go applications.
// It supports multiple file formats (YAML, TOML, JSON, XML, ENV) and environment variables.
package config

import (
	"bufio"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// ErrNotFound is returned by Lookup* helpers when a key is absent or nil.
var ErrNotFound = errors.New("config: key not found")

// ErrTypeMismatch is returned by Lookup* helpers when a value cannot be
// converted to the requested type.
var ErrTypeMismatch = errors.New("config: type mismatch")

// Config holds all configuration data as a nested map.
//
// Concurrency: Config is NOT safe for concurrent use. Callers that share a
// Config across goroutines must provide their own synchronization.
type Config struct {
	data map[string]interface{}
}

// New creates a new empty Config instance.
func New() *Config {
	return &Config{
		data: make(map[string]interface{}),
	}
}

// LoadFile loads a configuration file with an optional explicit format.
// If format is empty, it is detected from the file extension.
func (c *Config) LoadFile(path string, format string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", path, err)
	}

	if format == "" {
		format = detectFormat(path)
	}
	return c.LoadReader(bytes.NewReader(data), format)
}

// LoadReader loads configuration from an io.Reader with the given format.
// Supported formats: "yaml", "yml", "toml", "json", "xml", "env".
func (c *Config) LoadReader(r io.Reader, format string) error {
	var parsed map[string]interface{}
	var err error

	switch strings.ToLower(format) {
	case "yaml", "yml":
		parsed, err = parseYAML(r)
	case "toml":
		parsed, err = parseTOML(r)
	case "json":
		parsed, err = parseJSON(r)
	case "xml":
		parsed, err = parseXML(r)
	case "env":
		parsed, err = parseENV(r)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("error parsing %s: %w", format, err)
	}

	c.mergeMap(parsed)
	return nil
}

// LoadFiles loads multiple files in order. Later files override earlier values.
func (c *Config) LoadFiles(paths ...string) error {
	for _, p := range paths {
		if err := c.LoadFile(p, ""); err != nil {
			return err
		}
	}
	return nil
}

// Merge merges another Config into this one. The other Config's values take precedence.
// A nil other is a no-op.
func (c *Config) Merge(other *Config) {
	if other == nil {
		return
	}
	c.mergeMap(other.data)
}

// SetDefault sets a default value for a key if it does not already exist.
// Supports nested keys using dot notation (e.g. "database.host").
func (c *Config) SetDefault(key string, value interface{}) {
	if c.Get(key) == nil {
		c.Set(key, value)
	}
}

// Set sets a value for a key, overwriting any existing value.
// Supports nested keys using dot notation.
func (c *Config) Set(key string, value interface{}) {
	parts := strings.Split(key, ".")
	last := len(parts) - 1
	m := c.data
	for i, part := range parts {
		if i == last {
			m[part] = value
		} else {
			next, ok := m[part]
			if !ok {
				newMap := make(map[string]interface{})
				m[part] = newMap
				m = newMap
			} else {
				nextMap, ok := next.(map[string]interface{})
				if !ok {
					newMap := make(map[string]interface{})
					m[part] = newMap
					m = newMap
				} else {
					m = nextMap
				}
			}
		}
	}
}

// Get retrieves a value by key (dot notation). Returns nil if not found.
func (c *Config) Get(key string) interface{} {
	parts := strings.Split(key, ".")
	m := c.data
	for i, part := range parts {
		val, ok := m[part]
		if !ok {
			return nil
		}
		if i == len(parts)-1 {
			return val
		}
		next, ok := val.(map[string]interface{})
		if !ok {
			return nil
		}
		m = next
	}
	return nil
}

// GetString returns the value as a string, or fallback if not found.
func (c *Config) GetString(key string, fallback ...string) string {
	val := c.Get(key)
	if val == nil {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return ""
	}
	s, ok := val.(string)
	if !ok {
		return fmt.Sprintf("%v", val)
	}
	return s
}

// GetInt returns the value as int64, or fallback if not found.
//
// Compatibility note: if the key exists but cannot be converted, GetInt returns
// 0 (or the fallback is NOT used). Prefer LookupInt when you need to distinguish
// missing keys from conversion failures.
func (c *Config) GetInt(key string, fallback ...int64) int64 {
	val := c.Get(key)
	if val == nil {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return 0
	}
	switch v := val.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return i
		}
	}
	return 0
}

// LookupInt returns the int64 value for key, or ErrNotFound / ErrTypeMismatch.
func (c *Config) LookupInt(key string) (int64, error) {
	val := c.Get(key)
	if val == nil {
		return 0, fmt.Errorf("%w: %s", ErrNotFound, key)
	}
	switch v := val.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case string:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("%w: %s", ErrTypeMismatch, key)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("%w: %s", ErrTypeMismatch, key)
	}
}

// LookupString returns the string value for key, or ErrNotFound.
// Non-string values are stringified (same as GetString) and returned successfully.
func (c *Config) LookupString(key string) (string, error) {
	val := c.Get(key)
	if val == nil {
		return "", fmt.Errorf("%w: %s", ErrNotFound, key)
	}
	s, ok := val.(string)
	if !ok {
		return fmt.Sprintf("%v", val), nil
	}
	return s, nil
}

// LookupBool returns the bool value for key, or ErrNotFound / ErrTypeMismatch.
func (c *Config) LookupBool(key string) (bool, error) {
	val := c.Get(key)
	if val == nil {
		return false, fmt.Errorf("%w: %s", ErrNotFound, key)
	}
	if b, ok := val.(bool); ok {
		return b, nil
	}
	if s, ok := val.(string); ok {
		b, err := strconv.ParseBool(s)
		if err != nil {
			return false, fmt.Errorf("%w: %s", ErrTypeMismatch, key)
		}
		return b, nil
	}
	return false, fmt.Errorf("%w: %s", ErrTypeMismatch, key)
}

// GetBool returns the value as bool, or fallback if not found.
func (c *Config) GetBool(key string, fallback ...bool) bool {
	val := c.Get(key)
	if val == nil {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return false
	}
	b, ok := val.(bool)
	if ok {
		return b
	}
	if s, ok := val.(string); ok {
		if b, err := strconv.ParseBool(s); err == nil {
			return b
		}
	}
	return false
}

// GetFloat returns the value as float64, or fallback if not found.
func (c *Config) GetFloat(key string, fallback ...float64) float64 {
	val := c.Get(key)
	if val == nil {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return 0.0
	}
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f
		}
	}
	return 0.0
}

// GetDuration returns the value as time.Duration, or fallback if not found.
// Accepts strings like "5s", "10m", "1h" or numeric values as seconds.
func (c *Config) GetDuration(key string, fallback ...time.Duration) time.Duration {
	val := c.Get(key)
	if val == nil {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return 0
	}
	switch v := val.(type) {
	case time.Duration:
		return v
	case int, int64, float64:
		sec := toFloat64(v)
		return time.Duration(sec * float64(time.Second))
	case string:
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		if sec, err := strconv.ParseFloat(v, 64); err == nil {
			return time.Duration(sec * float64(time.Second))
		}
	}
	return 0
}

// GetStringSlice returns the value as []string, or fallback if not found.
func (c *Config) GetStringSlice(key string, fallback ...[]string) []string {
	val := c.Get(key)
	if val == nil {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return []string{}
	}
	switch v := val.(type) {
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = fmt.Sprintf("%v", item)
		}
		return result
	case []string:
		result := make([]string, len(v))
		copy(result, v)
		return result
	case string:
		parts := strings.Split(v, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		return parts
	}
	return []string{}
}

// GetMap returns a deep copy of the nested map at key, or nil if not found /
// not a map. Mutating the result does not affect the Config.
func (c *Config) GetMap(key string) map[string]interface{} {
	val := c.Get(key)
	if val == nil {
		return nil
	}
	m, ok := val.(map[string]interface{})
	if !ok {
		return nil
	}
	return cloneMap(m)
}

// All returns a deep copy of the entire configuration map.
// Mutating the result does not affect the Config.
func (c *Config) All() map[string]interface{} {
	return cloneMap(c.data)
}

// BindEnv loads environment variables with an optional prefix into the config.
// Variable names are lowercased and underscores (_) are replaced with dots (.).
// Example: prefix "APP_" + env "APP_DATABASE_HOST" becomes key "database.host".
//
// Environment variables OVERRIDE both values loaded from a file and defaults
// registered with SetDefault. (Before FIX-04 this method never overrode an
// existing key, which meant env vars were silently ignored for every key that
// had a default.)
func (c *Config) BindEnv(prefix string) {
	prefix = strings.ToUpper(prefix)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		name, value := parts[0], parts[1]
		if prefix != "" && !strings.HasPrefix(name, prefix) {
			continue
		}
		key := strings.ToLower(name)
		if prefix != "" {
			key = strings.TrimPrefix(key, strings.ToLower(prefix))
		}
		key = strings.ReplaceAll(key, "_", ".")
		c.Set(c.resolveEnvKey(strings.Split(key, ".")), value)
	}
}

// resolveEnvKey maps the underscore-derived segments of an environment variable
// name onto an existing configuration key when possible.
//
// Environment variable names cannot express whether an underscore is a nesting
// separator or part of a key, so "APP_AUTH_JWT_SECRET" naively becomes
// "auth.jwt.secret". When the configuration already contains a key literally
// named "auth.jwt_secret" (as config.yaml does), that existing key is preferred
// so environment variables can override underscore-containing keys such as
// jwt_secret, bot_token or api_key_seed. When no existing key matches, the
// plain dotted key is used (the historical behaviour).
func (c *Config) resolveEnvKey(segments []string) string {
	m := c.data
	var resolved []string
	for i := 0; i < len(segments); i++ {
		joined := strings.Join(segments[i:], "_")
		if _, ok := m[joined]; ok {
			resolved = append(resolved, joined)
			return strings.Join(resolved, ".")
		}
		next, ok := m[segments[i]].(map[string]interface{})
		if !ok {
			return strings.Join(segments, ".")
		}
		resolved = append(resolved, segments[i])
		m = next
	}
	return strings.Join(resolved, ".")
}

// -------------------- internal helpers --------------------

func detectFormat(path string) string {
	ext := filepath.Ext(path)
	switch strings.ToLower(ext) {
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".json":
		return "json"
	case ".xml":
		return "xml"
	case ".env":
		return "env"
	default:
		return ""
	}
}

func parseYAML(r io.Reader) (map[string]interface{}, error) {
	var data map[string]interface{}
	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func parseTOML(r io.Reader) (map[string]interface{}, error) {
	var data map[string]interface{}
	_, err := toml.DecodeReader(r, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func parseJSON(r io.Reader) (map[string]interface{}, error) {
	var data map[string]interface{}
	dec := json.NewDecoder(r)
	if err := dec.Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func parseXML(r io.Reader) (map[string]interface{}, error) {
	flat, err := parseXMLToFlatMap(r)
	if err != nil {
		return nil, err
	}
	// Convert dotted flat keys into nested maps so Get("app.port") works.
	nested := New()
	for k, v := range flat {
		nested.Set(k, v)
	}
	return nested.data, nil
}

// parseXMLToFlatMap is a simple XML extractor: element paths and attributes
// become dotted keys. It is not a general nested XML document model.
func parseXMLToFlatMap(r io.Reader) (map[string]interface{}, error) {
	dec := xml.NewDecoder(r)
	result := make(map[string]interface{})
	var currentKey string
	var currentValue strings.Builder
	var stack []string

	for {
		token, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch se := token.(type) {
		case xml.StartElement:
			stack = append(stack, se.Name.Local)
			currentKey = strings.Join(stack, ".")
			currentValue.Reset()
			for _, attr := range se.Attr {
				key := currentKey + "." + attr.Name.Local
				result[key] = attr.Value
			}
		case xml.EndElement:
			if currentValue.Len() > 0 {
				result[currentKey] = currentValue.String()
				currentValue.Reset()
			}
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			if len(stack) > 0 {
				currentKey = strings.Join(stack, ".")
			} else {
				currentKey = ""
			}
		case xml.CharData:
			text := strings.TrimSpace(string(se))
			if text != "" {
				currentValue.WriteString(text)
			}
		}
	}
	return result, nil
}

func parseENV(r io.Reader) (map[string]interface{}, error) {
	scanner := bufio.NewScanner(r)
	data := make(map[string]interface{})
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		data[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return data, nil
}

// mergeMap recursively merges src into c.data (src takes precedence).
func (c *Config) mergeMap(src map[string]interface{}) {
	mergeMaps(c.data, src)
}

func mergeMaps(dst, src map[string]interface{}) {
	for key, val := range src {
		if existing, ok := dst[key]; ok {
			existingMap, ok1 := existing.(map[string]interface{})
			valMap, ok2 := val.(map[string]interface{})
			if ok1 && ok2 {
				mergeMaps(existingMap, valMap)
				continue
			}
		}
		dst[key] = val
	}
}

func toFloat64(v interface{}) float64 {
	switch n := v.(type) {
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case float64:
		return n
	case float32:
		return float64(n)
	default:
		return 0
	}
}

func cloneMap(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return nil
	}
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = cloneValue(v)
	}
	return dst
}

func cloneValue(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		return cloneMap(t)
	case []interface{}:
		out := make([]interface{}, len(t))
		for i := range t {
			out[i] = cloneValue(t[i])
		}
		return out
	case []string:
		out := make([]string, len(t))
		copy(out, t)
		return out
	default:
		return v
	}
}
