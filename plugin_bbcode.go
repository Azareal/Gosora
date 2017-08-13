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
var bbcode_no_negative []byte
var bbcode_missing_tag []byte

var bbcode_bold *regexp.Regexp
var bbcode_italic *regexp.Regexp
var bbcode_underline *regexp.Regexp
var bbcode_strikethrough *regexp.Regexp
var bbcode_url *regexp.Regexp
var bbcode_url_label *regexp.Regexp
var bbcode_quotes *regexp.Regexp
var bbcode_code *regexp.Regexp

func init() {
	plugins["bbcode"] = NewPlugin("bbcode","BBCode","Azareal","http://github.com/Azareal","","","",init_bbcode,nil,deactivate_bbcode,nil,nil)
}

func init_bbcode() error {
	//plugins["bbcode"].AddHook("parse_assign", bbcode_parse_without_code)
	plugins["bbcode"].AddHook("parse_assign", bbcode_full_parse)

	bbcode_invalid_number = []byte("<span style='color: red;'>[Invalid Number]</span>")
	bbcode_no_negative = []byte("<span style='color: red;'>[No Negative Numbers]</span>")
	bbcode_missing_tag = []byte("<span style='color: red;'>[Missing Tag]</span>")

	bbcode_bold = regexp.MustCompile(`(?s)\[b\](.*)\[/b\]`)
	bbcode_italic = regexp.MustCompile(`(?s)\[i\](.*)\[/i\]`)
	bbcode_underline = regexp.MustCompile(`(?s)\[u\](.*)\[/u\]`)
	bbcode_strikethrough = regexp.MustCompile(`(?s)\[s\](.*)\[/s\]`)
	urlpattern := `(http|https|ftp|mailto*)(:??)\/\/([\.a-zA-Z\/]+)`
	bbcode_url = regexp.MustCompile(`\[url\]` + urlpattern + `\[/url\]`)
	bbcode_url_label = regexp.MustCompile(`(?s)\[url=` + urlpattern + `\](.*)\[/url\]`)
	bbcode_quotes = regexp.MustCompile(`\[quote\](.*)\[/quote\]`)
	bbcode_code = regexp.MustCompile(`\[code\](.*)\[/code\]`)

	random = rand.New(rand.NewSource(time.Now().UnixNano()))
	return nil
}

func deactivate_bbcode() {
	//plugins["bbcode"].RemoveHook("parse_assign", bbcode_parse_without_code)
	plugins["bbcode"].RemoveHook("parse_assign", bbcode_full_parse)
}

func bbcode_regex_parse(msg string) string {
	msg = bbcode_bold.ReplaceAllString(msg,"<b>$1</b>")
	msg = bbcode_italic.ReplaceAllString(msg,"<i>$1</i>")
	msg = bbcode_underline.ReplaceAllString(msg,"<u>$1</u>")
	msg = bbcode_strikethrough.ReplaceAllString(msg,"<s>$1</s>")
	msg = bbcode_url.ReplaceAllString(msg,"<a href=''$1$2//$3' rel='nofollow'>$1$2//$3</i>")
	msg = bbcode_url_label.ReplaceAllString(msg,"<a href=''$1$2//$3' rel='nofollow'>$4</i>")
	msg = bbcode_quotes.ReplaceAllString(msg,"<span class='postQuote'>$1</span>")
	//msg = bbcode_code.ReplaceAllString(msg,"<span class='codequotes'>$1</span>")
	return msg
}

