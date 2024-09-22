package server

import "regexp"

func isRootPath(path string) bool {
	if path == "/" {
		return true
	}

	r := regexp.MustCompile(`^[A-Z]+ /$`)
	return r.MatchString(path)
}
