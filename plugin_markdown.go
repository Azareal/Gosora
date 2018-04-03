package main

//import "fmt"
import (
	"strings"

	"./common"
)

var markdownMaxDepth = 25 // How deep the parser will go when parsing Markdown strings
var markdownUnclosedElement []byte

var markdownBoldTagOpen []byte
var markdownBoldTagClose []byte
var markdownItalicTagOpen []byte
var markdownItalicTagClose []byte
var markdownUnderlineTagOpen []byte
var markdownUnderlineTagClose []byte
var markdownStrikeTagOpen []byte
var markdownStrikeTagClose []byte

func init() {
	common.Plugins["markdown"] = common.NewPlugin("markdown", "Markdown", "Azareal", "http://github.com/Azareal", "", "", "", initMarkdown, nil, deactivateMarkdown, nil, nil)
}

func initMarkdown() error {
	common.Plugins["markdown"].AddHook("parse_assign", markdownParse)

	markdownUnclosedElement = []byte("<span style='color: red;'>[Unclosed Element]</span>")

	markdownBoldTagOpen = []byte("<b>")
	markdownBoldTagClose = []byte("</b>")
	markdownItalicTagOpen = []byte("<i>")
	markdownItalicTagClose = []byte("</i>")
	markdownUnderlineTagOpen = []byte("<u>")
	markdownUnderlineTagClose = []byte("</u>")
	markdownStrikeTagOpen = []byte("<s>")
	markdownStrikeTagClose = []byte("</s>")
	return nil
}

func deactivateMarkdown() {
	common.Plugins["markdown"].RemoveHook("parse_assign", markdownParse)
}

// An adapter for the parser, so that the parser can call itself recursively.
// This is less for the simple Markdown elements like bold and italics and more for the really complicated ones I plan on adding at some point.
func markdownParse(msg string) string {
	msg = _markdownParse(msg+" ", 0)
	if msg[len(msg)-1] == ' ' {
		msg = msg[:len(msg)-1]
	}
	return msg
}