// Only does the simple BBCode like [u], [b], [i] and [s]
func bbcode_simple_parse(msg string) string {
	var has_u, has_b, has_i, has_s bool
	msgbytes := []byte(msg)
	for i := 0; (i + 2) < len(msgbytes); i++ {
		if msgbytes[i] == '[' && msgbytes[i + 2] == ']' {
			if msgbytes[i + 1] == 'b' && !has_b {
				msgbytes[i] = '<'
				msgbytes[i + 2] = '>'
				has_b = true
			} else if msgbytes[i + 1] == 'i' && !has_i {
				msgbytes[i] = '<'
				msgbytes[i + 2] = '>'
				has_i = true
			} else if msgbytes[i + 1] == 'u' && !has_u {
				msgbytes[i] = '<'
				msgbytes[i + 2] = '>'
				has_u = true
			} else if msgbytes[i + 1] == 's' && !has_s {
				msgbytes[i] = '<'
				msgbytes[i + 2] = '>'
				has_s = true
			}
			i += 2
		}
	}

	// There's an unclosed tag in there x.x
	if has_i || has_u || has_b || has_s {
		close_under := []byte("</u>")
		close_italic := []byte("</i>")
		close_bold := []byte("</b>")
		close_strike := []byte("</s>")
		if has_i {
			msgbytes = append(msgbytes, close_italic...)
		}
		if has_u {
			msgbytes = append(msgbytes, close_under...)
		}
		if has_b {
			msgbytes = append(msgbytes, close_bold...)
		}
		if has_s {
			msgbytes = append(msgbytes, close_strike...)
		}
	}
	return string(msgbytes)
}

// Here for benchmarking purposes. Might add a plugin setting for disabling [code] as it has it's paws everywhere
func bbcode_parse_without_code(msg string) string {
	var has_u, has_b, has_i, has_s bool
	var complex_bbc bool
	msgbytes := []byte(msg)

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
				if msgbytes[i + 1] == 'b' && !has_b {
					msgbytes[i] = '<'
					msgbytes[i + 2] = '>'
					has_b = true
				} else if msgbytes[i + 1] == 'i' && !has_i {
					msgbytes[i] = '<'
					msgbytes[i + 2] = '>'
					has_i = true
				} else if msgbytes[i + 1] == 'u' && !has_u {
					msgbytes[i] = '<'
					msgbytes[i + 2] = '>'
					has_u = true
				} else if msgbytes[i + 1] == 's' && !has_s {
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
		close_under := []byte("</u>")
		close_italic := []byte("</i>")
		close_bold := []byte("</b>")
		close_strike := []byte("</s>")
		if has_i {
			msgbytes = append(bytes.TrimSpace(msgbytes), close_italic...)
		}
		if has_u {
			msgbytes = append(bytes.TrimSpace(msgbytes), close_under...)
		}
		if has_b {
			msgbytes = append(bytes.TrimSpace(msgbytes), close_bold...)
		}
		if has_s {
			msgbytes = append(bytes.TrimSpace(msgbytes), close_strike...)
		}
	}

	// Copy the new complex parser over once the rough edges have been smoothed over
	if complex_bbc {
		msg = bbcode_url.ReplaceAllString(msg,"<a href='$1$2//$3' rel='nofollow'>$1$2//$3</i>")
		msg = bbcode_url_label.ReplaceAllString(msg,"<a href='$1$2//$3' rel='nofollow'>$4</i>")
		msg = bbcode_quotes.ReplaceAllString(msg,"<span class='postQuote'>$1</span>")
		msg = bbcode_code.ReplaceAllString(msg,"<span class='codequotes'>$1</span>")
	}

	return string(msgbytes)
}

