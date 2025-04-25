package executor

import "strings"

func Topic(s ...string) string {
	return strings.Join(s, ".")
}
