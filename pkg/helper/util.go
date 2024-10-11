package helper

import "regexp"

func IsRootPath(path string) bool {
	if path == "/" {
		return true
	}

	r := regexp.MustCompile(`^[A-Z]+ /$`)
	return r.MatchString(path)
}

func ReplacePathParams(path string) string {
	r := regexp.MustCompile(`\:(?<p>[a-z0-9]+)`)
	return r.ReplaceAllString(path, `{$p}`)
}

func StringOrString(s1, s2 string) string {
	if s1 != "" {
		return s1
	}

	return s2
}

func IsByteArray(v any) bool {
	_, ok := v.([]byte)

	return ok
}
