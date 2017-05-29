package main

//import "fmt"
import "regexp"
//import "strings"

var markdown_max_depth int = 25 // How deep the parser will go when parsing Markdown strings
var markdown_unclosed_element []byte

var markdown_bold_tag_open []byte
var markdown_bold_tag_close []byte
var markdown_italic_tag_open []byte
var markdown_italic_tag_close []byte
var markdown_underline_tag_open []byte
var markdown_underline_tag_close []byte
var markdown_strike_tag_open []byte
var markdown_strike_tag_close []byte

var markdown_bold_italic *regexp.Regexp
var markdown_bold *regexp.Regexp
var markdown_italic *regexp.Regexp
var markdown_strike *regexp.Regexp
var markdown_underline *regexp.Regexp

func init() {
	plugins["markdown"] = NewPlugin("markdown","Markdown","Azareal","http://github.com/Azareal","","","",init_markdown,nil,deactivate_markdown)
}

func init_markdown() {
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
}

func deactivate_markdown() {
	//plugins["markdown"].RemoveHook("parse_assign", markdown_regex_parse)
	plugins["markdown"].RemoveHook("parse_assign", markdown_parse)
}

func markdown_regex_parse(data interface{}) interface{} {
	msg := data.(string)
	msg = markdown_bold_italic.ReplaceAllString(msg,"<i><b>$1</b></i>")
	msg = markdown_bold.ReplaceAllString(msg,"<b>$1</b>")
	msg = markdown_italic.ReplaceAllString(msg,"<i>$1</i>")
	msg = markdown_strike.ReplaceAllString(msg,"<s>$1</s>")
	msg = markdown_underline.ReplaceAllString(msg,"<u>$1</u>")
	return msg
}


// An adapter for the parser, so that the parser can call itself recursively.
// This is less for the simple Markdown elements like bold and italics and more for the really complicated ones I plan on adding at some point.
func markdown_parse(data interface{}) interface{} {
	return _markdown_parse(data.(string) + " ",0)
}

