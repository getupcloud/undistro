package urlutil

import (
	"net/url"
	"path/filepath"
)

func Equal(a, b string) bool {
	au, err := url.Parse(a)
	if err != nil {
		a = filepath.Clean(a)
		b = filepath.Clean(b)
		// If urls are paths, return true only if they are an exact match
		return a == b
	}
	bu, err := url.Parse(b)
	if err != nil {
		return false
	}

	for _, u := range []*url.URL{au, bu} {
		if u.Path == "" {
			u.Path = "/"
		}
		u.Path = filepath.Clean(u.Path)
	}
	return au.String() == bu.String()
}
