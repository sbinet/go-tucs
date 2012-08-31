package tucs

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
// EOF
