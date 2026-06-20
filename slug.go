package main

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

func sanitize(s string) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' {
			return unicode.ToLower(r)
		}
		return ' '
	}, s)
	s = strings.TrimSpace(s)
	s = nonAlphaNum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 60 {
		s = s[:60]
		if i := strings.LastIndex(s, "-"); i > 30 {
			s = s[:i]
		}
	}
	return s
}

func slugFromFilename(name string) string {
	stem := strings.TrimSuffix(name, filepath.Ext(name))
	return sanitize(stem)
}

func makeID(filename string) string {
	date := time.Now().Format("2006-01-02")
	slug := slugFromFilename(filename)
	if slug == "" {
		slug = "untitled"
	}
	return date + "_" + slug
}

func modelSlug(model string) string {
	s := strings.ReplaceAll(model, ":", "")
	s = strings.ReplaceAll(s, "/", "-")
	return s
}
