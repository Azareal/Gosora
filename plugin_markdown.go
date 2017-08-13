package main

//import "fmt"
import "regexp"
import "strings"

var markdown_max_depth int = 25 // How deep the parser will go when parsing Markdown strings
var markdown_unclosed_element []byte

var markdown_bold_tag_open, markdown_bold_tag_close []byte
var markdown_italic_tag_open, markdown_italic_tag_close []byte
var markdown_underline_tag_open, markdown_underline_tag_close []byte
var markdown_strike_tag_open, markdown_strike_tag_close []byte

var markdown_bold_italic *regexp.Regexp
var markdown_bold *regexp.Regexp
var markdown_italic *regexp.Regexp
var markdown_strike *regexp.Regexp
var markdown_underline *regexp.Regexp

func init() {
	plugins["markdown"] = NewPlugin("markdown","Markdown","Azareal","http://github.com/Azareal","","","",init_markdown,nil,deactivate_markdown,nil,nil)
}

func init_markdown() error {
	//plugins["markdown"].AddHook("parse_assign", markdown_regex_parse)
	plugins["markdown"].AddHook("parse_assign", markdown_parse)

	markdown_unclosed_element = []byte("<span style='color: red;'>[Unclosed Element]</span>")

	markdown_bold_tag_open = []byte("<b>")
	markdown_bold_tag_close = []byte("</b>")
	markdown_italic_tag_open = []byte("<i>")
	markdown_italic_tag_close = []byte("</i>")
	markdown_underline_tag_open = []byte("<u>")
	markdown_underline_tag_close = []byte("</u>")
	markdown_strike_tag_open = []byte("<s>")
	markdown_strike_tag_close = []byte("</s>")

	markdown_bold_italic = regexp.MustCompile(`\*\*\*(.*)\*\*\*`)
	markdown_bold = regexp.MustCompile(`\*\*(.*)\*\*`)
	markdown_italic = regexp.MustCompile(`\*(.*)\*`)
	//markdown_strike = regexp.MustCompile(`\~\~(.*)\~\~`)
	markdown_strike = regexp.MustCompile(`\~(.*)\~`)
	//markdown_underline = regexp.MustCompile(`\_\_(.*)\_\_`)
	markdown_underline = regexp.MustCompile(`\_(.*)\_`)
	return nil
}

func deactivate_markdown() {
	//plugins["markdown"].RemoveHook("parse_assign", markdown_regex_parse)
	plugins["markdown"].RemoveHook("parse_assign", markdown_parse)
}

func markdown_regex_parse(msg string) string {
	msg = markdown_bold_italic.ReplaceAllString(msg,"<i><b>$1</b></i>")
	msg = markdown_bold.ReplaceAllString(msg,"<b>$1</b>")
	msg = markdown_italic.ReplaceAllString(msg,"<i>$1</i>")
	msg = markdown_strike.ReplaceAllString(msg,"<s>$1</s>")
	msg = markdown_underline.ReplaceAllString(msg,"<u>$1</u>")
	return msg
}


// An adapter for the parser, so that the parser can call itself recursively.
// This is less for the simple Markdown elements like bold and italics and more for the really complicated ones I plan on adding at some point.
func markdown_parse(msg string) string {
	return strings.TrimSpace(_markdown_parse(msg + " ",0))
}

// Under Construction!
func _markdown_parse(msg string, n int) string {
	if n > markdown_max_depth {
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

		switch(msg[index]) {
			case '_':
				var startIndex int = index
				if (index + 1) >= len(msg) {
					break
				}

				index++
				index = markdown_skip_until_char(msg, index, '_')
				if (index - (startIndex + 1)) < 2 || index >= len(msg) {
					break
				}

				sIndex := startIndex + 1
				lIndex := index
				index++

				outbytes = append(outbytes, msg[lastElement:startIndex]...)
				outbytes = append(outbytes, markdown_underline_tag_open...)
				outbytes = append(outbytes, msg[sIndex:lIndex]...)
				outbytes = append(outbytes, markdown_underline_tag_close...)

				lastElement = index
				index--
			case '~':
				var startIndex int = index
				if (index + 1) >= len(msg) {
					break
				}

				index++
				index = markdown_skip_until_char(msg, index, '~')
				if (index - (startIndex + 1)) < 2 || index >= len(msg) {
					break
				}

				sIndex := startIndex + 1
				lIndex := index
				index++

				outbytes = append(outbytes, msg[lastElement:startIndex]...)
				outbytes = append(outbytes, markdown_strike_tag_open...)
				outbytes = append(outbytes, msg[sIndex:lIndex]...)
				outbytes = append(outbytes, markdown_strike_tag_close...)

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

					if msg[index + 1] == '*' {
						//log.Print("two asterisks")
						bold = true
						index++
						if msg[index + 1] != '*' {
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

				index = markdown_skip_until_asterisk(msg,index)

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
						outbytes = append(outbytes, markdown_unclosed_element...)
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
						outbytes = append(outbytes, markdown_unclosed_element...)
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
						outbytes = append(outbytes, markdown_unclosed_element...)
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
					outbytes = append(outbytes, markdown_unclosed_element...)
					lastElement = startIndex
					break
				}

				if sIndex < 0 || lIndex < 0 {
					//log.Print("unclosed markdown element @ sIndex < 0 || lIndex < 0")
					outbytes = append(outbytes, msg[lastElement:startIndex]...)
					outbytes = append(outbytes, markdown_unclosed_element...)
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
					outbytes = append(outbytes, markdown_bold_tag_open...)
				}
				if italic {
					outbytes = append(outbytes, markdown_italic_tag_open...)
				}

				outbytes = append(outbytes, msg[sIndex:lIndex]...)

				if italic {
					outbytes = append(outbytes, markdown_italic_tag_close...)
				}
				if bold {
					outbytes = append(outbytes, markdown_bold_tag_close...)
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

func markdown_find_char(data string ,index int ,char byte) bool {
	for ; index < len(data); index++ {
		item := data[index]
		if item > 32 {
			return (item == char)
		}
	}
	return false
}

func markdown_skip_until_char(data string, index int, char byte) int {
	for ; index < len(data); index++ {
		if data[index] == char {
			break
		}
	}
	return index
}

func markdown_skip_until_asterisk(data string, index int) int {
SwitchLoop:
	for ; index < len(data); index++ {
		switch(data[index]) {
			case 10:
				if ((index+1) < len(data)) && markdown_find_char(data,index,'*') {
					index = markdown_skip_list(data,index)
				}
			case '*': break SwitchLoop
		}
	}
	return index
}

// plugin_markdown doesn't support lists yet, but I want it to be easy to have nested lists when we do have them
func markdown_skip_list(data string, index int) int {
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
