package util

// StripPrefixesAndTrailingSlash skips any leading "./" or "/" such that the
// result index begins with another string. A result of "." coerces to the
// empty string "" because the current directory is handled by the guest.
//
// Results are the offset/len pair which is an optimization to avoid re-slicing
// overhead, as this function is called for every path operation.
//
// Note: Relative paths should be handled by the guest, as that's what knows
// what the current directory is. However, paths that escape the current
// directory e.g. "../.." have been found in `tinygo test` and this
// implementation takes care to avoid it.
func StripPrefixesAndTrailingSlash(path string) string {
	// strip trailing slashes
	pathLen := len(path)
	for ; pathLen > 0 && path[pathLen-1] == '/'; pathLen-- {
	}

	pathI := 0
loop:
	for pathI < pathLen {
		switch path[pathI] {
		case '/':
			pathI++
		case '.':
			nextI := pathI + 1
			if nextI < pathLen && path[nextI] == '/' {
				pathI = nextI + 1
			} else if nextI == pathLen {
				pathI = nextI
			} else {
				break loop
			}
		default:
			break loop
		}
	}
	return path[pathI:pathLen]
}
