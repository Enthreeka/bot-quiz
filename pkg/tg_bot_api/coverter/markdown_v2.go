package coverter

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"unicode/utf16"
)

var needEscape = make(map[rune]struct{})

func init() {
	for _, r := range []rune{'_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!'} {
		needEscape[r] = struct{}{}
	}
}

func ConvertToMarkdownV2(text string, messageEntities []tgbotapi.MessageEntity) string {
	insertions := make(map[int]string)
	for _, e := range messageEntities {
		var before, after string
		if e.IsBold() {
			before = "*"
			after = "*"
		} else if e.IsItalic() {
			before = "_"
			after = "_"
		} else if e.Type == "underline" {
			before = "__"
			after = "__"
		} else if e.Type == "strikethrough" {
			before = "~"
			after = "~"
		} else if e.IsCode() {
			before = "`"
			after = "`"
		} else if e.IsPre() {
			before = "```" + e.Language
			after = "```"
		} else if e.IsTextLink() {
			before = "["
			after = "](" + e.URL + ")"
		}
		if before != "" {
			insertions[e.Offset] += before
			insertions[e.Offset+e.Length] += after
		}
	}

	input := []rune(text)
	var output []rune
	utf16pos := 0
	for _, c := range input {
		output = append(output, []rune(insertions[utf16pos])...)
		if _, has := needEscape[c]; has {
			output = append(output, '\\')
		}
		output = append(output, c)
		utf16pos += len(utf16.Encode([]rune{c}))
	}
	output = append(output, []rune(insertions[utf16pos])...)
	return string(output)
}
