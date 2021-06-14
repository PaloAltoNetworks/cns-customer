package testutil

// CheckTagInTags checks if a string exist in a tag clause
func CheckTagInTags(tag string, tags []string) bool {
	for _, curr := range tags {
		if tag == curr {
			return true
		}
	}
	return false
}

// CheckTagInTagClauses checks if a string exist in every tag clause. As long as one clause is found missing the string
// it bails out.
func CheckTagInTagClauses(tag string, tagClauses [][]string) bool {
	for _, subjects := range tagClauses {
		if !CheckTagInTags(tag, subjects) {
			return false
		}
	}
	return true
}

// CheckTagExistInAnyTagClauses checks if a tag exist in any tag clauses. As long as there is one clasuse that has the string,
// it returns true.
func CheckTagExistInAnyTagClauses(tag string, tagClauses [][]string) bool {
	for _, subjects := range tagClauses {
		if CheckTagInTags(tag, subjects) {
			return true
		}
	}
	return false
}
