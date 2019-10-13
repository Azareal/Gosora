package common

import (
	"bytes"
	//"fmt"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// TODO: Somehow localise these?
var SpaceGap = []byte("          ")
var httpProtBytes = []byte("http://")
var DoubleForwardSlash = []byte("//")
var InvalidURL = []byte("<red>[Invalid URL]</red>")
var InvalidTopic = []byte("<red>[Invalid Topic]</red>")
var InvalidProfile = []byte("<red>[Invalid Profile]</red>")
var InvalidForum = []byte("<red>[Invalid Forum]</red>")
var unknownMedia = []byte("<red>[Unknown Media]</red>")
var URLOpen = []byte("<a href='")
var URLOpenUser = []byte("<a rel='ugc' href='")
var URLOpen2 = []byte("'>")
var bytesSinglequote = []byte("'")
var bytesGreaterthan = []byte(">")
var urlMention = []byte(" class='mention'")
var URLClose = []byte("</a>")
var imageOpen = []byte("<a href=\"")
var imageOpen2 = []byte("\"><img src='")
var imageClose = []byte("' class='postImage' /></a>")
var attachOpen = []byte("<a download class='attach' href=\"")
var attachClose = []byte("\">Attachment</a>")
var sidParam = []byte("?sid=")
var stypeParam = []byte("&amp;stype=")
var urlPattern = `(?s)([ {1}])((http|https|ftp|mailto)*)(:{??)\/\/([\.a-zA-Z\/]+)([ {1}])`
var urlReg *regexp.Regexp

func init() {
	urlReg = regexp.MustCompile(urlPattern)
}

var emojis map[string]string

type emojiHolder struct {
	Emojis []map[string]string `json:"emojis"`
}

func InitEmoji() error {
	data, err := ioutil.ReadFile("./config/emoji_default.json")
	if err != nil {
		return err
	}

	var emoji emojiHolder
	err = json.Unmarshal(data, &emoji)
	if err != nil {
		return err
	}

	emojis = make(map[string]string, len(emoji.Emojis))
	for _, item := range emoji.Emojis {
		for ikey, ival := range item {
			emojis[ikey] = ival
		}
	}

	data, err = ioutil.ReadFile("./config/emoji.json")
	if err == os.ErrPermission || err == os.ErrClosed {
		return err
	} else if err != nil {
		return nil
	}

	emoji = emojiHolder{}
	err = json.Unmarshal(data, &emoji)
	if err != nil {
		return err
	}

	for _, item := range emoji.Emojis {
		for ikey, ival := range item {
			emojis[ikey] = ival
		}
	}

	return nil
}

// TODO: Write a test for this
func shortcodeToUnicode(msg string) string {
	//re := regexp.MustCompile(":(.):")
	for shortcode, emoji := range emojis {
		msg = strings.Replace(msg, shortcode, emoji, -1)
	}
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
				msg += "&"
				if closeTag {
					//msg += "&lt;/"
					//msg += "&"
					i -= 5
				} else {
					//msg += "&"
					i -= 4
				}
				continue
			}
			// TODO: Scan through tags and make sure the suffix is present to reduce the number of false positives which hit the loop below
			//fmt.Printf("tags: %+v\n", tags)

			newI := -1
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

func WriteURL(sb *strings.Builder, url string, label string) {
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
		WriteURL(sb, BuildTopicURL("", tid), "#tid-"+strconv.Itoa(tid))
	},
	"rid-": func(sb *strings.Builder, msg string, i *int) {
		rid, intLen := CoerceIntString(msg[*i:])
		*i += intLen

		topic, err := TopicByReplyID(rid)
		if err != nil || !Forums.Exists(topic.ParentID) {
			sb.Write(InvalidTopic)
			return
		}
		// TODO: Send the user to the right page and post not just the right topic?
		WriteURL(sb, BuildTopicURL("", topic.ID), "#rid-"+strconv.Itoa(rid))
	},
	"fid-": func(sb *strings.Builder, msg string, i *int) {
		fid, intLen := CoerceIntString(msg[*i:])
		*i += intLen

		if !Forums.Exists(fid) {
			sb.Write(InvalidForum)
			return
		}
		WriteURL(sb, BuildForumURL("", fid), "#fid-"+strconv.Itoa(fid))
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

	// Search for URLs, mentions and hashlinks in the messages...
	var sb strings.Builder
	lastItem := 0
	i := 0
	//var c bool
	//fmt.Println("msg:", "'"+msg+"'")
	for ; len(msg) > i; i++ {
		//fmt.Printf("msg[%d]: %s\n",i,string(msg[i]))
		if (i == 0 && (msg[0] > 32)) || (len(msg) > (i+1) && (msg[i] < 33) && (msg[i+1] > 32)) {
			//fmt.Println("s1")
			if (i != 0) || msg[i] < 33 {
				i++
			}
			if len(msg) <= (i + 1) {
				break
			}
			//fmt.Println("s2")
			if msg[i] == '#' {
				//fmt.Println("msg[i+1]:", msg[i+1])
				//fmt.Println("string(msg[i+1]):", string(msg[i+1]))
				hashType := hashLinkTypes[msg[i+1]]
				if hashType == "" {
					//fmt.Println("uh1")
					sb.WriteString(msg[lastItem:i])
					i++
					lastItem = i
					continue
				}
				//fmt.Println("hashType:", hashType)
				if len(msg) <= (i + len(hashType) + 1) {
					sb.WriteString(msg[lastItem:i])
					lastItem = i
					continue
				}
				if msg[i+1:i+len(hashType)+1] != hashType {
					continue
				}

				//fmt.Println("msg[lastItem:i]:", msg[lastItem:i])
				sb.WriteString(msg[lastItem:i])
				i += len(hashType) + 1
				hashLinkMap[hashType](&sb, msg, &i)
				lastItem = i
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
					i--
					continue
				}

				sb.Write(URLOpen)
				sb.WriteString(menUser.Link)
				sb.Write(bytesSinglequote)
				sb.Write(urlMention)
				sb.Write(bytesGreaterthan)
				sb.WriteByte('@')
				sb.WriteString(menUser.Name)
				sb.Write(URLClose)
				lastItem = i
				i--
			} else if msg[i] == 'h' || msg[i] == 'f' || msg[i] == 'g' || msg[i] == '/' {
				//fmt.Println("s3")
				if len(msg) > i+3 && msg[i+1] == 't' && msg[i+2] == 't' && msg[i+3] == 'p' {
					if len(msg) > i+6 && msg[i+4] == 's' && msg[i+5] == ':' && msg[i+6] == '/' {
						// Do nothing
					} else if len(msg) > i+5 && msg[i+4] == ':' && msg[i+5] == '/' {
						// Do nothing
					} else {
						continue
					}
				} else if len(msg) > i+4 && msg[i+1] == 't' && msg[i+2] == 'p' && msg[i+3] == ':' && msg[i+4] == '/' {
					// Do nothing
				} else if len(msg) > i+4 && msg[i+1] == 'i' && msg[i+2] == 't' && msg[i+3] == ':' && msg[i+4] == '/' {
					// Do nothing
				} else if msg[i+1] == '/' {
					// Do nothing
				} else {
					continue
				}

				//fmt.Println("p1:",i)
				sb.WriteString(msg[lastItem:i])
				urlLen, ok := PartialURLStringLen(msg[i:])
				if len(msg) < i+urlLen {
					//fmt.Println("o1")
					if urlLen == 2 {
						sb.Write(DoubleForwardSlash)
					} else {
						sb.Write(InvalidURL)
					}
					i += len(msg) - 1
					lastItem = i
					break
				}
				if urlLen == 2 {
					sb.Write(DoubleForwardSlash)
					i += urlLen
					lastItem = i
					i--
					continue
				}
				//fmt.Println("msg[i:i+urlLen]:", "'"+msg[i:i+urlLen]+"'")
				if !ok {
					//fmt.Printf("o2: i = %d; i+urlLen = %d\n",i,i+urlLen)
					sb.Write(InvalidURL)
					i += urlLen
					lastItem = i
					i--
					continue
				}

				media, ok := parseMediaString(msg[i : i+urlLen])
				if !ok {
					//fmt.Println("o3")
					sb.Write(InvalidURL)
					i += urlLen
					lastItem = i
					continue
				}
				//fmt.Println("p2")

				addImage := func(url string) {
					sb.Grow(len(imageOpen) + len(url) + len(url) + len(imageOpen2) + len(imageClose))
					sb.Write(imageOpen)
					sb.WriteString(url)
					sb.Write(imageOpen2)
					sb.WriteString(url)
					sb.Write(imageClose)
					i += urlLen
					lastItem = i
				}

				// TODO: Reduce the amount of code duplication
				// TODO: Avoid allocating a string for media.Type?
				if media.Type == "attach" {
					addImage(media.URL + "?sid=" + strconv.Itoa(sectionID) + "&amp;stype=" + sectionType)
					continue
				} else if media.Type == "image" {
					addImage(media.URL)
					continue
				} else if media.Type == "aother" {
					sb.Write(attachOpen)
					sb.WriteString(media.URL)
					sb.Write(sidParam)
					sb.WriteString(strconv.Itoa(sectionID))
					sb.Write(stypeParam)
					sb.WriteString(sectionType)
					sb.Write(attachClose)
					i += urlLen
					lastItem = i
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
				//fmt.Println("p3")

				// TODO: Add support for rel="ugc"
				sb.Grow(len(URLOpen) + (len(msg[i:i+urlLen]) * 2) + len(URLOpen2) + len(URLClose))
				if media.Trusted {
					sb.Write(URLOpen)
				} else {
					sb.Write(URLOpenUser)
				}
				sb.WriteString(msg[i : i+urlLen])
				sb.Write(URLOpen2)
				sb.WriteString(media.FURL)
				sb.Write(URLClose)
				i += urlLen
				lastItem = i
				i--
			}
		}
	}
	if lastItem != i && sb.Len() != 0 {
		/*calclen := len(msg)
		if calclen <= lastItem {
			calclen = lastItem
		}*/
		//if i == len(msg) {
		sb.WriteString(msg[lastItem:])
		/*} else {
			sb.WriteString(msg[lastItem:calclen])
		}*/
	}
	if sb.Len() != 0 {
		msg = sb.String()
		//fmt.Println("sb.String():", "'"+sb.String()+"'")
	}

	msg = strings.Replace(msg, "\n", "<br>", -1)
	msg = GetHookTable().Sshook("parse_assign", msg)
	return msg
}

// 6, 7, 8, 6, 2, 7
// ftp://, http://, https:// git://, //, mailto: (not a URL, just here for length comparison purposes)
// TODO: Write a test for this
func validateURLString(data string) bool {
	i := 0
	if len(data) >= 6 {
		if data[0:6] == "ftp://" || data[0:6] == "git://" {
			i = 6
		} else if len(data) >= 7 && data[0:7] == "http://" {
			i = 7
		} else if len(data) >= 8 && data[0:8] == "https://" {
			i = 8
		}
	} else if len(data) >= 2 && data[0] == '/' && data[1] == '/' {
		i = 2
	}

	// ? - There should only be one : and that's only if the URL is on a non-standard port. Same for ?s.
	for ; len(data) > i; i++ {
		char := data[i]
		if char != '\\' && char != '_' && char != ':' && char != '?' && char != '&' && char != '=' && char != ';' && char != '@' && char != '#' && char != ']' && !(char > 44 && char < 58) && !(char > 64 && char < 92) && !(char > 96 && char < 123) { // 90 is Z, 91 is [
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
		char := data[i]
		if char != '\\' && char != '_' && char != ':' && char != '?' && char != '&' && char != '=' && char != ';' && char != '@' && char != '#' && char != ']' && !(char > 44 && char < 58) && !(char > 64 && char < 92) && !(char > 96 && char < 123) { // 90 is Z, 91 is [
			return InvalidURL
		}
	}

	url = append(url, data...)
	return url
}

// TODO: Write a test for this
func PartialURLString(data string) (url []byte) {
	i := 0
	end := len(data) - 1
	if len(data) >= 6 {
		if data[0:6] == "ftp://" || data[0:6] == "git://" {
			i = 6
		} else if len(data) >= 7 && data[0:7] == "http://" {
			i = 7
		} else if len(data) >= 8 && data[0:8] == "https://" {
			i = 8
		}
	} else if len(data) >= 2 && data[0] == '/' && data[1] == '/' {
		i = 2
	}

	// ? - There should only be one : and that's only if the URL is on a non-standard port. Same for ?s.
	for ; end >= i; i++ {
		char := data[i]
		if char != '\\' && char != '_' && char != ':' && char != '?' && char != '&' && char != '=' && char != ';' && char != '@' && char != '#' && char != ']' && !(char > 44 && char < 58) && !(char > 64 && char < 92) && !(char > 96 && char < 123) { // 90 is Z, 91 is [
			end = i
		}
	}

	url = append(url, []byte(data[0:end])...)
	return url
}

// TODO: Write a test for this
// TODO: Handle the host bits differently from the paths...
func PartialURLStringLen(data string) (int, bool) {
	i := 0
	if len(data) >= 6 {
		//log.Print(string(data[0:5]))
		if data[0:6] == "ftp://" || data[0:6] == "git://" {
			i = 6
		} else if len(data) >= 7 && data[0:7] == "http://" {
			i = 7
		} else if len(data) >= 8 && data[0:8] == "https://" {
			i = 8
		}
	} else if len(data) >= 2 && data[0] == '/' && data[1] == '/' {
		i = 2
	}
	//fmt.Println("Data Length: ",len(data))
	if len(data) < i {
		//fmt.Println("e1:",i)
		return i + 1, false
	}

	// ? - There should only be one : and that's only if the URL is on a non-standard port. Same for ?s.
	f := i
	//fmt.Println("f:",f)
	for ; len(data) > i; i++ {
		char := data[i]
		if char < 33 { // space and invisibles
			//fmt.Println("e2:",i)
			return i, i != f
		} else if char != '\\' && char != '_' && char != ':' && char != '?' && char != '&' && char != '=' && char != ';' && char != '@' && char != '#' && char != ']' && !(char > 44 && char < 58) && !(char > 64 && char < 92) && !(char > 96 && char < 123) { // 90 is Z, 91 is [
			//log.Print("Bad Character: ", char)
			//fmt.Println("e3")
			return i, false
		}
	}

	//fmt.Println("e4:", i)
	/*if data[i-1] < 33 {
		return i-1, i != f
	}*/
	//fmt.Println("e5")
	return i, i != f
}

// TODO: Write a test for this
// TODO: Get this to support IPv6 hosts, this isn't currently done as this is used in the bbcode plugin where it thinks the [ is a IPv6 host
func PartialURLStringLen2(data string) int {
	i := 0
	if len(data) >= 6 {
		//log.Print(string(data[0:5]))
		if data[0:6] == "ftp://" || data[0:6] == "git://" {
			i = 6
		} else if len(data) >= 7 && data[0:7] == "http://" {
			i = 7
		} else if len(data) >= 8 && data[0:8] == "https://" {
			i = 8
		}
	} else if len(data) >= 2 && data[0] == '/' && data[1] == '/' {
		i = 2
	}

	// ? - There should only be one : and that's only if the URL is on a non-standard port. Same for ?s.
	for ; len(data) > i; i++ {
		char := data[i]
		if char != '\\' && char != '_' && char != ':' && char != '?' && char != '&' && char != '=' && char != ';' && char != '@' && char != '#' && !(char > 44 && char < 58) && !(char > 64 && char < 91) && !(char > 96 && char < 123) { // 90 is Z, 91 is [
			//log.Print("Bad Character: ", char)
			return i
		}
	}
	//log.Print("Data Length: ",len(data))
	return len(data)
}

type MediaEmbed struct {
	Type string //image
	URL  string
	FURL string
	Body string

	Trusted bool // samesite urls
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

	samesite := hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" || hostname == Site.URL
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
	media.Trusted = samesite

	path := url.EscapedPath()
	pathFrags := strings.Split(path, "/")
	if len(pathFrags) >= 2 {
		if samesite && pathFrags[1] == "attachs" && (scheme == "http" || scheme == "https") {
			var sport string
			// ? - Assumes the sysadmin hasn't mixed up the two standard ports
			if port != "443" && port != "80" && port != "" {
				sport = ":" + port
			}
			media.URL = scheme + "://" + hostname + sport + path
			extarr := strings.Split(path, ".")
			if len(extarr) == 0 {
				// TODO: Write a unit test for this
				return media, false
			}
			ext := extarr[len(extarr)-1]
			if ImageFileExts.Contains(ext) {
				media.Type = "attach"
			} else {
				media.Type = "aother"
			}
			return media, true
		}
	}

	// ? - I don't think this hostname will hit every YT domain
	// TODO: Make this a more customisable handler rather than hard-coding it in here
	if strings.HasSuffix(hostname, ".youtube.com") && path == "/watch" {
		video, ok := query["v"]
		if ok && len(video) >= 1 && video[0] != "" {
			media.Type = "raw"
			// TODO: Filter the URL to make sure no nasties end up in there
			media.Body = "<iframe class='postIframe' src='https://www.youtube-nocookie.com/embed/" + video[0] + "' frameborder=0 allowfullscreen></iframe>"
			return media, true
		}
	}

	if lastFrag := pathFrags[len(pathFrags)-1]; lastFrag != "" {
		// TODO: Write a function for getting the file extension of a string
		if extarr := strings.Split(lastFrag, "."); len(extarr) >= 2 {
			ext := extarr[len(extarr)-1]
			if ImageFileExts.Contains(ext) {
				media.Type = "image"
				var sport string
				if port != "443" && port != "80" && port != "" {
					sport = ":" + port
				}
				media.URL = scheme + "://" + hostname + sport + path
				return media, true
			}
		}
	}

	var sport string
	if port != "443" && port != "80" && port != "" {
		sport = ":" + port
	}
	media.FURL = hostname + sport + path

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
func Paginate(currentPage int, lastPage int, maxPages int) (out []int) {
	diff := lastPage - currentPage
	pre := 3
	if diff < 3 {
		pre = maxPages - diff
	}

	page := currentPage - pre
	if page < 0 {
		page = 0
	}
	for len(out) < maxPages && page < lastPage {
		page++
		out = append(out, page)
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
