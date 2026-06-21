package main

import "strings"

func modelSlug(model string) string {
	s := strings.ReplaceAll(model, ":", "")
	s = strings.ReplaceAll(s, "/", "-")
	return s
}
