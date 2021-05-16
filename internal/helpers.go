package internal

import "strings"

func Contains(arr []string, str string) bool {
	lowerStr := strings.ToLower(str)
	for idx := range arr {
		if strings.ToLower(arr[idx]) == lowerStr {
			return true
		}
	}
	return false
}
