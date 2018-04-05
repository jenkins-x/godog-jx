package util

import (
	"regexp"
	"sort"
	"strings"
)

// RegexpSplit splits a string into an array using the regexSep as a separator
func RegexpSplit(text string, regexSeperator string) []string {
	reg := regexp.MustCompile(regexSeperator)
	indexes := reg.FindAllStringIndex(text, -1)
	lastIdx := 0
	result := make([]string, len(indexes)+1)
	for i, element := range indexes {
		result[i] = text[lastIdx:element[0]]
		lastIdx = element[1]
	}
	result[len(indexes)] = text[lastIdx:]
	return result
}

// StringIndexes returns all the indices where the value occurs in the given string
func StringIndexes(text string, value string) []int {
	answer := []int{}
	t := text
	valueLen := len(value)
	offset := 0
	for {
		idx := strings.Index(t, value)
		if idx < 0 {
			break
		}
		answer = append(answer, idx+offset)
		offset += valueLen
		t = t[idx+valueLen:]
	}
	return answer
}

func StringArrayIndex(array []string, value string) int {
	for i, v := range array {
		if v == value {
			return i
		}
	}
	return -1
}

// FirstNotEmptyString returns the first non empty string or the empty string if none can be found
func FirstNotEmptyString(values ...string) string {
	if values != nil {
		for _, v := range values {
			if v != "" {
				return v
			}
		}
	}
	return ""
}

// SortedMapKeys returns the sorted keys of the given map
func SortedMapKeys(m map[string]string) []string {
	answer := []string{}
	for k, _ := range m {
		answer = append(answer, k)
	}
	sort.Strings(answer)
	return answer
}

func ReverseStrings(a []string) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}

// StringArrayToLower returns a string slice with all the values converted to lower case
func StringArrayToLower(values []string) []string {
	answer := []string{}
	for _, v := range values {
		answer = append(answer, strings.ToLower(v))
	}
	return answer
}
