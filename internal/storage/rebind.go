package storage

import (
	"strconv"
	"strings"
)

// Rebind converts "?" placeholders to PostgreSQL positional placeholders.
func Rebind(query string) string {
	var out strings.Builder
	out.Grow(len(query) + 8)
	index := 1
	for _, ch := range query {
		if ch == '?' {
			out.WriteByte('$')
			out.WriteString(strconv.Itoa(index))
			index++
			continue
		}
		out.WriteRune(ch)
	}
	return out.String()
}
