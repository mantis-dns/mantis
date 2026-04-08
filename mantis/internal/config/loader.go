package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	toml "github.com/pelletier/go-toml/v2"
)

// Load reads configuration from a TOML file and overlays environment variables.
// If path is empty, only defaults and env vars are used.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		if err := toml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse config file: %w", err)
		}
	}

	applyEnvOverrides(cfg)

	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

// applyEnvOverrides walks the Config struct and applies MANTIS_ prefixed env vars.
func applyEnvOverrides(cfg *Config) {
	applyEnvToStruct(reflect.ValueOf(cfg).Elem(), "MANTIS")
}

func applyEnvToStruct(v reflect.Value, prefix string) {
	t := v.Type()
	for i := range t.NumField() {
		field := t.Field(i)
		fv := v.Field(i)

		tag := field.Tag.Get("toml")
		if tag == "" || tag == "-" {
			continue
		}

		envKey := prefix + "_" + strings.ToUpper(tag)

		if field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeOf(Duration{}) {
			applyEnvToStruct(fv, envKey)
			continue
		}

		envVal, ok := os.LookupEnv(envKey)
		if !ok {
			continue
		}

		setFieldFromString(fv, envVal)
	}
}

func setFieldFromString(fv reflect.Value, val string) {
	if fv.Type() == reflect.TypeOf(Duration{}) {
		if d, err := time.ParseDuration(val); err == nil {
			fv.Set(reflect.ValueOf(Duration{d}))
		}
		return
	}

	switch fv.Kind() {
	case reflect.String:
		fv.SetString(val)
	case reflect.Int:
		if n, err := strconv.Atoi(val); err == nil {
			fv.SetInt(int64(n))
		}
	case reflect.Int64:
		if fv.Type() == reflect.TypeOf(time.Duration(0)) {
			if d, err := time.ParseDuration(val); err == nil {
				fv.SetInt(int64(d))
			}
		} else {
			if n, err := strconv.ParseInt(val, 10, 64); err == nil {
				fv.SetInt(n)
			}
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(val); err == nil {
			fv.SetBool(b)
		}
	case reflect.Slice:
		if fv.Type().Elem().Kind() == reflect.String {
			parts := strings.Split(val, ",")
			for i, p := range parts {
				parts[i] = strings.TrimSpace(p)
			}
			fv.Set(reflect.ValueOf(parts))
		}
	}
}
