package main
//import "strings"
import "regexp"

var bbcode_bold *regexp.Regexp
var bbcode_italic *regexp.Regexp
var bbcode_underline *regexp.Regexp
var bbcode_strikethrough *regexp.Regexp
var bbcode_url *regexp.Regexp
var bbcode_url_label *regexp.Regexp

func init() {
	plugins["bbcode"] = NewPlugin("bbcode","BBCode","Azareal","http://github.com/Azareal","","","",init_bbcode,nil,deactivate_bbcode)
}

func init_bbcode() {
	plugins["bbcode"].AddHook("parse_assign", bbcode_parse_without_code)
	//plugins["bbcode"].AddHook("parse_assign", bbcode_full_parse)
	
	bbcode_bold = regexp.MustCompile(`(?s)\[b\](.*)\[/b\]`)
	bbcode_italic = regexp.MustCompile(`(?s)\[i\](.*)\[/i\]`)
	bbcode_underline = regexp.MustCompile(`(?s)\[u\](.*)\[/u\]`)
	bbcode_strikethrough = regexp.MustCompile(`(?s)\[s\](.*)\[/s\]`)
	urlpattern := `(http|https|ftp|mailto*)(:??)\/\/([\.a-zA-Z\/]+)`
	bbcode_url = regexp.MustCompile(`\[url\]` + urlpattern + `\[/url\]`)
	bbcode_url_label = regexp.MustCompile(`(?s)\[url=` + urlpattern + `\](.*)\[/url\]`)
}

func deactivate_bbcode() {
	plugins["bbcode"].RemoveHook("parse_assign", bbcode_parse_without_code)
	//plugins["bbcode"].RemoveHook("parse_assign", bbcode_full_parse)
}

func bbcode_regex_parse(data interface{}) interface{} {
	msg := data.(string)
	msg = bbcode_bold.ReplaceAllString(msg,"<b>$1</b>")
	msg = bbcode_italic.ReplaceAllString(msg,"<i>$1</i>")
	msg = bbcode_underline.ReplaceAllString(msg,"<u>$1</u>")
	msg = bbcode_strikethrough.ReplaceAllString(msg,"<s>$1</s>")
	msg = bbcode_url.ReplaceAllString(msg,"<a href=\"$1$2//$3\" rel=\"nofollow\">$1$2//$3</i>")
	msg = bbcode_url_label.ReplaceAllString(msg,"<a href=\"$1$2//$3\" rel=\"nofollow\">$4</i>")
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
	for i := 0; i < len(msgbytes); i++ {
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
	for i := 0; i < len(msgbytes); i++ {
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
	msg = string(msgbytes)
	
	if complex_bbc {
		msg = bbcode_url.ReplaceAllString(msg,"<a href=\"$1$2//$3\" rel=\"nofollow\">$1$2//$3</i>")
		msg = bbcode_url_label.ReplaceAllString(msg,"<a href=\"$1$2//$3\" rel=\"nofollow\">$4</i>")
	}
	return msg
}

// Does every type of BBCode
func bbcode_full_parse(data interface{}) interface{} {
	msg := data.(string)
	msgbytes := []byte(msg)
	has_u := false
	has_b := false
	has_i := false
	has_s := false
	has_c := false
	complex_bbc := false
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
						if msgbytes[i + 2] == 'c' && msgbytes[i + 3] == 'o' && msgbytes[i + 4] == 'd' && msgbytes[i + 5] == 'e' {
							has_c = false
							i += 6
						}
						complex_bbc = true
					}
				} else {
					if msgbytes[i + 1] == 'c' && msgbytes[i + 2] == 'o' && msgbytes[i + 3] == 'd' && msgbytes[i + 4] == 'e' {
						has_c = true
						i += 5
					}
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
	msg = string(msgbytes)
	
	if complex_bbc {
		msg = bbcode_url.ReplaceAllString(msg,"<a href=\"$1$2//$3\" rel=\"nofollow\">$1$2//$3</i>")
		msg = bbcode_url_label.ReplaceAllString(msg,"<a href=\"$1$2//$3\" rel=\"nofollow\">$4</i>")
		//msg = strings.Replace(msg,"[code]","",-1)
		//msg = strings.Replace(msg,"[/code]","",-1)
	}
	return msg
}