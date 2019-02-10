package main

import (
	"bytes"
	"math/rand"
	"regexp"
	"strconv"
	"time"

	"github.com/Azareal/Gosora/common"
)

var bbcodeRandom *rand.Rand
var bbcodeInvalidNumber []byte
var bbcodeNoNegative []byte
var bbcodeMissingTag []byte

var bbcodeBold *regexp.Regexp
var bbcodeItalic *regexp.Regexp
var bbcodeUnderline *regexp.Regexp
var bbcodeStrikethrough *regexp.Regexp
var bbcodeURL *regexp.Regexp
var bbcodeURLLabel *regexp.Regexp
var bbcodeQuotes *regexp.Regexp
var bbcodeCode *regexp.Regexp

func init() {
	common.Plugins.Add(&common.Plugin{UName: "bbcode", Name: "BBCode", Author: "Azareal", URL: "https://github.com/Azareal", Init: initBbcode, Deactivate: deactivateBbcode})
}

func initBbcode(plugin *common.Plugin) error {
	plugin.AddHook("parse_assign", bbcodeFullParse)

	bbcodeInvalidNumber = []byte("<span style='color: red;'>[Invalid Number]</span>")
	bbcodeNoNegative = []byte("<span style='color: red;'>[No Negative Numbers]</span>")
	bbcodeMissingTag = []byte("<span style='color: red;'>[Missing Tag]</span>")

	bbcodeBold = regexp.MustCompile(`(?s)\[b\](.*)\[/b\]`)
	bbcodeItalic = regexp.MustCompile(`(?s)\[i\](.*)\[/i\]`)
	bbcodeUnderline = regexp.MustCompile(`(?s)\[u\](.*)\[/u\]`)
	bbcodeStrikethrough = regexp.MustCompile(`(?s)\[s\](.*)\[/s\]`)
	urlpattern := `(http|https|ftp|mailto*)(:??)\/\/([\.a-zA-Z\/]+)`
	bbcodeURL = regexp.MustCompile(`\[url\]` + urlpattern + `\[/url\]`)
	bbcodeURLLabel = regexp.MustCompile(`(?s)\[url=` + urlpattern + `\](.*)\[/url\]`)
	bbcodeQuotes = regexp.MustCompile(`\[quote\](.*)\[/quote\]`)
	bbcodeCode = regexp.MustCompile(`\[code\](.*)\[/code\]`)

	bbcodeRandom = rand.New(rand.NewSource(time.Now().UnixNano()))
	return nil
}

func deactivateBbcode(plugin *common.Plugin) {
	plugin.RemoveHook("parse_assign", bbcodeFullParse)
}

func bbcodeRegexParse(msg string) string {
	msg = bbcodeBold.ReplaceAllString(msg, "<b>$1</b>")
	msg = bbcodeItalic.ReplaceAllString(msg, "<i>$1</i>")
	msg = bbcodeUnderline.ReplaceAllString(msg, "<u>$1</u>")
	msg = bbcodeStrikethrough.ReplaceAllString(msg, "<s>$1</s>")
	msg = bbcodeURL.ReplaceAllString(msg, "<a href=''$1$2//$3' rel='nofollow'>$1$2//$3</i>")
	msg = bbcodeURLLabel.ReplaceAllString(msg, "<a href=''$1$2//$3' rel='nofollow'>$4</i>")
	msg = bbcodeQuotes.ReplaceAllString(msg, "<span class='postQuote'>$1</span>")
	//msg = bbcodeCode.ReplaceAllString(msg,"<span class='codequotes'>$1</span>")
	return msg
}

