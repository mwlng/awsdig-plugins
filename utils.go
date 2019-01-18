package plugins

import (
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

func GetNameFromTags(tags *ec2.Tags) *string {
    for _, tag := tags {
        if *tag.Key == "Name" {
            return tag.Value
        }
    }
    return nil
}
