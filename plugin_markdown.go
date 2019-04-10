package main

import (
	"strings"

	"github.com/Azareal/Gosora/common"
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
var markdownQuoteTagOpen []byte
var markdownQuoteTagClose []byte
var markdownH1TagOpen []byte
var markdownH1TagClose []byte

func init() {
	common.Plugins.Add(&common.Plugin{UName: "markdown", Name: "Markdown", Author: "Azareal", URL: "https://github.com/Azareal", Init: initMarkdown, Deactivate: deactivateMarkdown})
}

func initMarkdown(plugin *common.Plugin) error {
	plugin.AddHook("parse_assign", markdownParse)

	markdownUnclosedElement = []byte("<red>[Unclosed Element]</red>")

	markdownBoldTagOpen = []byte("<b>")
	markdownBoldTagClose = []byte("</b>")
	markdownItalicTagOpen = []byte("<i>")
	markdownItalicTagClose = []byte("</i>")
	markdownUnderlineTagOpen = []byte("<u>")
	markdownUnderlineTagClose = []byte("</u>")
	markdownStrikeTagOpen = []byte("<s>")
	markdownStrikeTagClose = []byte("</s>")
	markdownQuoteTagOpen = []byte("<blockquote>")
	markdownQuoteTagClose = []byte("</blockquote>")
	markdownH1TagOpen = []byte("<h2>")
	markdownH1TagClose = []byte("</h2>")
	return nil
}

func deactivateMarkdown(plugin *common.Plugin) {
	plugin.RemoveHook("parse_assign", markdownParse)
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
		return "<red>[Markdown Error: Overflowed the max depth of 20]</red>"
	}

	var outbytes []byte
	var lastElement int
	var breaking = false
	common.DebugLogf("Initial Message: %+v\n", strings.Replace(msg, "\r", "\\r", -1))

	for index := 0; index < len(msg); index++ {
		var simpleMatch = func(char byte, o []byte, c []byte) {
			var startIndex = index
			if (index + 1) >= len(msg) {
				breaking = true
				return
			}

			index++
			index = markdownSkipUntilChar(msg, index, char)
			if (index-(startIndex+1)) < 1 || index >= len(msg) {
				breaking = true
				return
			}

			sIndex := startIndex + 1
			lIndex := index
			index++

			outbytes = append(outbytes, msg[lastElement:startIndex]...)
			outbytes = append(outbytes, o...)
			// TODO: Implement this without as many type conversions
			outbytes = append(outbytes, []byte(_markdownParse(msg[sIndex:lIndex], n+1))...)
			outbytes = append(outbytes, c...)

			lastElement = index
			index--
		}

		var startLine = func() {
			var startIndex = index
			if (index + 1) >= len(msg) /*|| (index + 2) >= len(msg)*/ {
				breaking = true
				return
			}
			index++

			index = markdownSkipUntilNotChar(msg, index, 32)
			if (index + 1) >= len(msg) {
				breaking = true
				return
			}
			//index++

			index = markdownSkipUntilStrongSpace(msg, index)
			sIndex := startIndex + 1
			lIndex := index
			index++

			outbytes = append(outbytes, msg[lastElement:startIndex]...)
			outbytes = append(outbytes, markdownH1TagOpen...)
			// TODO: Implement this without as many type conversions
			//fmt.Println("msg[sIndex:lIndex]:", string(msg[sIndex:lIndex]))
			// TODO: Quick hack to eliminate trailing spaces...
			outbytes = append(outbytes, []byte(strings.TrimSpace(_markdownParse(msg[sIndex:lIndex], n+1)))...)
			outbytes = append(outbytes, markdownH1TagClose...)

			lastElement = index
			index--
		}

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
			simpleMatch('_', markdownUnderlineTagOpen, markdownUnderlineTagClose)
			if breaking {
				break
			}
		case '~':
			simpleMatch('~', markdownStrikeTagOpen, markdownStrikeTagClose)
			if breaking {
				break
			}
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

			var preBreak = func() {
				outbytes = append(outbytes, msg[lastElement:startIndex]...)
				lastElement = startIndex
			}

			sIndex := startIndex
			lIndex := index
			if bold && italic {
				if (index + 3) >= len(msg) {
					preBreak()
					break
				}
				index += 3
				sIndex += 3
			} else if bold {
				if (index + 2) >= len(msg) {
					preBreak()
					break
				}
				index += 2
				sIndex += 2
			} else {
				if (index + 1) >= len(msg) {
					preBreak()
					break
				}
				index++
				sIndex++
			}

			if lIndex <= sIndex {
				preBreak()
				break
			}
			if sIndex < 0 || lIndex < 0 {
				preBreak()
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
		// TODO: Add a inline quote variant
		case '`':
			simpleMatch('`', markdownQuoteTagOpen, markdownQuoteTagClose)
			if breaking {
				break
			}
		case 10: // newline
			if (index + 1) >= len(msg) {
				break
			}
			index++

			if msg[index] != '#' {
				continue
			}
			startLine()
			if breaking {
				break
			}
		case '#':
			if index != 0 {
				continue
			}
			startLine()
			if breaking {
				break
			}
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

func markdownSkipUntilNotChar(data string, index int, char byte) int {
	for ; index < len(data); index++ {
		if data[index] != char {
			break
		}
	}
	return index
}

func markdownSkipUntilStrongSpace(data string, index int) int {
	var inSpace = false
	for ; index < len(data); index++ {
		if inSpace && data[index] == 32 {
			index--
			break
		} else if data[index] == 32 {
			inSpace = true
		} else if data[index] < 32 {
			break
		} else {
			inSpace = false
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
