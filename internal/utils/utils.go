package utils

import (
	"github.com/golang-collections/collections/set"
)

func ConvertListToSet(list []string) *set.Set {
	newSet := set.New()
	for _, v := range list {
		newSet.Insert(v)
	}
	return newSet
}
