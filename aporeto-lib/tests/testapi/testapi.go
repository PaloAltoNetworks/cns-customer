package testapi

import (
	"context"
	"time"

	"go.aporeto.io/elemental"
	"go.aporeto.io/manipulate"
)

// GetMany takes an identifiable list and parameters and returns an interface to that object list if that list of objects exists in the backend
// This is for testing use only
func GetMany(ctx context.Context, m manipulate.Manipulator, object elemental.Identifiables, filterOptions []manipulate.ContextOption) (interface{}, error) {

	subctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	mctx := manipulate.NewContext(subctx, filterOptions...)

	if err := m.RetrieveMany(mctx, object); err != nil {
		return nil, err
	}

	return object, nil

}

// Get takes an identifiable and parameters and returns an interface to that object if that object exists in the backend
// This is for testing use only
func Get(ctx context.Context, m manipulate.Manipulator, object elemental.Identifiable, filterOptions []manipulate.ContextOption) (interface{}, error) {

	subctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	mctx := manipulate.NewContext(subctx, filterOptions...)

	if err := m.Retrieve(mctx, object); err != nil {
		return nil, err
	}

	return object, nil

}

// // CheckTagInTags checks if a string exist in a tag clause
// func CheckTagInTags(tag string, tags []string) bool {
// 	for _, curr := range tags {
// 		if tag == curr {
// 			return true
// 		}
// 	}
// 	return false
// }

// // CheckTagInTagClauses checks if a string exist in every tag clause. As long as one clause is found missing the string
// // it bails out.
// func CheckTagInTagClauses(tag string, tagClauses [][]string) bool {
// 	for _, subjects := range tagClauses {
// 		if !CheckTagInTags(tag, subjects) {
// 			return false
// 		}
// 	}
// 	return true
// }

// // CheckTagExistInAnyTagClauses checks if a tag exist in any tag clauses. As long as there is one clasuse that has the string,
// // it returns true.
// func CheckTagExistInAnyTagClauses(tag string, tagClauses [][]string) bool {
// 	for _, subjects := range tagClauses {
// 		if CheckTagInTags(tag, subjects) {
// 			return true
// 		}
// 	}
// 	return false
// }
