package main
//import "log"
//import "fmt"
import "bytes"
//import "strings"
import "strconv"
import "regexp"
import "time"
import "math/rand"

var random *rand.Rand
var bbcode_invalid_number []byte
var bbcode_missing_tag []byte

var bbcode_bold *regexp.Regexp
var bbcode_italic *regexp.Regexp
var bbcode_underline *regexp.Regexp
var bbcode_strikethrough *regexp.Regexp
var bbcode_url *regexp.Regexp
var bbcode_url_label *regexp.Regexp
var bbcode_quotes *regexp.Regexp

func init() {
	plugins["bbcode"] = NewPlugin("bbcode","BBCode","Azareal","http://github.com/Azareal","","","",init_bbcode,nil,deactivate_bbcode)
}

func init_bbcode() {
	//plugins["bbcode"].AddHook("parse_assign", bbcode_parse_without_code)
	plugins["bbcode"].AddHook("parse_assign", bbcode_full_parse)
	
	bbcode_invalid_number = []byte("<span style='color: red;'>[Invalid Number]</span>")
	bbcode_missing_tag = []byte("<span style='color: red;'>[Missing Tag]</span>")
	
	bbcode_bold = regexp.MustCompile(`(?s)\[b\](.*)\[/b\]`)
	bbcode_italic = regexp.MustCompile(`(?s)\[i\](.*)\[/i\]`)
	bbcode_underline = regexp.MustCompile(`(?s)\[u\](.*)\[/u\]`)
	bbcode_strikethrough = regexp.MustCompile(`(?s)\[s\](.*)\[/s\]`)
	urlpattern := `(http|https|ftp|mailto*)(:??)\/\/([\.a-zA-Z\/]+)`
	bbcode_url = regexp.MustCompile(`\[url\]` + urlpattern + `\[/url\]`)
	bbcode_url_label = regexp.MustCompile(`(?s)\[url=` + urlpattern + `\](.*)\[/url\]`)
	bbcode_quotes = regexp.MustCompile(`\[quote\](.*)\[/quote\]`)
	
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func deactivate_bbcode() {
	//plugins["bbcode"].RemoveHook("parse_assign", bbcode_parse_without_code)
	plugins["bbcode"].RemoveHook("parse_assign", bbcode_full_parse)
}

func bbcode_regex_parse(data interface{}) interface{} {
	msg := data.(string)
	msg = bbcode_bold.ReplaceAllString(msg,"<b>$1</b>")
	msg = bbcode_italic.ReplaceAllString(msg,"<i>$1</i>")
	msg = bbcode_underline.ReplaceAllString(msg,"<u>$1</u>")
	msg = bbcode_strikethrough.ReplaceAllString(msg,"<s>$1</s>")
	msg = bbcode_url.ReplaceAllString(msg,"<a href=\"$1$2//$3\" rel=\"nofollow\">$1$2//$3</i>")
	msg = bbcode_url_label.ReplaceAllString(msg,"<a href=\"$1$2//$3\" rel=\"nofollow\">$4</i>")
	msg = bbcode_quotes.ReplaceAllString(msg,"<div class=\"postQuote\">$1</div>")
	return msg
}

// Only does the simple BBCode like [u], [b], [i] and [s]
func bbcode_simple_parse(data interface{}) interface{} {
	msg := data.(string)
	msgbytes := []byte(msg)
	has_u := false
	has_b := false
	has_i := false
	has_s := false
	for i := 0; (i + 2) < len(msgbytes); i++ {
		if msgbytes[i] == '[' && msgbytes[i + 2] == ']' {
			if msgbytes[i + 1] == 'b' {
				msgbytes[i] = '<'
				msgbytes[i + 2] = '>'
				has_b = true
			} else if msgbytes[i + 1] == 'i' {
				msgbytes[i] = '<'
				msgbytes[i + 2] = '>'
				has_i = true
			} else if msgbytes[i + 1] == 'u' {
				msgbytes[i] = '<'
				msgbytes[i + 2] = '>'
				has_u = true
			} else if msgbytes[i + 1] == 's' {
				msgbytes[i] = '<'
				msgbytes[i + 2] = '>'
				has_s = true
			}
			i += 2
		}
	}
	
	// There's an unclosed tag in there x.x
	if has_i || has_u || has_b || has_s {
		closer := []byte("</u></i></b></s>")
		msgbytes = append(msgbytes, closer...)
	}
	return string(msgbytes)
}

// Here for benchmarking purposes. Might add a plugin setting for disabling [code] as it has it's paws everywhere
func bbcode_parse_without_code(data interface{}) interface{} {
	msg := data.(string)
	msgbytes := []byte(msg)
	has_u := false
	has_b := false
	has_i := false
	has_s := false
	complex_bbc := false
	for i := 0; (i + 3) < len(msgbytes); i++ {
		if msgbytes[i] == '[' {
			if msgbytes[i + 2] != ']' {
				if msgbytes[i + 1] == '/' {
					if msgbytes[i + 3] == ']' {
						if msgbytes[i + 2] == 'b' {
							msgbytes[i] = '<'
							msgbytes[i + 3] = '>'
							has_b = false
						} else if msgbytes[i + 2] == 'i' {
							msgbytes[i] = '<'
							msgbytes[i + 3] = '>'
							has_i = false
						} else if msgbytes[i + 2] == 'u' {
							msgbytes[i] = '<'
							msgbytes[i + 3] = '>'
							has_u = false
						} else if msgbytes[i + 2] == 's' {
							msgbytes[i] = '<'
							msgbytes[i + 3] = '>'
							has_s = false
						}
						i += 3
					} else {
						complex_bbc = true
					}
				} else {
					complex_bbc = true
				}
			} else {
				if msgbytes[i + 1] == 'b' {
					msgbytes[i] = '<'
					msgbytes[i + 2] = '>'
					has_b = true
				} else if msgbytes[i + 1] == 'i' {
					msgbytes[i] = '<'
					msgbytes[i + 2] = '>'
					has_i = true
				} else if msgbytes[i + 1] == 'u' {
					msgbytes[i] = '<'
					msgbytes[i + 2] = '>'
					has_u = true
				} else if msgbytes[i + 1] == 's' {
					msgbytes[i] = '<'
					msgbytes[i + 2] = '>'
					has_s = true
				}
				i += 2
			}
		}
	}
	
	// There's an unclosed tag in there x.x
	if has_i || has_u || has_b || has_s {
		closer := []byte("</u></i></b></s>")
		msgbytes = append(msgbytes, closer...)
	}
	
	// Copy the new complex parser over once the rough edges have been smoothed over
	if complex_bbc {
		msg = bbcode_url.ReplaceAllString(msg,"<a href=\"$1$2//$3\" rel=\"nofollow\">$1$2//$3</i>")
		msg = bbcode_url_label.ReplaceAllString(msg,"<a href=\"$1$2//$3\" rel=\"nofollow\">$4</i>")
		msg = bbcode_quotes.ReplaceAllString(msg,"<div class=\"postQuote\">$1</div>")
	}
	return string(msgbytes)
}

// Does every type of BBCode
func bbcode_full_parse(data interface{}) interface{} {
	msg := data.(string)
	//fmt.Println("BBCode PrePre String:")
	//fmt.Println("`"+msg+"`")
	//fmt.Println("----")
	msgbytes := []byte(msg)
	has_u := false
	has_b := false
	has_i := false
	has_s := false
	has_c := false
	complex_bbc := false
	msgbytes = append(msgbytes,space_gap...)
	//fmt.Println("BBCode Simple Pre:")
	//fmt.Println("`"+string(msgbytes)+"`")
	//fmt.Println("----")
	for i := 0; i < len(msgbytes); i++ {
		if msgbytes[i] == '[' {
			if msgbytes[i + 2] != ']' {
				if msgbytes[i + 1] == '/' {
					if msgbytes[i + 3] == ']' {
						if !has_c {
							if msgbytes[i + 2] == 'b' {
								msgbytes[i] = '<'
								msgbytes[i + 3] = '>'
								has_b = false
							} else if msgbytes[i + 2] == 'i' {
								msgbytes[i] = '<'
								msgbytes[i + 3] = '>'
								has_i = false
							} else if msgbytes[i + 2] == 'u' {
								msgbytes[i] = '<'
								msgbytes[i + 3] = '>'
								has_u = false
							} else if msgbytes[i + 2] == 's' {
								msgbytes[i] = '<'
								msgbytes[i + 3] = '>'
								has_s = false
							}
							i += 3
						}
					} else {
						if msgbytes[i+2] == 'c' && msgbytes[i+3] == 'o' && msgbytes[i+4] == 'd' && msgbytes[i+5] == 'e' && msgbytes[i+6] == ']' {
							has_c = false
							i += 7
						}
						//if msglen >= (i+6) {
						//	fmt.Println("boo")
						//	fmt.Println(msglen)
						//	fmt.Println(i+6)
						//	fmt.Println(string(msgbytes[i:i+6]))
						//}
						complex_bbc = true
					}
				} else {
					if msgbytes[i+1] == 'c' && msgbytes[i+2] == 'o' && msgbytes[i+3] == 'd' && msgbytes[i+4] == 'e' && msgbytes[i+5] == ']' {
						has_c = true
						i += 6
					}
					//if msglen >= (i+5) {
					//	fmt.Println("boo2")
					//	fmt.Println(string(msgbytes[i:i+5]))
					//}
					complex_bbc = true
				}
			} else if !has_c {
				if msgbytes[i + 1] == 'b' {
					msgbytes[i] = '<'
					msgbytes[i + 2] = '>'
					has_b = true
				} else if msgbytes[i + 1] == 'i' {
					msgbytes[i] = '<'
					msgbytes[i + 2] = '>'
					has_i = true
				} else if msgbytes[i + 1] == 'u' {
					msgbytes[i] = '<'
					msgbytes[i + 2] = '>'
					has_u = true
				} else if msgbytes[i + 1] == 's' {
					msgbytes[i] = '<'
					msgbytes[i + 2] = '>'
					has_s = true
				}
				i += 2
			}
		}
	}
	
	// There's an unclosed tag in there x.x
	if has_i || has_u || has_b || has_s {
		closer := []byte("</u></i></b></s>")
		msgbytes = append(msgbytes, closer...)
	}
	
	if complex_bbc {
		i := 0
		var start int
		var lastTag int
		outbytes := make([]byte, len(msgbytes))
		//fmt.Println("BBCode Pre:")
		//fmt.Println("`"+string(msgbytes)+"`")
		//fmt.Println("----")
		for ; i < len(msgbytes); i++ {
			MainLoop:
			if msgbytes[i] == '[' {
				OuterComplex:
				if msgbytes[i + 1] == 'u' {
					if msgbytes[i+2] == 'r' && msgbytes[i+3] == 'l' && msgbytes[i+4] == ']' {
						start = i + 5
						outbytes = append(outbytes, msgbytes[lastTag:i]...)
						i = start
						i += partial_url_bytes_len(msgbytes[start:])
						
						//fmt.Println("Partial Bytes:")
						//fmt.Println(string(msgbytes[start:]))
						//fmt.Println("-----")
						if !bytes.Equal(msgbytes[i:i+6],[]byte("[/url]")) {
							//fmt.Println("Invalid Bytes:")
							//fmt.Println(string(msgbytes[i:i+6]))
							//fmt.Println("-----")
							outbytes = append(outbytes, invalid_url...)
							goto MainLoop
						}
						
						outbytes = append(outbytes, url_open...)
						outbytes = append(outbytes, msgbytes[start:i]...)
						outbytes = append(outbytes, url_open2...)
						outbytes = append(outbytes, msgbytes[start:i]...)
						outbytes = append(outbytes, url_close...)
						i += 6
						lastTag = i
					}
				} else if msgbytes[i + 1] == 'r' {
					if bytes.Equal(msgbytes[i+2:i+6],[]byte("and]")) {
						outbytes = append(outbytes, msgbytes[lastTag:i]...)
						start = i + 6
						i = start
						for ;; i++ {
							if msgbytes[i] == '[' {
								if !bytes.Equal(msgbytes[i+1:i+7],[]byte("/rand]")) {
									outbytes = append(outbytes, bbcode_missing_tag...)
									goto OuterComplex
								}
								break
							} else if (len(msgbytes) - 1) < (i + 10) {
								outbytes = append(outbytes, bbcode_missing_tag...)
									goto OuterComplex
							}
						}
						
						number, err := strconv.ParseInt(string(msgbytes[start:i]),10,64)
						if err != nil {
							outbytes = append(outbytes, bbcode_invalid_number...)
							goto MainLoop
						}
						
						dat := []byte(strconv.FormatInt((random.Int63n(number)),10))
						outbytes = append(outbytes, dat...)
						//log.Print("Outputted the random number")
						i += 7
						lastTag = i
					}
				}
			}
		}
		//fmt.Println(string(outbytes))
		if lastTag != i {
			outbytes = append(outbytes, msgbytes[lastTag:]...)
			//fmt.Println("Outbytes:")
			//fmt.Println(`"`+string(outbytes)+`"`)
			//fmt.Println("----")
		}
		
		if len(outbytes) != 0 {
			//fmt.Println("BBCode Post:")
			//fmt.Println(`"`+string(outbytes[0:len(outbytes) - 10])+`"`)
			//fmt.Println("----")
			msg = string(outbytes[0:len(outbytes) - 10])
		} else {
			msg = string(msgbytes[0:len(msgbytes) - 10])
		}
		
		//msg = bbcode_url.ReplaceAllString(msg,"<a href=\"$1$2//$3\" rel=\"nofollow\">$1$2//$3</i>")
		msg = bbcode_url_label.ReplaceAllString(msg,"<a href=\"$1$2//$3\" rel=\"nofollow\">$4</i>")
		msg = bbcode_quotes.ReplaceAllString(msg,"<div class=\"postQuote\">$1</div>")
		// Convert [code] into class="codequotes"
		//fmt.Println("guuuaaaa")
	} else {
		msg = string(msgbytes[0:len(msgbytes) - 10])
	}
	return msg
}
