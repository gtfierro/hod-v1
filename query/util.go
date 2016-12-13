package query

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

func compareOrClauseLists(o1, o2 []OrClause) bool {
	var (
		found bool
	)

	if len(o1) != len(o2) {
		return false
	}

	for _, val1 := range o1 {
		found = false
		for _, val2 := range o2 {
			if compareOrClause(val1, val2) {
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

func compareOrClause(o1, o2 OrClause) bool {
	if !compareFilterSliceAsSet(o1.Terms, o2.Terms) {
		return false
	}
	if !compareFilterSliceAsSet(o1.LeftTerms, o2.LeftTerms) {
		return false
	}
	if !compareFilterSliceAsSet(o1.RightTerms, o2.RightTerms) {
		return false
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
