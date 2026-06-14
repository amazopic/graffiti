package greet

import "strings"

func Hello(name string) string {
	return "hi " + upper(name)
}

func upper(s string) string {
	return strings.ToUpper(s)
}
