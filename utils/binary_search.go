package utils

import "strings"

func BinarySearch(list []string, value string) (int, bool) {
	low := 0
	high := len(list)
	return binarySearch(list, value, low, high)
}

func binarySearch(list []string, value string, low int, high int) (int, bool) {
	mid := low + int((high-low)/2)
	found := false
	if low == high {
		return mid, found
	}
	result := strings.Compare(value, list[mid])
	if result > 0 {
		return binarySearch(list, value, mid+1, high)
	} else if result < 0 {
		return binarySearch(list, value, low, mid)
	} else {
		found = true
	}
	return mid, found
}