// Only does the simple BBCode like [u], [b], [i] and [s]
func bbcodeSimpleParse(msg string) string {
	var hasU, hasB, hasI, hasS bool
	msgbytes := []byte(msg)
	for i := 0; (i + 2) < len(msgbytes); i++ {
		if msgbytes[i] == '[' && msgbytes[i+2] == ']' {
			if msgbytes[i+1] == 'b' && !hasB {
				msgbytes[i] = '<'
				msgbytes[i+2] = '>'
				hasB = true
			} else if msgbytes[i+1] == 'i' && !hasI {
				msgbytes[i] = '<'
				msgbytes[i+2] = '>'
				hasI = true
			} else if msgbytes[i+1] == 'u' && !hasU {
				msgbytes[i] = '<'
				msgbytes[i+2] = '>'
				hasU = true
			} else if msgbytes[i+1] == 's' && !hasS {
				msgbytes[i] = '<'
				msgbytes[i+2] = '>'
				hasS = true
			}
			i += 2
		}
	}

	// There's an unclosed tag in there x.x
	if hasI || hasU || hasB || hasS {
		closeUnder := []byte("</u>")
		closeItalic := []byte("</i>")
		closeBold := []byte("</b>")
		closeStrike := []byte("</s>")
		if hasI {
			msgbytes = append(msgbytes, closeItalic...)
		}
		if hasU {
			msgbytes = append(msgbytes, closeUnder...)
		}
		if hasB {
			msgbytes = append(msgbytes, closeBold...)
		}
		if hasS {
			msgbytes = append(msgbytes, closeStrike...)
		}
	}
	return string(msgbytes)
}

// Here for benchmarking purposes. Might add a plugin setting for disabling [code] as it has it's paws everywhere
func bbcodeParseWithoutCode(msg string) string {
	var hasU, hasB, hasI, hasS bool
	var complexBbc bool
	msgbytes := []byte(msg)

	for i := 0; (i + 3) < len(msgbytes); i++ {
		if msgbytes[i] == '[' {
			if msgbytes[i+2] != ']' {
				if msgbytes[i+1] == '/' {
					if msgbytes[i+3] == ']' {
						if msgbytes[i+2] == 'b' {
							msgbytes[i] = '<'
							msgbytes[i+3] = '>'
							hasB = false
						} else if msgbytes[i+2] == 'i' {
							msgbytes[i] = '<'
							msgbytes[i+3] = '>'
							hasI = false
						} else if msgbytes[i+2] == 'u' {
							msgbytes[i] = '<'
							msgbytes[i+3] = '>'
							hasU = false
						} else if msgbytes[i+2] == 's' {
							msgbytes[i] = '<'
							msgbytes[i+3] = '>'
							hasS = false
						}
						i += 3
					} else {
						complexBbc = true
					}
				} else {
					complexBbc = true
				}
			} else {
				if msgbytes[i+1] == 'b' && !hasB {
					msgbytes[i] = '<'
					msgbytes[i+2] = '>'
					hasB = true
				} else if msgbytes[i+1] == 'i' && !hasI {
					msgbytes[i] = '<'
					msgbytes[i+2] = '>'
					hasI = true
				} else if msgbytes[i+1] == 'u' && !hasU {
					msgbytes[i] = '<'
					msgbytes[i+2] = '>'
					hasU = true
				} else if msgbytes[i+1] == 's' && !hasS {
					msgbytes[i] = '<'
					msgbytes[i+2] = '>'
					hasS = true
				}
				i += 2
			}
		}
	}

	// There's an unclosed tag in there x.x
	if hasI || hasU || hasB || hasS {
		closeUnder := []byte("</u>")
		closeItalic := []byte("</i>")
		closeBold := []byte("</b>")
		closeStrike := []byte("</s>")
		if hasI {
			msgbytes = append(bytes.TrimSpace(msgbytes), closeItalic...)
		}
		if hasU {
			msgbytes = append(bytes.TrimSpace(msgbytes), closeUnder...)
		}
		if hasB {
			msgbytes = append(bytes.TrimSpace(msgbytes), closeBold...)
		}
		if hasS {
			msgbytes = append(bytes.TrimSpace(msgbytes), closeStrike...)
		}
	}

	// Copy the new complex parser over once the rough edges have been smoothed over
	if complexBbc {
		msg = string(msgbytes)
		msg = bbcodeURL.ReplaceAllString(msg, "<a href='$1$2//$3' rel='nofollow'>$1$2//$3</i>")
		msg = bbcodeURLLabel.ReplaceAllString(msg, "<a href='$1$2//$3' rel='nofollow'>$4</i>")
		msg = bbcodeQuotes.ReplaceAllString(msg, "<span class='postQuote'>$1</span>")
		return bbcodeCode.ReplaceAllString(msg, "<span class='codequotes'>$1</span>")
	}
	return string(msgbytes)
}

