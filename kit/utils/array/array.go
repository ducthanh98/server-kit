package array

func Contains[T comparable](arr []T, str T) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