// Under Construction!
func _markdown_parse(msg string, n int) string {
	if n > markdown_max_depth {
		return "<span style='color: red;'>[Markdown Error: Overflowed the max depth of 20]</span>"
	}
	
	var outbytes []byte
	var lastElement int
	//fmt.Println("enter message loop")
	//fmt.Printf("Message: %v\n",strings.Replace(msg,"\r","\\r",-1))
	
	for index := 0; index < len(msg); index++ {
		/*//fmt.Println("--OUTER MARKDOWN LOOP START--")
		//fmt.Println("index",index)
		//fmt.Println("msg[index]",msg[index])
		//fmt.Println("string(msg[index])",string(msg[index]))
		//fmt.Println("--OUTER MARKDOWN LOOP END--")
		//fmt.Println(" ")*/
		
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
				//fmt.Println("------")
				//fmt.Println("[]byte(msg):",[]byte(msg))
				//fmt.Println("len(msg)",len(msg))
				//fmt.Println("start index",index)
				//fmt.Println("start msg[index]",msg[index])
				//fmt.Println("start string(msg[index])",string(msg[index]))
				//fmt.Println("start []byte(msg[:index])",[]byte(msg[:index]))
				
				var startIndex int = index
				var italic bool = true
				var bold bool
				if (index + 2) < len(msg) {
					//fmt.Println("start index + 1",index + 1)
					//fmt.Println("start msg[index]",msg[index + 1])
					//fmt.Println("start string(msg[index])",string(msg[index + 1]))
					
					if msg[index + 1] == '*' {
						//fmt.Println("two asterisks")
						bold = true
						index++
						if msg[index + 1] != '*' {
							italic = false
						} else {
							//fmt.Println("three asterisks")
							index++
						}
					}
				}
				
				//fmt.Println("lastElement",lastElement)
				//fmt.Println("startIndex:",startIndex)
				//fmt.Println("msg[startIndex]",msg[startIndex])
				//fmt.Println("string(msg[startIndex])",string(msg[startIndex]))
				
				//fmt.Println("preabrupt index",index)
				//fmt.Println("preabrupt msg[index]",msg[index])
				//fmt.Println("preabrupt string(msg[index])",string(msg[index]))
				//fmt.Println("preabrupt []byte(msg[:index])",[]byte(msg[:index]))
				//fmt.Println("preabrupt msg[:index]",msg[:index])
				
				// Does the string terminate abruptly?
				if (index + 1) >= len(msg) {
					break
				}
				
				index++
				
				//fmt.Println("preskip index",index)
				//fmt.Println("preskip msg[index]",msg[index])
				//fmt.Println("preskip string(msg[index])",string(msg[index]))
				
				index = markdown_skip_until_asterisk(msg,index)
				
				if index >= len(msg) {
					break
				}
				
				//fmt.Println("index",index)
				//fmt.Println("[]byte(msg[:index])",[]byte(msg[:index]))
				//fmt.Println("msg[index]",msg[index])
				
				sIndex := startIndex
				lIndex := index
				if bold && italic {
					//fmt.Println("bold & italic final code")
					if (index + 3) >= len(msg) {
						//fmt.Println("unclosed markdown element @ exit element")
						outbytes = append(outbytes, msg[lastElement:startIndex]...)
						outbytes = append(outbytes, markdown_unclosed_element...)
						lastElement = startIndex
						break
					}
					index += 3
					sIndex += 3
				} else if bold {
					//fmt.Println("bold final code")
					if (index + 2) >= len(msg) {
						//fmt.Println("true unclosed markdown element @ exit element")
						outbytes = append(outbytes, msg[lastElement:startIndex]...)
						outbytes = append(outbytes, markdown_unclosed_element...)
						lastElement = startIndex
						break
					}
					index += 2
					sIndex += 2
				} else {
					//fmt.Println("italic final code")
					if (index + 1) >= len(msg) {
						//fmt.Println("true unclosed markdown element @ exit element")
						outbytes = append(outbytes, msg[lastElement:startIndex]...)
						outbytes = append(outbytes, markdown_unclosed_element...)
						lastElement = startIndex
						break
					}
					index++
					sIndex++
				}
				
				//fmt.Println("sIndex",sIndex)
				//fmt.Println("lIndex",lIndex)
				
				if lIndex <= sIndex {
					//fmt.Println("unclosed markdown element @ lIndex <= sIndex")
					outbytes = append(outbytes, msg[lastElement:startIndex]...)
					outbytes = append(outbytes, markdown_unclosed_element...)
					lastElement = startIndex
					break
				}
				
				if sIndex < 0 || lIndex < 0 {
					//fmt.Println("unclosed markdown element @ sIndex < 0 || lIndex < 0")
					outbytes = append(outbytes, msg[lastElement:startIndex]...)
					outbytes = append(outbytes, markdown_unclosed_element...)
					lastElement = startIndex
					break
				}
				
				//fmt.Println("final sIndex",sIndex)
				//fmt.Println("final lIndex",lIndex)
				//fmt.Println("final index",index)
				//fmt.Println("final msg[index]",msg[index])
				//fmt.Println("final string(msg[index])",string(msg[index]))
				
				//fmt.Println("final msg[sIndex]",msg[sIndex])
				//fmt.Println("final string(msg[sIndex])",string(msg[sIndex]))
				//fmt.Println("final msg[lIndex]",msg[lIndex])
				//fmt.Println("final string(msg[lIndex])",string(msg[lIndex]))
				
				//fmt.Println("[]byte(msg[:sIndex])",[]byte(msg[:sIndex]))
				//fmt.Println("[]byte(msg[:lIndex])",[]byte(msg[:lIndex]))
				
				outbytes = append(outbytes, msg[lastElement:startIndex]...)
				
				if bold {
					outbytes = append(outbytes, markdown_bold_tag_open...)
				}
				if italic {
					outbytes = append(outbytes, markdown_italic_tag_open...)
				}
				
				outbytes = append(outbytes, msg[sIndex:lIndex]...)
				
				if bold {
					outbytes = append(outbytes, markdown_bold_tag_close...)
				}
				if italic {
					outbytes = append(outbytes, markdown_italic_tag_close...)
				}
				
				lastElement = index
				index--
			//case '`':
			//case '_':
			//case '~':
			//case 10: // newline
		}
	}
	
	//fmt.Println("exit message loop")
	//fmt.Println(" ")
	
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
