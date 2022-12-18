package util

import (
	"encoding/json"
	"strconv"
	"time"

	e "github.com/pkg/errors"
	Z "github.com/rwxrob/bonzai/z"
)

func GetGeneric[T any](x *Z.Cmd, key string, f func(string) (T, error)) (T, error) {
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

func Get(x *Z.Cmd, key string) (string, error) {
	return GetGeneric(x, key, IdErr[string])
}

func GetInt(x *Z.Cmd, key string) (int, error) {
	return GetGeneric(x, key, strconv.Atoi)
}

func GetInt64(x *Z.Cmd, key string) (int64, error) {
	return GetGeneric(x, key, func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	})
}

func GetBool(x *Z.Cmd, key string) (bool, error) {
	return GetGeneric(x, key, strconv.ParseBool)
}

func GetFloat64(x *Z.Cmd, key string) (float64, error) {
	return GetGeneric(x, key, func(s string) (float64, error) {
		return strconv.ParseFloat(s, 64)
	})
}

func GetDuration(x *Z.Cmd, key string) (time.Duration, error) {
	return GetGeneric(x, key, func(s string) (time.Duration, error) {
		return time.ParseDuration(s)
	})
}

func GetCommand(x *Z.Cmd, key string) ([]string, error) {
	return GetGeneric(x, key, func(s string) ([]string, error) {
		if s == "" {
			return nil, nil
		}
		var cmd []string
		if err := json.Unmarshal([]byte(s), &cmd); err != nil {
			return nil, e.Wrapf(err, "parse %s as []string", key)
		}
		return cmd, nil
	})
}
