package x

import (
	"math"
	"strings"
)

// Levenshtein calculates the Levenshtein distance between two strings.
func Levenshtein(a, b string) int {
	lenA := len(a)
	lenB := len(b)

	distance := make([][]int, lenA+1)
	for i := range distance {
		distance[i] = make([]int, lenB+1)
	}

	for i := 0; i <= lenA; i++ {
		distance[i][0] = i
	}
	for j := 0; j <= lenB; j++ {
		distance[0][j] = j
	}

	for i := 1; i <= lenA; i++ {
		for j := 1; j <= lenB; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			distance[i][j] = minInt(
				distance[i-1][j]+1,      // Deletion
				distance[i][j-1]+1,      // Insertion
				distance[i-1][j-1]+cost, // Substitution
			)
		}
	}

	return distance[lenA][lenB]
}

// minInt returns the minimum of three integers
func minInt(x, y, z int) int {
	return int(math.Min(float64(x), math.Min(float64(y), float64(z))))
}

// Similarity calculates the similarity between two strings and returns a float64
func Similarity(a, b string) float64 {
	if a == b {
		return 1.0 // 完全相同
	}

	distance := Levenshtein(a, b)
	maxLen := maxInt(len(a), len(b))

	if maxLen == 0 {
		return 1.0 // 两个字符串都为空
	}

	return 1.0 - float64(distance)/float64(maxLen)
}

// ContainsAllChars 判断字符串 chars 中的所有字符是否都出现在字符串 str 中。
// 注意：此函数区分大小写。
func ContainsAllChars(str, chars string) bool {
	// 遍历 chars 中的每一个字符
	for _, char := range chars {
		// 检查当前字符是否在 str 中
		if !strings.ContainsRune(str, char) {
			return false // 只要有一个字符不在 str 中，就返回 false
		}
	}
	return true // 所有字符都在 str 中出现
}

// maxInt returns the maximum of two integers
func maxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}
