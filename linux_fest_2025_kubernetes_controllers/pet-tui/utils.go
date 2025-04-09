package main

import "strings"

// bar creates a visual representation of a value
func bar(val int) string {
	full := val / 10
	return strings.Repeat("█", full) + strings.Repeat("░", 10-full)
}
