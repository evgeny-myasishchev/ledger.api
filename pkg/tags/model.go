package tags

import (
	"strconv"
	"strings"
)

// GetTagIDsFromString will parse string that contains tagIDs
// enclosed with {}, e.g: {1},{2},{4}
func GetTagIDsFromString(tagIDs string) []int {
	if tagIDs == "" {
		return []int{}
	}
	parts := strings.Split(tagIDs, ",")
	result := make([]int, cap(parts))
	for i, part := range parts {
		tagID, err := strconv.Atoi(part[1 : len(part)-1])
		if err != nil {
			panic(err)
		}
		result[i] = tagID
	}
	return result
}
