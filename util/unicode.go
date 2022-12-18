package util

import (
	"strconv"
	"strings"
)

// ParseUnicode function parses a string with unicode code points
// and returns a string with the corresponding characters
// Example: ParseUnicode("U+1F1FAU+1F1E6") returns "🇺🇦" (Ukraine flag).
func ParseUnicode(s string) (string, error) {
	var r []rune
	codePoints := strings.Split(s, "U+")[1:]

	for _, cp := range codePoints {
		i, err := strconv.ParseInt(cp, 16, 32)
		if err != nil {
			return "", err
		}
		r = append(r, rune(i))
	}
	return string(r), nil
}
