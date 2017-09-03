package main

//import "fmt"
import "regexp"
import "strings"

var markdownMaxDepth int = 25 // How deep the parser will go when parsing Markdown strings
var markdownUnclosedElement []byte

var markdownBoldTagOpen, markdownBoldTagClose []byte
var markdownItalicTagOpen, markdownItalicTagClose []byte
var markdownUnderlineTagOpen, markdownUnderlineTagClose []byte
var markdownStrikeTagOpen, markdownStrikeTagClose []byte

var markdownBoldItalic *regexp.Regexp
var markdownBold *regexp.Regexp
var markdownItalic *regexp.Regexp
var markdownStrike *regexp.Regexp
var markdownUnderline *regexp.Regexp

func init() {
	plugins["markdown"] = NewPlugin("markdown", "Markdown", "Azareal", "http://github.com/Azareal", "", "", "", initMarkdown, nil, deactivateMarkdown, nil, nil)
}

func initMarkdown() error {
	//plugins["markdown"].AddHook("parse_assign", markdownRegexParse)
	plugins["markdown"].AddHook("parse_assign", markdownParse)

	markdownUnclosedElement = []byte("<span style='color: red;'>[Unclosed Element]</span>")

	markdownBoldTagOpen = []byte("<b>")
	markdownBoldTagClose = []byte("</b>")
	markdownItalicTagOpen = []byte("<i>")
	markdownItalicTagClose = []byte("</i>")
	markdownUnderlineTagOpen = []byte("<u>")
	markdownUnderlineTagClose = []byte("</u>")
	markdownStrikeTagOpen = []byte("<s>")
	markdownStrikeTagClose = []byte("</s>")

	markdownBoldItalic = regexp.MustCompile(`\*\*\*(.*)\*\*\*`)
	markdownBold = regexp.MustCompile(`\*\*(.*)\*\*`)
	markdownItalic = regexp.MustCompile(`\*(.*)\*`)
	//markdownStrike = regexp.MustCompile(`\~\~(.*)\~\~`)
	markdownStrike = regexp.MustCompile(`\~(.*)\~`)
	//markdown_underline = regexp.MustCompile(`\_\_(.*)\_\_`)
	markdownUnderline = regexp.MustCompile(`\_(.*)\_`)
	return nil
}

func deactivateMarkdown() {
	//plugins["markdown"].RemoveHook("parse_assign", markdownRegexParse)
	plugins["markdown"].RemoveHook("parse_assign", markdownParse)
}

func markdownRegexParse(msg string) string {
	msg = markdownBoldItalic.ReplaceAllString(msg, "<i><b>$1</b></i>")
	msg = markdownBold.ReplaceAllString(msg, "<b>$1</b>")
	msg = markdownItalic.ReplaceAllString(msg, "<i>$1</i>")
	msg = markdownStrike.ReplaceAllString(msg, "<s>$1</s>")
	msg = markdownUnderline.ReplaceAllString(msg, "<u>$1</u>")
	return msg
}

// An adapter for the parser, so that the parser can call itself recursively.
// This is less for the simple Markdown elements like bold and italics and more for the really complicated ones I plan on adding at some point.
func markdownParse(msg string) string {
	return strings.TrimSpace(_markdownParse(msg+" ", 0))
}

