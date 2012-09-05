package tucs

import (
	"os"
)

func in_intslice(val int, slice []int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// idx_intslice returns the index in slice of the first element equal to val
func idx_intslice(val int, slice []int) int {
	for i, v := range slice {
		if v == val {
			return i
		}
	}
	return -1
}

// PathExists returns whether the given file or directory exists or not.
func PathExists(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// EOF