// Under Construction!
func _markdownParse(msg string, n int) string {
	if n > markdownMaxDepth {
		return "<span style='color: red;'>[Markdown Error: Overflowed the max depth of 20]</span>"
	}

	var outbytes []byte
	var lastElement int
	common.DebugLogf("Initial Message: %+v\n", strings.Replace(msg, "\r", "\\r", -1))

	for index := 0; index < len(msg); index++ {
		switch msg[index] {
		// TODO: Do something slightly less hacky for skipping URLs
		case '/':
			if len(msg) > (index+2) && msg[index+1] == '/' {
				for ; index < len(msg) && msg[index] != ' '; index++ {

				}
				index--
				continue
			}
		case '_':
			var startIndex = index
			if (index + 1) >= len(msg) {
				break
			}

			index++
			index = markdownSkipUntilChar(msg, index, '_')
			if (index-(startIndex+1)) < 1 || index >= len(msg) {
				break
			}

			sIndex := startIndex + 1
			lIndex := index
			index++

			outbytes = append(outbytes, msg[lastElement:startIndex]...)
			outbytes = append(outbytes, markdownUnderlineTagOpen...)
			// TODO: Implement this without as many type conversions
			outbytes = append(outbytes, []byte(_markdownParse(msg[sIndex:lIndex], n+1))...)
			outbytes = append(outbytes, markdownUnderlineTagClose...)

			lastElement = index
			index--
		case '~':
			var startIndex = index
			if (index + 1) >= len(msg) {
				break
			}

			index++
			index = markdownSkipUntilChar(msg, index, '~')
			if (index-(startIndex+1)) < 1 || index >= len(msg) {
				break
			}

			sIndex := startIndex + 1
			lIndex := index
			index++

			outbytes = append(outbytes, msg[lastElement:startIndex]...)
			outbytes = append(outbytes, markdownStrikeTagOpen...)
			// TODO: Implement this without as many type conversions
			outbytes = append(outbytes, []byte(_markdownParse(msg[sIndex:lIndex], n+1))...)
			outbytes = append(outbytes, markdownStrikeTagClose...)

			lastElement = index
			index--
		case '*':
			var startIndex = index
			var italic = true
			var bold = false
			if (index + 2) < len(msg) {
				if msg[index+1] == '*' {
					bold = true
					index++
					if msg[index+1] != '*' {
						italic = false
					} else {
						index++
					}
				}
			}

			// Does the string terminate abruptly?
			if (index + 1) >= len(msg) {
				break
			}
			index++

			index = markdownSkipUntilAsterisk(msg, index)
			if index >= len(msg) {
				break
			}

			sIndex := startIndex
			lIndex := index
			if bold && italic {
				if (index + 3) >= len(msg) {
					outbytes = append(outbytes, msg[lastElement:startIndex]...)
					lastElement = startIndex
					break
				}
				index += 3
				sIndex += 3
			} else if bold {
				if (index + 2) >= len(msg) {
					outbytes = append(outbytes, msg[lastElement:startIndex]...)
					lastElement = startIndex
					break
				}
				index += 2
				sIndex += 2
			} else {
				if (index + 1) >= len(msg) {
					outbytes = append(outbytes, msg[lastElement:startIndex]...)
					lastElement = startIndex
					break
				}
				index++
				sIndex++
			}

			if lIndex <= sIndex {
				outbytes = append(outbytes, msg[lastElement:startIndex]...)
				lastElement = startIndex
				break
			}

			if sIndex < 0 || lIndex < 0 {
				outbytes = append(outbytes, msg[lastElement:startIndex]...)
				lastElement = startIndex
				break
			}

			outbytes = append(outbytes, msg[lastElement:startIndex]...)

			if bold {
				outbytes = append(outbytes, markdownBoldTagOpen...)
			}
			if italic {
				outbytes = append(outbytes, markdownItalicTagOpen...)
			}

			// TODO: Implement this without as many type conversions
			outbytes = append(outbytes, []byte(_markdownParse(msg[sIndex:lIndex], n+1))...)

			if italic {
				outbytes = append(outbytes, markdownItalicTagClose...)
			}
			if bold {
				outbytes = append(outbytes, markdownBoldTagClose...)
			}

			lastElement = index
			index--
		case '\\':
			if (index + 1) < len(msg) {
				if isMarkdownStartChar(msg[index+1]) && msg[index+1] != '\\' {
					outbytes = append(outbytes, msg[lastElement:index]...)
					index++
					lastElement = index
				}
			}
			//case '`':
			//case 10: // newline
		}
	}

	if len(outbytes) == 0 {
		return msg
	} else if lastElement < (len(msg) - 1) {
		msg = string(outbytes) + msg[lastElement:]
		return msg
	}
	return string(outbytes)
}

func isMarkdownStartChar(char byte) bool {
	return char == '\\' || char == '~' || char == '_' || char == 10 || char == '`' || char == '*'
}

func markdownFindChar(data string, index int, char byte) bool {
	for ; index < len(data); index++ {
		item := data[index]
		if item > 32 {
			return (item == char)
		}
	}
	return false
}

func markdownSkipUntilChar(data string, index int, char byte) int {
	for ; index < len(data); index++ {
		if data[index] == char {
			break
		}
	}
	return index
}

func markdownSkipUntilAsterisk(data string, index int) int {
SwitchLoop:
	for ; index < len(data); index++ {
		switch data[index] {
		case 10:
			if ((index + 1) < len(data)) && markdownFindChar(data, index, '*') {
				index = markdownSkipList(data, index)
			}
		case '*':
			break SwitchLoop
		}
	}
	return index
}

// plugin_markdown doesn't support lists yet, but I want it to be easy to have nested lists when we do have them
func markdownSkipList(data string, index int) int {
	var lastNewline int
	var datalen = len(data)

	for ; index < datalen; index++ {
	SkipListInnerLoop:
		if data[index] == 10 {
			lastNewline = index
			for ; index < datalen; index++ {
				if data[index] > 32 {
					break
				} else if data[index] == 10 {
					goto SkipListInnerLoop
				}
			}

			if index >= datalen {
				if data[index] != '*' && data[index] != '-' {
					if (lastNewline + 1) < datalen {
						return lastNewline + 1
					}
					return lastNewline
				}
			}
		}
	}

	return index
}