// Under Construction!
func _markdownParse(msg string, n int) string {
	if n > markdownMaxDepth {
		return "<span style='color: red;'>[Markdown Error: Overflowed the max depth of 20]</span>"
	}

	var outbytes []byte
	var lastElement int
	//log.Print("enter message loop")
	//log.Print("Message: %v\n",strings.Replace(msg,"\r","\\r",-1))

	for index := 0; index < len(msg); index++ {
		/*//log.Print("--OUTER MARKDOWN LOOP START--")
		//log.Print("index",index)
		//log.Print("msg[index]",msg[index])
		//log.Print("string(msg[index])",string(msg[index]))
		//log.Print("--OUTER MARKDOWN LOOP END--")
		//log.Print(" ")*/

		switch msg[index] {
		case '_':
			var startIndex int = index
			if (index + 1) >= len(msg) {
				break
			}

			index++
			index = markdownSkipUntilChar(msg, index, '_')
			if (index-(startIndex+1)) < 2 || index >= len(msg) {
				break
			}

			sIndex := startIndex + 1
			lIndex := index
			index++

			outbytes = append(outbytes, msg[lastElement:startIndex]...)
			outbytes = append(outbytes, markdownUnderlineTagOpen...)
			outbytes = append(outbytes, msg[sIndex:lIndex]...)
			outbytes = append(outbytes, markdownUnderlineTagClose...)

			lastElement = index
			index--
		case '~':
			var startIndex int = index
			if (index + 1) >= len(msg) {
				break
			}

			index++
			index = markdownSkipUntilChar(msg, index, '~')
			if (index-(startIndex+1)) < 2 || index >= len(msg) {
				break
			}

			sIndex := startIndex + 1
			lIndex := index
			index++

			outbytes = append(outbytes, msg[lastElement:startIndex]...)
			outbytes = append(outbytes, markdownStrikeTagOpen...)
			outbytes = append(outbytes, msg[sIndex:lIndex]...)
			outbytes = append(outbytes, markdownStrikeTagClose...)

			lastElement = index
			index--
		case '*':
			//log.Print("------")
			//log.Print("[]byte(msg):",[]byte(msg))
			//log.Print("len(msg)",len(msg))
			//log.Print("start index",index)
			//log.Print("start msg[index]",msg[index])
			//log.Print("start string(msg[index])",string(msg[index]))
			//log.Print("start []byte(msg[:index])",[]byte(msg[:index]))

			var startIndex int = index
			var italic bool = true
			var bold bool
			if (index + 2) < len(msg) {
				//log.Print("start index + 1",index + 1)
				//log.Print("start msg[index]",msg[index + 1])
				//log.Print("start string(msg[index])",string(msg[index + 1]))

				if msg[index+1] == '*' {
					//log.Print("two asterisks")
					bold = true
					index++
					if msg[index+1] != '*' {
						italic = false
					} else {
						//log.Print("three asterisks")
						index++
					}
				}
			}

			//log.Print("lastElement",lastElement)
			//log.Print("startIndex:",startIndex)
			//log.Print("msg[startIndex]",msg[startIndex])
			//log.Print("string(msg[startIndex])",string(msg[startIndex]))

			//log.Print("preabrupt index",index)
			//log.Print("preabrupt msg[index]",msg[index])
			//log.Print("preabrupt string(msg[index])",string(msg[index]))
			//log.Print("preabrupt []byte(msg[:index])",[]byte(msg[:index]))
			//log.Print("preabrupt msg[:index]",msg[:index])

			// Does the string terminate abruptly?
			if (index + 1) >= len(msg) {
				break
			}

			index++

			//log.Print("preskip index",index)
			//log.Print("preskip msg[index]",msg[index])
			//log.Print("preskip string(msg[index])",string(msg[index]))

			index = markdownSkipUntilAsterisk(msg, index)

			if index >= len(msg) {
				break
			}

			//log.Print("index",index)
			//log.Print("[]byte(msg[:index])",[]byte(msg[:index]))
			//log.Print("msg[index]",msg[index])

			sIndex := startIndex
			lIndex := index
			if bold && italic {
				//log.Print("bold & italic final code")
				if (index + 3) >= len(msg) {
					//log.Print("unclosed markdown element @ exit element")
					outbytes = append(outbytes, msg[lastElement:startIndex]...)
					outbytes = append(outbytes, markdownUnclosedElement...)
					lastElement = startIndex
					break
				}
				index += 3
				sIndex += 3
			} else if bold {
				//log.Print("bold final code")
				if (index + 2) >= len(msg) {
					//log.Print("true unclosed markdown element @ exit element")
					outbytes = append(outbytes, msg[lastElement:startIndex]...)
					outbytes = append(outbytes, markdownUnclosedElement...)
					lastElement = startIndex
					break
				}
				index += 2
				sIndex += 2
			} else {
				//log.Print("italic final code")
				if (index + 1) >= len(msg) {
					//log.Print("true unclosed markdown element @ exit element")
					outbytes = append(outbytes, msg[lastElement:startIndex]...)
					outbytes = append(outbytes, markdownUnclosedElement...)
					lastElement = startIndex
					break
				}
				index++
				sIndex++
			}

			//log.Print("sIndex",sIndex)
			//log.Print("lIndex",lIndex)

			if lIndex <= sIndex {
				//log.Print("unclosed markdown element @ lIndex <= sIndex")
				outbytes = append(outbytes, msg[lastElement:startIndex]...)
				outbytes = append(outbytes, markdownUnclosedElement...)
				lastElement = startIndex
				break
			}

			if sIndex < 0 || lIndex < 0 {
				//log.Print("unclosed markdown element @ sIndex < 0 || lIndex < 0")
				outbytes = append(outbytes, msg[lastElement:startIndex]...)
				outbytes = append(outbytes, markdownUnclosedElement...)
				lastElement = startIndex
				break
			}

			//log.Print("final sIndex",sIndex)
			//log.Print("final lIndex",lIndex)
			//log.Print("final index",index)
			//log.Print("final msg[index]",msg[index])
			//log.Print("final string(msg[index])",string(msg[index]))

			//log.Print("final msg[sIndex]",msg[sIndex])
			//log.Print("final string(msg[sIndex])",string(msg[sIndex]))
			//log.Print("final msg[lIndex]",msg[lIndex])
			//log.Print("final string(msg[lIndex])",string(msg[lIndex]))

			//log.Print("[]byte(msg[:sIndex])",[]byte(msg[:sIndex]))
			//log.Print("[]byte(msg[:lIndex])",[]byte(msg[:lIndex]))

			outbytes = append(outbytes, msg[lastElement:startIndex]...)

			if bold {
				outbytes = append(outbytes, markdownBoldTagOpen...)
			}
			if italic {
				outbytes = append(outbytes, markdownItalicTagOpen...)
			}

			outbytes = append(outbytes, msg[sIndex:lIndex]...)

			if italic {
				outbytes = append(outbytes, markdownItalicTagClose...)
			}
			if bold {
				outbytes = append(outbytes, markdownBoldTagClose...)
			}

			lastElement = index
			index--
			//case '`':
			//case '_':
			//case '~':
			//case 10: // newline
		}
	}

	//log.Print("exit message loop")

	if len(outbytes) == 0 {
		return msg
	} else if lastElement < (len(msg) - 1) {
		return string(outbytes) + msg[lastElement:]
	}
	return string(outbytes)
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
	var datalen int = len(data)

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
