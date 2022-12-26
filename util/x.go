package util

import (
	"encoding/json"
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"
	"time"

	e "github.com/pkg/errors"
	Z "github.com/rwxrob/bonzai/z"
)

func GetF[T any](x *Z.Cmd, key string, f func(string) (T, error)) (T, error) {
	dummy := *new(T)
	s, err := x.Get(key)
	if err != nil {
		return dummy, e.Wrapf(err, "get %s", key)
	}
	res, err := f(s)
	if err != nil {
		return dummy, e.Wrapf(err, "parse %s as %T", key, dummy)
	}
	return res, nil
}

func Get[T any](x *Z.Cmd, key string) (T, error) {
	return GetF(x, key, FromString[T])
}

func GetEnum[T comparable](x *Z.Cmd, key string, values []T) (T, error) {
	y, err := Get[T](x, key)
	if err != nil {
		return y, err
	}
	for _, v := range values {
		if y == v {
			return y, nil
		}
	}
	return y, e.Errorf("invalid value %v for %s (must be one of %v)", y, key, values)
}

func FromString[T any](x string) (T, error) {
	dummy := *new(T)
	y, err := fromStringGen[T](x)
	if err != nil {
		return dummy, err
	}
	return any(y).(T), nil
}

func fromStringGen[T any](x string) (any, error) {
	switch dummy := any(*new(T)).(type) {
	case string:
		return x, nil
	case int:
		return strconv.Atoi(x)
	case uint:
		return ParseUint(x)
	case int64:
		return strconv.ParseInt(x, 10, 64)
	case uint64:
		return strconv.ParseUint(x, 10, 64)
	case bool:
		return strconv.ParseBool(x)
	case float64:
		return strconv.ParseFloat(x, 64)
	case time.Duration:
		return time.ParseDuration(x)
	default:
		if x == "" {
			return nil, nil
		}
		var cmd T
		if err := json.Unmarshal([]byte(x), &cmd); err != nil {
			return nil, e.Wrapf(err, "parse %s as %T", x, dummy)
		}
		return cmd, nil
	}
}

func ParseUint(s string) (uint, error) {
	x, err := strconv.ParseUint(s, 10, 32)
	return uint(x), err
}

func InitFromDefs(dyn template.FuncMap, name string, defs map[string]string, keys []string) {
	for k, v := range defs {
		dyn[k] = func() string { return v }
	}
	sort.Strings(keys)
	dyn[name] = func() string {
		return strings.Join(Map(func(k string) string {
			return fmt.Sprintf("      %s - %s", k, defs[k])
		}, keys), "\n")
	}
}

func ShortcutsFromDefs(keys []string) map[string][]string {
	return MapFromKV(
		keys,
		Map(
			func(v string) []string { return []string{"var", "set", v} },
			keys,
		),
	)
}
