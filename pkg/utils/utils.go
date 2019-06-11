package utils

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/c-bata/go-prompt"
)

func ExtractSuggestionsText(suggestions []prompt.Suggest) []string {
	text := make([]string, len(suggestions))
	for i, s := range suggestions {
		text[i] = s.Text
	}
	return text
}

func UrlDecode(input string) string {
	if value, err := url.QueryUnescape(input); err == nil {
		return value
	}
	return ""
}

func ExtractNameTag(tags []*ec2.Tag) *ec2.Tag {
	for _, tag := range tags {
		if *tag.Key == "Name" {
			return tag
		}
	}
	return nil
}

func PathToStrings(inputPath string) []string {
	strs := strings.Split(inputPath, "/")
	str := ""
	resultStrs := []string{}
	total := len(strs)
	for i, s := range strs {
		l := len(s)
		if l > 0 && s[l-1] == '\\' {
			if i+1 < total {
				if len(str) == 0 {
					str = fmt.Sprintf("%s/%s", s, strs[i+1])
				} else {
					str = fmt.Sprintf("%s/%s", str, strs[i+1])
				}
			} else {
				if len(str) == 0 {
					resultStrs = append(resultStrs, s)
				} else {
					resultStrs = append(resultStrs, str)
				}
			}
		} else if len(str) > 0 {
			resultStrs = append(resultStrs, str)
			str = ""
		} else {
			resultStrs = append(resultStrs, s)
		}
	}
	return resultStrs
}
