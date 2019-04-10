package common

import (
	"bytes"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// TODO: Somehow localise these?
var SpaceGap = []byte("          ")
var httpProtBytes = []byte("http://")
var InvalidURL = []byte("<red>[Invalid URL]</red>")
var InvalidTopic = []byte("<red>[Invalid Topic]</red>")
var InvalidProfile = []byte("<red>[Invalid Profile]</red>")
var InvalidForum = []byte("<red>[Invalid Forum]</red>")
var unknownMedia = []byte("<red>[Unknown Media]</red>")
var URLOpen = []byte("<a href='")
var URLOpen2 = []byte("'>")
var bytesSinglequote = []byte("'")
var bytesGreaterthan = []byte(">")
var urlMention = []byte(" class='mention'")
var URLClose = []byte("</a>")
var imageOpen = []byte("<a href=\"")
var imageOpen2 = []byte("\"><img src='")
var imageClose = []byte("' class='postImage' /></a>")
var urlPattern = `(?s)([ {1}])((http|https|ftp|mailto)*)(:{??)\/\/([\.a-zA-Z\/]+)([ {1}])`
var urlReg *regexp.Regexp

func init() {
	urlReg = regexp.MustCompile(urlPattern)
}

// TODO: Write a test for this
func shortcodeToUnicode(msg string) string {
	//re := regexp.MustCompile(":(.):")
	msg = strings.Replace(msg, ":grinning:", "😀", -1)
	msg = strings.Replace(msg, ":grin:", "😁", -1)
	msg = strings.Replace(msg, ":joy:", "😂", -1)
	msg = strings.Replace(msg, ":rofl:", "🤣", -1)
	msg = strings.Replace(msg, ":smiley:", "😃", -1)
	msg = strings.Replace(msg, ":smile:", "😄", -1)
	msg = strings.Replace(msg, ":sweat_smile:", "😅", -1)
	msg = strings.Replace(msg, ":laughing:", "😆", -1)
	msg = strings.Replace(msg, ":satisfied:", "😆", -1)
	msg = strings.Replace(msg, ":wink:", "😉", -1)
	msg = strings.Replace(msg, ":blush:", "😊", -1)
	msg = strings.Replace(msg, ":yum:", "😋", -1)
	msg = strings.Replace(msg, ":sunglasses:", "😎", -1)
	msg = strings.Replace(msg, ":heart_eyes:", "😍", -1)
	msg = strings.Replace(msg, ":kissing_heart:", "😘", -1)
	msg = strings.Replace(msg, ":kissing:", "😗", -1)
	msg = strings.Replace(msg, ":kissing_smiling_eyes:", "😙", -1)
	msg = strings.Replace(msg, ":kissing_closed_eyes:", "😚", -1)
	msg = strings.Replace(msg, ":relaxed:", "☺️", -1)
	msg = strings.Replace(msg, ":slight_smile:", "🙂", -1)
	msg = strings.Replace(msg, ":hugging:", "🤗", -1)
	msg = strings.Replace(msg, ":thinking:", "🤔", -1)
	msg = strings.Replace(msg, ":neutral_face:", "😐", -1)
	msg = strings.Replace(msg, ":expressionless:", "😑", -1)
	msg = strings.Replace(msg, ":no_mouth:", "😶", -1)
	msg = strings.Replace(msg, ":rolling_eyes:", "🙄", -1)
	msg = strings.Replace(msg, ":smirk:", "😏", -1)
	msg = strings.Replace(msg, ":persevere:", "😣", -1)
	msg = strings.Replace(msg, ":disappointed_relieved:", "😥", -1)
	msg = strings.Replace(msg, ":open_mouth:", "😮", -1)
	msg = strings.Replace(msg, ":zipper_mouth:", "🤐", -1)
	msg = strings.Replace(msg, ":hushed:", "😯", -1)
	msg = strings.Replace(msg, ":sleepy:", "😪", -1)
	msg = strings.Replace(msg, ":tired_face:", "😫", -1)
	msg = strings.Replace(msg, ":sleeping:", "😴", -1)
	msg = strings.Replace(msg, ":relieved:", "😌", -1)
	msg = strings.Replace(msg, ":nerd:", "🤓", -1)
	msg = strings.Replace(msg, ":stuck_out_tongue:", "😛", -1)
	msg = strings.Replace(msg, ":worried:", "😟", -1)
	msg = strings.Replace(msg, ":drooling_face:", "🤤", -1)
	msg = strings.Replace(msg, ":disappointed:", "😞", -1)
	msg = strings.Replace(msg, ":astonished:", "😲", -1)
	msg = strings.Replace(msg, ":slight_frown:", "🙁", -1)
	msg = strings.Replace(msg, ":skull_crossbones:", "☠️", -1)
	msg = strings.Replace(msg, ":skull:", "💀", -1)
	msg = strings.Replace(msg, ":point_up:", "☝️", -1)
	msg = strings.Replace(msg, ":v:", "✌️️", -1)
	msg = strings.Replace(msg, ":writing_hand:", "✍️", -1)
	msg = strings.Replace(msg, ":heart:", "❤️️", -1)
	msg = strings.Replace(msg, ":heart_exclamation:", "❣️", -1)
	msg = strings.Replace(msg, ":hotsprings:", "♨️", -1)
	msg = strings.Replace(msg, ":airplane:", "✈️️", -1)
	msg = strings.Replace(msg, ":hourglass:", "⌛", -1)
	msg = strings.Replace(msg, ":watch:", "⌚", -1)
	msg = strings.Replace(msg, ":comet:", "☄️", -1)
	msg = strings.Replace(msg, ":snowflake:", "❄️", -1)
	msg = strings.Replace(msg, ":cloud:", "☁️", -1)
	msg = strings.Replace(msg, ":sunny:", "☀️", -1)
	msg = strings.Replace(msg, ":spades:", "♠️", -1)
	msg = strings.Replace(msg, ":hearts:", "♥️️", -1)
	msg = strings.Replace(msg, ":diamonds:", "♦️", -1)
	msg = strings.Replace(msg, ":clubs:", "♣️", -1)
	msg = strings.Replace(msg, ":phone:", "☎️", -1)
	msg = strings.Replace(msg, ":telephone:", "☎️", -1)
	msg = strings.Replace(msg, ":biohazard:", "☣️", -1)
	msg = strings.Replace(msg, ":radioactive:", "☢️", -1)
	msg = strings.Replace(msg, ":scissors:", "✂️", -1)
	msg = strings.Replace(msg, ":arrow_upper_right:", "↗️", -1)
	msg = strings.Replace(msg, ":arrow_right:", "➡️", -1)
	msg = strings.Replace(msg, ":arrow_lower_right:", "↘️", -1)
	msg = strings.Replace(msg, ":arrow_lower_left:", "↙️", -1)
	msg = strings.Replace(msg, ":arrow_upper_left:", "↖️", -1)
	msg = strings.Replace(msg, ":arrow_up_down:", "↕️", -1)
	msg = strings.Replace(msg, ":left_right_arrow:", "↔️", -1)
	msg = strings.Replace(msg, ":leftwards_arrow_with_hook:", "↩️", -1)
	msg = strings.Replace(msg, ":arrow_right_hook:", "↪️", -1)
	msg = strings.Replace(msg, ":arrow_forward:", "▶️", -1)
	msg = strings.Replace(msg, ":arrow_backward:", "◀️", -1)
	msg = strings.Replace(msg, ":female:", "♀️", -1)
	msg = strings.Replace(msg, ":male:", "♂️", -1)
	msg = strings.Replace(msg, ":ballot_box_with_check:", "☑️", -1)
	msg = strings.Replace(msg, ":heavy_check_mark:", "✔️️", -1)
	msg = strings.Replace(msg, ":heavy_multiplication_x:", "✖️", -1)
	msg = strings.Replace(msg, ":pisces:", "♓", -1)
	msg = strings.Replace(msg, ":aquarius:", "♒", -1)
	msg = strings.Replace(msg, ":capricorn:", "♑", -1)
	msg = strings.Replace(msg, ":sagittarius:", "♐", -1)
	msg = strings.Replace(msg, ":scorpius:", "♏", -1)
	msg = strings.Replace(msg, ":libra:", "♎", -1)
	msg = strings.Replace(msg, ":virgo:", "♍", -1)
	msg = strings.Replace(msg, ":leo:", "♌", -1)
	msg = strings.Replace(msg, ":cancer:", "♋", -1)
	msg = strings.Replace(msg, ":gemini:", "♊", -1)
	msg = strings.Replace(msg, ":taurus:", "♉", -1)
	msg = strings.Replace(msg, ":aries:", "♈", -1)
	msg = strings.Replace(msg, ":peace:", "☮️", -1)
	msg = strings.Replace(msg, ":eight_spoked_asterisk:", "✳️", -1)
	msg = strings.Replace(msg, ":eight_pointed_black_star:", "✴️", -1)
	msg = strings.Replace(msg, ":snowman2:", "☃️", -1)
	msg = strings.Replace(msg, ":umbrella2:", "☂️", -1)
	msg = strings.Replace(msg, ":pencil2:", "✏️", -1)
	msg = strings.Replace(msg, ":black_nib:", "✒️", -1)
	msg = strings.Replace(msg, ":email:", "✉️", -1)
	msg = strings.Replace(msg, ":envelope:", "✉️", -1)
	msg = strings.Replace(msg, ":keyboard:", "⌨️", -1)
	msg = strings.Replace(msg, ":white_small_square:", "▫️", -1)
	msg = strings.Replace(msg, ":black_small_square:", "▪️", -1)
	msg = strings.Replace(msg, ":secret:", "㊙️", -1)
	msg = strings.Replace(msg, ":congratulations:", "㊗️", -1)
	msg = strings.Replace(msg, ":m:", "Ⓜ️", -1)
	msg = strings.Replace(msg, ":tm:", "™️️", -1)
	msg = strings.Replace(msg, ":registered:", "®️", -1)
	msg = strings.Replace(msg, ":copyright:", "©️", -1)
	msg = strings.Replace(msg, ":wavy_dash:", "〰️", -1)
	msg = strings.Replace(msg, ":bangbang:", "‼️", -1)
	msg = strings.Replace(msg, ":sparkle:", "❇️", -1)
	msg = strings.Replace(msg, ":star_of_david:", "✡️", -1)
	msg = strings.Replace(msg, ":wheel_of_dharma:", "☸️", -1)
	msg = strings.Replace(msg, ":yin_yang:", "☯️", -1)
	msg = strings.Replace(msg, ":cross:", "✝️", -1)
	msg = strings.Replace(msg, ":orthodox_cross:", "☦️", -1)
	msg = strings.Replace(msg, ":star_and_crescent:", "☪️", -1)
	msg = strings.Replace(msg, ":frowning2:", "☹️", -1)
	msg = strings.Replace(msg, ":information_source:", "ℹ️", -1)
	msg = strings.Replace(msg, ":interrobang:", "⁉️", -1)

	return msg
}

type TagToAction struct {
	Suffix      string
	Do          func(*TagToAction, bool, int, []rune) (int, string) // func(tagToAction,open,i,runes) (newI, output)
	Depth       int                                                 // For use by Do
	PartialMode bool
}

// TODO: Write a test for this
func tryStepForward(i int, step int, runes []rune) (int, bool) {
	i += step
	if i < len(runes) {
		return i, true
	}
	return i - step, false
}

// TODO: Write a test for this
func tryStepBackward(i int, step int, runes []rune) (int, bool) {
	if i == 0 {
		return i, false
	}
	return i - 1, true
}

// TODO: Preparse Markdown and normalize it into HTML?
func PreparseMessage(msg string) string {
	// TODO: Kick this check down a level into SanitiseBody?
	if !utf8.ValidString(msg) {
		return ""
	}
	msg = strings.Replace(msg, "<p><br>", "\n\n", -1)
	msg = strings.Replace(msg, "<p>", "\n\n", -1)
	msg = strings.Replace(msg, "</p>", "", -1)
	// TODO: Make this looser by moving it to the reverse HTML parser?
	msg = strings.Replace(msg, "<br>", "\n\n", -1)
	msg = strings.Replace(msg, "<br />", "\n\n", -1) // XHTML style
	msg = strings.Replace(msg, "&nbsp;", "", -1)
	msg = strings.Replace(msg, "\r", "", -1) // Windows artifact
	//msg = strings.Replace(msg, "\n\n\n\n", "\n\n\n", -1)
	msg = GetHookTable().Sshook("preparse_preassign", msg)
	// There are a few useful cases for having spaces, but I'd like to stop the WYSIWYG from inserting random lines here and there
	msg = SanitiseBody(msg)

	var runes = []rune(msg)
	msg = ""

	// TODO: We can maybe reduce the size of this by using an offset?
	// TODO: Move some of these closures out of this function to make things a little more efficient
	var allowedTags = [][]string{
		'e': []string{"m"},
		's': []string{"", "trong", "pan"},
		'd': []string{"el"},
		'u': []string{""},
		'b': []string{"", "lockquote"},
		'i': []string{""},
		'h': []string{"1", "2", "3"},
		//'p': []string{""},
		'g': []string{""}, // Quick and dirty fix for Grammarly
	}
	var buildLitMatch = func(tag string) func(*TagToAction, bool, int, []rune) (int, string) {
		return func(action *TagToAction, open bool, _ int, _ []rune) (int, string) {
			if open {
				action.Depth++
				return -1, "<" + tag + ">"
			}
			if action.Depth <= 0 {
				return -1, ""
			}
			action.Depth--
			return -1, "</" + tag + ">"
		}
	}
	var tagToAction = [][]*TagToAction{
		'e': []*TagToAction{&TagToAction{"m", buildLitMatch("em"), 0, false}},
		's': []*TagToAction{
			&TagToAction{"", buildLitMatch("del"), 0, false},
			&TagToAction{"trong", buildLitMatch("strong"), 0, false},
			// Hides the span tags Trumbowyg loves blasting out randomly
			&TagToAction{"pan", func(act *TagToAction, open bool, i int, runes []rune) (int, string) {
				if open {
					act.Depth++
					//fmt.Println("skipping attributes")
					for ; i < len(runes); i++ {
						if runes[i] == '&' && peekMatch(i, "gt;", runes) {
							//fmt.Println("found tag exit")
							return i + 3, " "
						}
					}
					return -1, " "
				}
				if act.Depth <= 0 {
					return -1, " "
				}
				act.Depth--
				return -1, " "
			}, 0, true},
		},
		'd': []*TagToAction{&TagToAction{"el", buildLitMatch("del"), 0, false}},
		'u': []*TagToAction{&TagToAction{"", buildLitMatch("u"), 0, false}},
		'b': []*TagToAction{
			&TagToAction{"", buildLitMatch("strong"), 0, false},
			&TagToAction{"lockquote", buildLitMatch("blockquote"), 0, false},
		},
		'i': []*TagToAction{&TagToAction{"", buildLitMatch("em"), 0, false}},
		'h': []*TagToAction{
			&TagToAction{"1", buildLitMatch("h2"), 0, false},
			&TagToAction{"2", buildLitMatch("h3"), 0, false},
			&TagToAction{"3", buildLitMatch("h4"), 0, false},
		},
		//'p': []*TagToAction{&TagToAction{"", buildLitMatch2("\n\n", ""), 0, false}},
		'g': []*TagToAction{
			&TagToAction{"", func(act *TagToAction, open bool, i int, runes []rune) (int, string) {
				if open {
					act.Depth++
					//fmt.Println("skipping attributes")
					for ; i < len(runes); i++ {
						if runes[i] == '&' && peekMatch(i, "gt;", runes) {
							//fmt.Println("found tag exit")
							return i + 3, " "
						}
					}
					return -1, " "
				}
				if act.Depth <= 0 {
					return -1, " "
				}
				act.Depth--
				return -1, " "
			}, 0, true},
		},
	}
	// TODO: Implement a less literal parser
	for i := 0; i < len(runes); i++ {
		char := runes[i]
		// TODO: Make the slashes escapable too in case someone means to use a literaly slash, maybe as an example of how to escape elements?
		if char == '\\' {
			if peekMatch(i, "&lt;", runes) {
				msg += "&"
				i++
			}
		} else if char == '&' && peekMatch(i, "lt;", runes) {
			var ok bool
			i, ok = tryStepForward(i, 4, runes)
			if !ok {
				msg += "&lt;"
				break
			}
			char := runes[i]
			if int(char) >= len(allowedTags) {
				//fmt.Println("sentinel char out of bounds")
				msg += "&"
				i -= 4
				continue
			}

			var closeTag bool
			if char == '/' {
				//fmt.Println("found close tag")
				i, ok = tryStepForward(i, 1, runes)
				if !ok {
					msg += "&lt;/"
					break
				}
				char = runes[i]
				closeTag = true
			}

			tags := allowedTags[char]
			if len(tags) == 0 {
				//fmt.Println("couldn't find char in allowedTags")
				if closeTag {
					//msg += "&lt;/"
					msg += "&"
					i -= 5
				} else {
					msg += "&"
					i -= 4
				}
				continue
			}
			// TODO: Scan through tags and make sure the suffix is present to reduce the number of false positives which hit the loop below
			//fmt.Printf("tags: %+v\n", tags)

			var newI = -1
			var out string
			toActionList := tagToAction[char]
			for _, toAction := range toActionList {
				// TODO: Optimise this, maybe with goto or a function call to avoid scanning the text twice?
				if (toAction.PartialMode && !closeTag && peekMatch(i, toAction.Suffix, runes)) || peekMatch(i, toAction.Suffix+"&gt;", runes) {
					newI, out = toAction.Do(toAction, !closeTag, i, runes)
					if newI != -1 {
						i = newI
					} else if out != "" {
						i += len(toAction.Suffix + "&gt;")
					}
					break
				}
			}
			if out == "" {
				msg += "&"
				if closeTag {
					i -= 5
				} else {
					i -= 4
				}
			} else if out != " " {
				msg += out
			}
		} else if char == '@' && (i == 0 || runes[i-1] < 33) {
			// TODO: Handle usernames containing spaces, maybe in the front-end with AJAX
			// Do not mention-ify ridiculously long things
			var ok bool
			i, ok = tryStepForward(i, 1, runes)
			if !ok {
				msg += "@"
				continue
			}
			start := i

			for j := 0; i < len(runes) && j < Config.MaxUsernameLength; j++ {
				cchar := runes[i]
				if cchar < 33 {
					break
				}
				i++
			}

			username := string(runes[start:i])
			if username == "" {
				msg += "@"
				i = start - 1
				continue
			}

			user, err := Users.GetByName(username)
			if err != nil {
				if err != ErrNoRows {
					LogError(err)
				}
				msg += "@"
				i = start - 1
				continue
			}
			msg += "@" + strconv.Itoa(user.ID)
			i--
		} else {
			msg += string(char)
		}
	}

	for _, actionList := range tagToAction {
		for _, toAction := range actionList {
			if toAction.Depth > 0 {
				for ; toAction.Depth > 0; toAction.Depth-- {
					_, out := toAction.Do(toAction, false, len(runes), runes)
					if out != "" {
						msg += out
					}
				}
			}
		}
	}
	return strings.TrimSpace(shortcodeToUnicode(msg))
}

// TODO: Test this
// TODO: Use this elsewhere in the parser?
func peek(cur int, skip int, runes []rune) rune {
	if (cur + skip) < len(runes) {
		return runes[cur+skip]
	}
	return 0 // null byte
}

// TODO: Test this
func peekMatch(cur int, phrase string, runes []rune) bool {
	if cur+len(phrase) > len(runes) {
		return false
	}
	for i, char := range phrase {
		if cur+i+1 >= len(runes) {
			return false
		}
		if runes[cur+i+1] != char {
			return false
		}
	}
	return true
}

// ! Not concurrency safe
func AddHashLinkType(prefix string, handler func(*strings.Builder, string, *int)) {
	// There can only be one hash link type starting with a specific character at the moment
	hashType := hashLinkTypes[prefix[0]]
	if hashType != "" {
		return
	}
	hashLinkMap[prefix] = handler
	hashLinkTypes[prefix[0]] = prefix
}

func writeURL(sb *strings.Builder, url string, label string) {
	sb.Write(URLOpen)
	sb.WriteString(url)
	sb.Write(URLOpen2)
	sb.WriteString(label)
	sb.Write(URLClose)
}

var hashLinkTypes = []string{'t': "tid-", 'r': "rid-", 'f': "fid-"}
var hashLinkMap = map[string]func(*strings.Builder, string, *int){
	"tid-": func(sb *strings.Builder, msg string, i *int) {
		tid, intLen := CoerceIntString(msg[*i:])
		*i += intLen

		topic, err := Topics.Get(tid)
		if err != nil || !Forums.Exists(topic.ParentID) {
			sb.Write(InvalidTopic)
			return
		}
		writeURL(sb, BuildTopicURL("", tid), "#tid-"+strconv.Itoa(tid))
	},
	"rid-": func(sb *strings.Builder, msg string, i *int) {
		rid, intLen := CoerceIntString(msg[*i:])
		*i += intLen

		topic, err := TopicByReplyID(rid)
		if err != nil || !Forums.Exists(topic.ParentID) {
			sb.Write(InvalidTopic)
			return
		}
		writeURL(sb, BuildTopicURL("", topic.ID), "#rid-"+strconv.Itoa(rid))
	},
	"fid-": func(sb *strings.Builder, msg string, i *int) {
		fid, intLen := CoerceIntString(msg[*i:])
		*i += intLen

		if !Forums.Exists(fid) {
			sb.Write(InvalidForum)
			return
		}
		writeURL(sb, BuildForumURL("", fid), "#fid-"+strconv.Itoa(fid))
	},
	// TODO: Forum Shortcode Link
}

// TODO: Write a test for this
// TODO: We need a lot more hooks here. E.g. To add custom media types and handlers.
// TODO: Use templates to reduce the amount of boilerplate?
func ParseMessage(msg string, sectionID int, sectionType string /*, user User*/) string {
	// TODO: Word boundary detection for these to avoid mangling code
	msg = strings.Replace(msg, ":)", "😀", -1)
	msg = strings.Replace(msg, ":(", "😞", -1)
	msg = strings.Replace(msg, ":D", "😃", -1)
	msg = strings.Replace(msg, ":P", "😛", -1)
	msg = strings.Replace(msg, ":O", "😲", -1)
	msg = strings.Replace(msg, ":p", "😛", -1)
	msg = strings.Replace(msg, ":o", "😲", -1)
	msg = strings.Replace(msg, ";)", "😉", -1)

	// Word filter list. E.g. Swear words and other things the admins don't like
	wordFilters, err := WordFilters.GetAll()
	if err != nil {
		LogError(err)
		return ""
	}
	for _, filter := range wordFilters {
		msg = strings.Replace(msg, filter.Find, filter.Replacement, -1)
	}
	msg += "          "

	// Search for URLs, mentions and hashlinks in the messages...
	var sb strings.Builder
	var lastItem = 0
	var i = 0
	for ; len(msg) > (i + 1); i++ {
		if (i == 0 && (msg[0] > 32)) || ((msg[i] < 33) && (msg[i+1] > 32)) {
			if (i != 0) || msg[i] < 33 {
				i++
			}
			if msg[i] == '#' {
				hashType := hashLinkTypes[msg[i+1]]
				if hashType == "" {
					continue
				}
				if msg[i+1:len(hashType)+1] == hashType {
					sb.WriteString(msg[lastItem:i])
					i += len(hashType) + 1
					hashLinkMap[hashType](&sb, msg, &i)
					lastItem = i
				}
				i--
			} else if msg[i] == '@' {
				sb.WriteString(msg[lastItem:i])
				i++
				start := i
				uid, intLen := CoerceIntString(msg[start:])
				i += intLen

				menUser, err := Users.Get(uid)
				if err != nil {
					sb.Write(InvalidProfile)
					lastItem = i
					continue
				}

				sb.Write(URLOpen)
				sb.WriteString(menUser.Link)
				sb.Write(bytesSinglequote)
				sb.Write(urlMention)
				sb.Write(bytesGreaterthan)
				sb.WriteString("@" + menUser.Name)
				sb.Write(URLClose)
				lastItem = i
				i--
			} else if msg[i] == 'h' || msg[i] == 'f' || msg[i] == 'g' || msg[i] == '/' {
				if msg[i+1] == 't' && msg[i+2] == 't' && msg[i+3] == 'p' {
					if msg[i+4] == 's' && msg[i+5] == ':' && msg[i+6] == '/' {
						// Do nothing
					} else if msg[i+4] == ':' && msg[i+5] == '/' {
						// Do nothing
					} else {
						continue
					}
				} else if msg[i+1] == 't' && msg[i+2] == 'p' && msg[i+3] == ':' && msg[i+4] == '/' {
					// Do nothing
				} else if msg[i+1] == 'i' && msg[i+2] == 't' && msg[i+3] == ':' && msg[i+4] == '/' {
					// Do nothing
				} else if msg[i+1] == '/' {
					// Do nothing
				} else {
					continue
				}

				sb.WriteString(msg[lastItem:i])
				urlLen := PartialURLStringLen(msg[i:])
				if msg[i+urlLen] > 32 { // space and invisibles
					sb.Write(InvalidURL)
					i += urlLen
					continue
				}

				media, ok := parseMediaString(msg[i : i+urlLen])
				if !ok {
					sb.Write(InvalidURL)
					i += urlLen
					continue
				}

				var addImage = func(url string) {
					sb.Write(imageOpen)
					sb.WriteString(url)
					sb.Write(imageOpen2)
					sb.WriteString(url)
					sb.Write(imageClose)
					i += urlLen
					lastItem = i
				}

				// TODO: Reduce the amount of code duplication
				if media.Type == "attach" {
					addImage(media.URL + "?sectionID=" + strconv.Itoa(sectionID) + "&amp;sectionType=" + sectionType)
					continue
				} else if media.Type == "image" {
					addImage(media.URL)
					continue
				} else if media.Type == "raw" {
					sb.WriteString(media.Body)
					i += urlLen
					lastItem = i
					continue
				} else if media.Type != "" {
					sb.Write(unknownMedia)
					i += urlLen
					continue
				}

				sb.Write(URLOpen)
				sb.WriteString(msg[i : i+urlLen])
				sb.Write(URLOpen2)
				sb.WriteString(msg[i : i+urlLen])
				sb.Write(URLClose)
				i += urlLen
				lastItem = i
				i--
			}
		}
	}
	if lastItem != i && sb.Len() != 0 {
		calclen := len(msg) - 10
		if calclen <= lastItem {
			calclen = lastItem
		}
		sb.WriteString(msg[lastItem:calclen])
		msg = sb.String()
	}

	msg = strings.Replace(msg, "\n", "<br>", -1)
	msg = GetHookTable().Sshook("parse_assign", msg)
	return msg
}

// 6, 7, 8, 6, 2, 7
// ftp://, http://, https:// git://, //, mailto: (not a URL, just here for length comparison purposes)
// TODO: Write a test for this
func validateURLString(data string) bool {
	datalen := len(data)
	i := 0

	if datalen >= 6 {
		if data[0:6] == "ftp://" || data[0:6] == "git://" {
			i = 6
		} else if datalen >= 7 && data[0:7] == "http://" {
			i = 7
		} else if datalen >= 8 && data[0:8] == "https://" {
			i = 8
		}
	} else if datalen >= 2 && data[0] == '/' && data[1] == '/' {
		i = 2
	}

	// ? - There should only be one : and that's only if the URL is on a non-standard port. Same for ?s.
	for ; datalen > i; i++ {
		if data[i] != '\\' && data[i] != '_' && data[i] != ':' && data[i] != '?' && data[i] != '&' && data[i] != '=' && data[i] != ';' && data[i] != '@' && data[i] != '#' && !(data[i] > 44 && data[i] < 58) && !(data[i] > 64 && data[i] < 91) && !(data[i] > 96 && data[i] < 123) {
			return false
		}
	}
	return true
}

// TODO: Write a test for this
func validatedURLBytes(data []byte) (url []byte) {
	datalen := len(data)
	i := 0

	if datalen >= 6 {
		if bytes.Equal(data[0:6], []byte("ftp://")) || bytes.Equal(data[0:6], []byte("git://")) {
			i = 6
		} else if datalen >= 7 && bytes.Equal(data[0:7], httpProtBytes) {
			i = 7
		} else if datalen >= 8 && bytes.Equal(data[0:8], []byte("https://")) {
			i = 8
		}
	} else if datalen >= 2 && data[0] == '/' && data[1] == '/' {
		i = 2
	}

	// ? - There should only be one : and that's only if the URL is on a non-standard port. Same for ?s.
	for ; datalen > i; i++ {
		if data[i] != '\\' && data[i] != '_' && data[i] != ':' && data[i] != '?' && data[i] != '&' && data[i] != '=' && data[i] != ';' && data[i] != '@' && data[i] != '#' && !(data[i] > 44 && data[i] < 58) && !(data[i] > 64 && data[i] < 91) && !(data[i] > 96 && data[i] < 123) {
			return InvalidURL
		}
	}

	url = append(url, data...)
	return url
}

// TODO: Write a test for this
func PartialURLString(data string) (url []byte) {
	datalen := len(data)
	i := 0
	end := datalen - 1

	if datalen >= 6 {
		if data[0:6] == "ftp://" || data[0:6] == "git://" {
			i = 6
		} else if datalen >= 7 && data[0:7] == "http://" {
			i = 7
		} else if datalen >= 8 && data[0:8] == "https://" {
			i = 8
		}
	} else if datalen >= 2 && data[0] == '/' && data[1] == '/' {
		i = 2
	}

	// ? - There should only be one : and that's only if the URL is on a non-standard port. Same for ?s.
	for ; end >= i; i++ {
		if data[i] != '\\' && data[i] != '_' && data[i] != ':' && data[i] != '?' && data[i] != '&' && data[i] != '=' && data[i] != ';' && data[i] != '@' && data[i] != '#' && !(data[i] > 44 && data[i] < 58) && !(data[i] > 64 && data[i] < 91) && !(data[i] > 96 && data[i] < 123) {
			end = i
		}
	}

	url = append(url, []byte(data[0:end])...)
	return url
}

// TODO: Write a test for this
func PartialURLStringLen(data string) int {
	datalen := len(data)
	i := 0

	if datalen >= 6 {
		//log.Print(string(data[0:5]))
		if data[0:6] == "ftp://" || data[0:6] == "git://" {
			i = 6
		} else if datalen >= 7 && data[0:7] == "http://" {
			i = 7
		} else if datalen >= 8 && data[0:8] == "https://" {
			i = 8
		}
	} else if datalen >= 2 && data[0] == '/' && data[1] == '/' {
		i = 2
	}

	// ? - There should only be one : and that's only if the URL is on a non-standard port. Same for ?s.
	for ; datalen > i; i++ {
		if data[i] != '\\' && data[i] != '_' && data[i] != ':' && data[i] != '?' && data[i] != '&' && data[i] != '=' && data[i] != ';' && data[i] != '@' && data[i] != '#' && !(data[i] > 44 && data[i] < 58) && !(data[i] > 64 && data[i] < 91) && !(data[i] > 96 && data[i] < 123) {
			//log.Print("Bad Character: ", data[i])
			return i
		}
	}
	//log.Print("Data Length: ",datalen)
	return datalen
}

type MediaEmbed struct {
	Type string //image
	URL  string
	Body string
}

// TODO: Write a test for this
func parseMediaString(data string) (media MediaEmbed, ok bool) {
	if !validateURLString(data) {
		return media, false
	}
	url, err := url.Parse(data)
	if err != nil {
		return media, false
	}

	hostname := url.Hostname()
	scheme := url.Scheme
	port := url.Port()
	query := url.Query()

	// TODO: Treat 127.0.0.1 and [::1] as localhost too
	var samesite = hostname == "localhost" || hostname == Site.URL
	if samesite {
		hostname = strings.Split(Site.URL, ":")[0]
		// ?- Test this as I'm not sure it'll do what it should. If someone's running SSL on port 80 or non-SSL on port 443 then... Well... They're in far worse trouble than this...
		port = Site.Port
		if Site.EnableSsl {
			scheme = "https"
		}
	}
	if scheme == "" {
		scheme = "http"
	}

	path := url.EscapedPath()
	pathFrags := strings.Split(path, "/")
	if len(pathFrags) >= 2 {
		if samesite && pathFrags[1] == "attachs" && (scheme == "http" || scheme == "https") {
			media.Type = "attach"
			var sport string
			// ? - Assumes the sysadmin hasn't mixed up the two standard ports
			if port != "443" && port != "80" {
				sport = ":" + port
			}
			media.URL = scheme + "://" + hostname + sport + path
			return media, true
		}
	}

	// ? - I don't think this hostname will hit every YT domain
	// TODO: Make this a more customisable handler rather than hard-coding it in here
	if hostname == "www.youtube.com" && path == "/watch" {
		video, ok := query["v"]
		if ok && len(video) >= 1 && video[0] != "" {
			media.Type = "raw"
			// TODO: Filter the URL to make sure no nasties end up in there
			media.Body = "<iframe class='postIframe' src='https://www.youtube-nocookie.com/embed/" + video[0] + "' frameborder='0' allowfullscreen></iframe>"
			return media, true
		}
	}

	lastFrag := pathFrags[len(pathFrags)-1]
	if lastFrag != "" {
		// TODO: Write a function for getting the file extension of a string
		extarr := strings.Split(lastFrag, ".")
		if len(extarr) >= 2 {
			ext := extarr[len(extarr)-1]
			if ImageFileExts.Contains(ext) {
				media.Type = "image"
				var sport string
				if port != "443" && port != "80" {
					sport = ":" + port
				}
				media.URL = scheme + "://" + hostname + sport + path
				return media, true
			}
		}
	}

	return media, true
}

// TODO: Write a test for this
func CoerceIntString(data string) (res int, length int) {
	if !(data[0] > 47 && data[0] < 58) {
		return 0, 1
	}

	i := 0
	for ; len(data) > i; i++ {
		if !(data[i] > 47 && data[i] < 58) {
			conv, err := strconv.Atoi(data[0:i])
			if err != nil {
				return 0, i
			}
			return conv, i
		}
	}

	conv, err := strconv.Atoi(data)
	if err != nil {
		return 0, i
	}
	return conv, i
}

// TODO: Write tests for this
// Make sure we reflect changes to this in the JS port in /public/global.js
func Paginate(count int, perPage int, maxPages int) []int {
	if count < perPage {
		return []int{1}
	}
	var page int
	var out []int
	for current := 0; current < count; current += perPage {
		page++
		out = append(out, page)
		if len(out) >= maxPages {
			break
		}
	}
	return out
}

// TODO: Write tests for this
// Make sure we reflect changes to this in the JS port in /public/global.js
func PageOffset(count int, page int, perPage int) (int, int, int) {
	var offset int
	lastPage := LastPage(count, perPage)
	if page > 1 {
		offset = (perPage * page) - perPage
	} else if page == -1 {
		page = lastPage
		offset = (perPage * page) - perPage
	} else {
		page = 1
	}

	// ? - This has been commented out as it created a bug in the user manager where the first user on a page wouldn't be accessible
	// We don't want the offset to overflow the slices, if everything's in memory
	/*if offset >= (count - 1) {
		offset = 0
	}*/
	return offset, page, lastPage
}

// TODO: Write tests for this
// Make sure we reflect changes to this in the JS port in /public/global.js
func LastPage(count int, perPage int) int {
	return (count / perPage) + 1
}
