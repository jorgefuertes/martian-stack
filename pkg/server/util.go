package server

import "regexp"

func isRootPath(path string) bool {
	if path == "/" {
		return true
	}

	r := regexp.MustCompile(`^[A-Z]+ /$`)
	return r.MatchString(path)
}

func replacePathParams(path string) string {
	r := regexp.MustCompile(`\:(?<p>[a-z0-9]+)`)
	return r.ReplaceAllString(path, `{$p}`)
}

func stringOrString(s1, s2 string) string {
	if s1 != "" {
		return s1
	}

	return s2
}
