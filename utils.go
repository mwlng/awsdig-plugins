package plugins

import (
    "net/url"

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
