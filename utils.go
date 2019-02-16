package plugins

import (
    "strings"
    "net/url"

    "github.com/aws/aws-sdk-go/service/ec2"
    "github.com/c-bata/go-prompt"
)

func ExtractSuggestionText(suggestions []prompt.Suggest) *[]string {
     text := make([]string, len(suggestions))
     for i, s := range suggestions {
         text[i] = s.Text
     }
     return &text
}

func UrlDecode(input *string) *string {
    if values, err := url.ParseQuery(*input); err == nil {
        for key, _ := range values {
            return &key
        }
    }
    return nil
}

func GetNameFromTags(tags *[]*ec2.Tag) *string {
    for _, tag := range *tags {
        if *tag.Key == "Name" {
            return tag.Value
        }
    }
    return nil
}

func PathToStrings(inputPath *string) *[]string {
    rawStrings := strings.Split(*inputPath, "/")
    str := ""
    resultStrings := []string{}
    for i, s := range(rawStrings) {
        l := len(s)
        if l > 0 && s[l-1] == '\\' {
            if len(str) == 0 {
                str = fmt.Sprintf("%s/%s", s, rawStrings[i+1])
            } else {
                str = fmt.Sprintf("%s/%s", str, rawStrings[i+1])
            }
        } else if len(str) > 0 {
            resultStrings = append(resultStrings, str)
            str = ""
        } else {
            resultStrings = append(resultStrings, s)
        }
    }
    return &resultStrings
}