// Does every type of BBCode
func bbcodeFullParse(msg string) string {
	var hasU, hasB, hasI, hasS, hasC bool
	var complexBbc bool

	msgbytes := []byte(msg)
	msgbytes = append(msgbytes, common.SpaceGap...)
	for i := 0; i < len(msgbytes); i++ {
		if msgbytes[i] == '[' {
			if msgbytes[i+2] != ']' {
				if msgbytes[i+1] == '/' {
					if msgbytes[i+3] == ']' {
						if !hasC {
							if msgbytes[i+2] == 'b' {
								msgbytes[i] = '<'
								msgbytes[i+3] = '>'
								hasB = false
							} else if msgbytes[i+2] == 'i' {
								msgbytes[i] = '<'
								msgbytes[i+3] = '>'
								hasI = false
							} else if msgbytes[i+2] == 'u' {
								msgbytes[i] = '<'
								msgbytes[i+3] = '>'
								hasU = false
							} else if msgbytes[i+2] == 's' {
								msgbytes[i] = '<'
								msgbytes[i+3] = '>'
								hasS = false
							}
							i += 3
						}
					} else {
						if msgbytes[i+2] == 'c' && msgbytes[i+3] == 'o' && msgbytes[i+4] == 'd' && msgbytes[i+5] == 'e' && msgbytes[i+6] == ']' {
							hasC = false
							i += 7
						}
						complexBbc = true
					}
				} else {
					if msgbytes[i+1] == 'c' && msgbytes[i+2] == 'o' && msgbytes[i+3] == 'd' && msgbytes[i+4] == 'e' && msgbytes[i+5] == ']' {
						hasC = true
						i += 6
					}
					complexBbc = true
				}
			} else if !hasC {
				if msgbytes[i+1] == 'b' && !hasB {
					msgbytes[i] = '<'
					msgbytes[i+2] = '>'
					hasB = true
				} else if msgbytes[i+1] == 'i' && !hasI {
					msgbytes[i] = '<'
					msgbytes[i+2] = '>'
					hasI = true
				} else if msgbytes[i+1] == 'u' && !hasU {
					msgbytes[i] = '<'
					msgbytes[i+2] = '>'
					hasU = true
				} else if msgbytes[i+1] == 's' && !hasS {
					msgbytes[i] = '<'
					msgbytes[i+2] = '>'
					hasS = true
				}
				i += 2
			}
		}
	}

	// There's an unclosed tag in there somewhere x.x
	if hasI || hasU || hasB || hasS {
		closeUnder := []byte("</u>")
		closeItalic := []byte("</i>")
		closeBold := []byte("</b>")
		closeStrike := []byte("</s>")
		if hasI {
			msgbytes = append(bytes.TrimSpace(msgbytes), closeItalic...)
		}
		if hasU {
			msgbytes = append(bytes.TrimSpace(msgbytes), closeUnder...)
		}
		if hasB {
			msgbytes = append(bytes.TrimSpace(msgbytes), closeBold...)
		}
		if hasS {
			msgbytes = append(bytes.TrimSpace(msgbytes), closeStrike...)
		}
		msgbytes = append(msgbytes, common.SpaceGap...)
	}

	if complexBbc {
		i := 0
		var start, lastTag int
		var outbytes []byte
		for ; i < len(msgbytes); i++ {
			if msgbytes[i] == '[' {
				if msgbytes[i+1] == 'u' {
					if msgbytes[i+2] == 'r' && msgbytes[i+3] == 'l' && msgbytes[i+4] == ']' {
						i, start, lastTag, outbytes = bbcodeParseURL(i, start, lastTag, msgbytes, outbytes)
						continue
					}
				} else if msgbytes[i+1] == 'r' {
					if bytes.Equal(msgbytes[i+2:i+6], []byte("and]")) {
						i, start, lastTag, outbytes = bbcodeParseRand(i, start, lastTag, msgbytes, outbytes)
					}
				}
			}
		}
		if lastTag != i {
			outbytes = append(outbytes, msgbytes[lastTag:]...)
		}

		if len(outbytes) != 0 {
			msg = string(outbytes[0 : len(outbytes)-10])
		} else {
			msg = string(msgbytes[0 : len(msgbytes)-10])
		}

		//msg = bbcode_url.ReplaceAllString(msg,"<a href=\"$1$2//$3\" rel=\"nofollow\">$1$2//$3</i>")
		msg = bbcodeURLLabel.ReplaceAllString(msg, "<a href='$1$2//$3' rel='nofollow'>$4</i>")
		msg = bbcodeQuotes.ReplaceAllString(msg, "<span class='postQuote'>$1</span>")
		msg = bbcodeCode.ReplaceAllString(msg, "<span class='codequotes'>$1</span>")
	} else {
		msg = string(msgbytes[0 : len(msgbytes)-10])
	}

	return msg
}

