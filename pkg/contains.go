package pkg

// Contains returns whether the slice contains the specified element or not
func Contains(sl []string, str string) bool {
	for _, s := range sl {
		if s == str {
			return true
		}
	}
	return false
}