// Does every type of BBCode
func bbcode_full_parse(msg string) string {
	var has_u, has_b, has_i, has_s, has_c bool
	var complex_bbc bool

	msgbytes := []byte(msg)
	msgbytes = append(msgbytes,space_gap...)
	//log.Print("BBCode Simple Pre:","`"+string(msgbytes)+"`")
	//log.Print("----")

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
						//	log.Print("boo")
						//	log.Print(msglen)
						//	log.Print(i+6)
						//	log.Print(string(msgbytes[i:i+6]))
						//}
						complex_bbc = true
					}
				} else {
					if msgbytes[i+1] == 'c' && msgbytes[i+2] == 'o' && msgbytes[i+3] == 'd' && msgbytes[i+4] == 'e' && msgbytes[i+5] == ']' {
						has_c = true
						i += 6
					}
					//if msglen >= (i+5) {
					//	log.Print("boo2")
					//	log.Print(string(msgbytes[i:i+5]))
					//}
					complex_bbc = true
				}
			} else if !has_c {
				if msgbytes[i + 1] == 'b' && !has_b {
					msgbytes[i] = '<'
					msgbytes[i + 2] = '>'
					has_b = true
				} else if msgbytes[i + 1] == 'i' && !has_i {
					msgbytes[i] = '<'
					msgbytes[i + 2] = '>'
					has_i = true
				} else if msgbytes[i + 1] == 'u' && !has_u {
					msgbytes[i] = '<'
					msgbytes[i + 2] = '>'
					has_u = true
				} else if msgbytes[i + 1] == 's' && !has_s {
					msgbytes[i] = '<'
					msgbytes[i + 2] = '>'
					has_s = true
				}
				i += 2
			}
		}
	}

	// There's an unclosed tag in there somewhere x.x
	if has_i || has_u || has_b || has_s {
		close_under := []byte("</u>")
		close_italic := []byte("</i>")
		close_bold := []byte("</b>")
		close_strike := []byte("</s>")
		if has_i {
			msgbytes = append(bytes.TrimSpace(msgbytes), close_italic...)
		}
		if has_u {
			msgbytes = append(bytes.TrimSpace(msgbytes), close_under...)
		}
		if has_b {
			msgbytes = append(bytes.TrimSpace(msgbytes), close_bold...)
		}
		if has_s {
			msgbytes = append(bytes.TrimSpace(msgbytes), close_strike...)
		}
		msgbytes = append(msgbytes,space_gap...)
	}

	if complex_bbc {
		i := 0
		var start, lastTag int
		var outbytes []byte
		//log.Print("BBCode Pre:","`"+string(msgbytes)+"`")
		//log.Print("----")
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
						//log.Print("Partial Bytes:",string(msgbytes[start:]))
						//log.Print("-----")
						if !bytes.Equal(msgbytes[i:i+6],[]byte("[/url]")) {
							//log.Print("Invalid Bytes:",string(msgbytes[i:i+6]))
							//log.Print("-----")
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

						// TO-DO: Add support for negative numbers?
						if number < 0 {
							outbytes = append(outbytes, bbcode_no_negative...)
							goto MainLoop
						}

						var dat []byte
						if number == 0 {
							dat = []byte("0")
						} else {
							dat = []byte(strconv.FormatInt((random.Int63n(number)),10))
						}

						outbytes = append(outbytes, dat...)
						//log.Print("Outputted the random number")
						i += 7
						lastTag = i
					}
				}
			}
		}
		//log.Print(string(outbytes))
		if lastTag != i {
			outbytes = append(outbytes, msgbytes[lastTag:]...)
			//log.Print("Outbytes:",`"`+string(outbytes)+`"`)
			//log.Print("----")
		}

		if len(outbytes) != 0 {
			//log.Print("BBCode Post:",`"`+string(outbytes[0:len(outbytes) - 10])+`"`)
			//log.Print("----")
			msg = string(outbytes[0:len(outbytes) - 10])
		} else {
			msg = string(msgbytes[0:len(msgbytes) - 10])
		}

		//msg = bbcode_url.ReplaceAllString(msg,"<a href=\"$1$2//$3\" rel=\"nofollow\">$1$2//$3</i>")
		msg = bbcode_url_label.ReplaceAllString(msg,"<a href='$1$2//$3' rel='nofollow'>$4</i>")
		msg = bbcode_quotes.ReplaceAllString(msg,"<span class='postQuote'>$1</span>")
		msg = bbcode_code.ReplaceAllString(msg,"<span class='codequotes'>$1</span>")
	} else {
		msg = string(msgbytes[0:len(msgbytes) - 10])
	}

	return msg
}
