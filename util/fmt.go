package util

import (
	"fmt"
	"regexp"
)

var re = regexp.MustCompile(`{(.+?)}`)

func Fprint(format string, argsNamed map[string]any) string {
	return re.ReplaceAllStringFunc(format, func(s string) string {
		if s[0] == '{' && s[1] == '{' {
			return s
		}
		key := s[1 : len(s)-1]
		if v, ok := argsNamed[key]; ok {
			return fmt.Sprint(v)
		}
		return s
	})
}
