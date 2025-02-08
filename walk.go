package main

func isMatched(path string, exprs []string) bool {
	for _, expr := range exprs {
		if PathMatchUnvalidated(expr, path) {
			return true
		}
	}
	return false
}

func Iif[T any](condition bool, a, b T) T {
	if condition {
		return a
	}
	return b
}
