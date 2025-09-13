package x

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func TestPinYin(t *testing.T) {
	fmt.Println(ChineseToPinyin("离家自动化"))
	testString := "Hello 😄, 这是一段测试文本。\n\t包含换行和空格！  \n😊"
	cleanedString := CleanString(testString)
	fmt.Println("清理后的字符串:", cleanedString)
}

// CleanString 去掉表情、图案、换行、空格、回车和制表符
func CleanString(input string) string {
	// 定义正则表达式
	re := regexp.MustCompile(`[\p{So}\p{Cn}\n\r\t]+`) // 表情和图案
	cleaned := re.ReplaceAllString(input, "")         // 去掉表情和图案

	// 去掉空格
	cleaned = strings.ReplaceAll(cleaned, " ", "")

	return cleaned
}