// TODO: Strip the containing [url] so the media parser can work it's magic instead? Or do we want to allow something like [url=]label[/url] here?
func bbcodeParseURL(i int, start int, lastTag int, msgbytes []byte, outbytes []byte) (int, int, int, []byte) {
	start = i + 5
	outbytes = append(outbytes, msgbytes[lastTag:i]...)
	i = start
	i += common.PartialURLStringLen(string(msgbytes[start:]))
	if !bytes.Equal(msgbytes[i:i+6], []byte("[/url]")) {
		outbytes = append(outbytes, common.InvalidURL...)
		return i, start, lastTag, outbytes
	}

	outbytes = append(outbytes, common.URLOpen...)
	outbytes = append(outbytes, msgbytes[start:i]...)
	outbytes = append(outbytes, common.URLOpen2...)
	outbytes = append(outbytes, msgbytes[start:i]...)
	outbytes = append(outbytes, common.URLClose...)
	i += 6
	lastTag = i

	return i, start, lastTag, outbytes
}

func bbcodeParseRand(i int, start int, lastTag int, msgbytes []byte, outbytes []byte) (int, int, int, []byte) {
	outbytes = append(outbytes, msgbytes[lastTag:i]...)
	start = i + 6
	i = start
	for ; ; i++ {
		if msgbytes[i] == '[' {
			if !bytes.Equal(msgbytes[i+1:i+7], []byte("/rand]")) {
				outbytes = append(outbytes, bbcodeMissingTag...)
				return i, start, lastTag, outbytes
			}
			break
		} else if (len(msgbytes) - 1) < (i + 10) {
			outbytes = append(outbytes, bbcodeMissingTag...)
			return i, start, lastTag, outbytes
		}
	}

	number, err := strconv.ParseInt(string(msgbytes[start:i]), 10, 64)
	if err != nil {
		outbytes = append(outbytes, bbcodeInvalidNumber...)
		return i, start, lastTag, outbytes
	}

	// TODO: Add support for negative numbers?
	if number < 0 {
		outbytes = append(outbytes, bbcodeNoNegative...)
		return i, start, lastTag, outbytes
	}

	var dat []byte
	if number == 0 {
		dat = []byte("0")
	} else {
		dat = []byte(strconv.FormatInt((bbcodeRandom.Int63n(number)), 10))
	}

	outbytes = append(outbytes, dat...)
	i += 7
	lastTag = i
	return i, start, lastTag, outbytes
}
