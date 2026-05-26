package fsutil

import "path/filepath"

// PathsEqual reports whether two paths resolve to the same absolute location.
func PathsEqual(a, b string) (bool, error) {
	aAbs, err := filepath.Abs(a)
	if err != nil {
		return false, err
	}
	bAbs, err := filepath.Abs(b)
	if err != nil {
		return false, err
	}
	return aAbs == bAbs, nil
}
