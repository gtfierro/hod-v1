package lang

func compareFilterSliceAsSet(s1, s2 []Filter) bool {
	var (
		found bool
	)

	if len(s1) != len(s2) {
		return false
	}

	for _, val1 := range s1 {
		found = false
		for _, val2 := range s2 {
			if val1.Equals(val2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func comparePathSliceAsSet(s1, s2 []PathPattern) bool {
	var (
		found bool
	)

	if len(s1) != len(s2) {
		return false
	}

	for _, val1 := range s1 {
		found = false
		for _, val2 := range s2 {
			if val1 == val2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
