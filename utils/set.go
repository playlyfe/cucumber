package utils

func SetAdd(set []string, fields ...string) []string {
	if set == nil {
		set = []string{}
	}
	for _, field := range fields {
		pos, exists := BinarySearch(set, field)
		if exists {
			continue
		} else {
			set = append(set[:pos], append([]string{field}, set[pos:]...)...)
		}
	}
	return set
}

func SetRemove(set []string, fields ...string) []string {
	if set == nil {
		set = []string{}
	}
	for _, field := range fields {
		pos, exists := BinarySearch(set, field)
		if !exists {
			continue
		} else {
			set = append(set[:pos], set[pos+1:]...)
		}
	}
	return set
}

func SetExists(set []string, value string) bool {
	if set == nil {
		set = []string{}
	}
	_, exists := BinarySearch(set, value)
	return exists
}

func SetCreate(set []string) []string {
	if set == nil {
		set = []string{}
	}
	return SetAdd([]string{}, set...)
}
