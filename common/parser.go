package common

import (
	"bytes"
	//"fmt"
	//"log"

	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// TODO: Use the template system?
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
var URLOpenUser = []byte("<a rel='ugc'href='")
var URLOpen2 = []byte("'>")
var bytesSinglequote = []byte("'")
var bytesGreaterThan = []byte(">")
var urlMention = []byte("'class='mention'")
var URLClose = []byte("</a>")
var videoOpen = []byte("<video controls src=\"")
var videoOpen2 = []byte("\"><a class='attach'href=\"")
var videoClose = []byte("\"download>Attachment</a></video>")
var audioOpen = []byte("<audio controls src=\"")
var audioOpen2 = []byte("\"><a class='attach'href=\"")
var audioClose = []byte("\"download>Attachment</a></audio>")
var imageOpen = []byte("<a href=\"")
var imageOpen2 = []byte("\"><img src='")
var imageClose = []byte("'class='postImage'></a>")
var attachOpen = []byte("<a class='attach'href=\"")
var attachClose = []byte("\"download>Attachment</a>")
var sidParam = []byte("?sid=")
var stypeParam = []byte("&amp;stype=")

/*var textShortOpen = []byte("<a class='attach'href=\"")
var textShortOpen2 = []byte("\">View</a> / <a class='attach'href=\"")
var textShortClose = []byte("\"download>Download</a>")*/
var textOpen = []byte("<div class='attach_box'><div class='attach_info'>")
var textOpen2 = []byte("</div><div class='attach_opts'><a class='attach'href=\"")
var textOpen3 = []byte("\">View</a> / <a class='attach'href=\"")
var textClose = []byte("\"download>Download</a></div></div>")
var urlPattern = `(?s)([ {1}])((http|https|ftp|mailto)*)(:{??)\/\/([\.a-zA-Z\/]+)([ {1}])`
var urlReg *regexp.Regexp

const imageSizeHint = len("<a href=\"") + len("\"><img src='") + len("'class='postImage'></a>")
const videoSizeHint = len("<video controls src=\"") + len("\"><a class='attach'href=\"") + len("\"download>Attachment</a></video>") + len("?sid=") + len("&amp;stype=") + 8
const audioSizeHint = len("<audio controls src=\"") + len("\"><a class='attach'href=\"") + len("\"download>Attachment</a></audio>") + len("?sid=") + len("&amp;stype=") + 8
const mentionSizeHint = len("<a href='") + len("'class='mention'") + len(">@") + len("</a>")

func init() {
	urlReg = regexp.MustCompile(urlPattern)
}

var emojis map[string]string

type emojiHolder struct {
	NoDefault bool                `json:"no_defaults"`
	Emojis    []map[string]string `json:"emojis"`
}

func InitEmoji() error {
	var emoji emojiHolder
	err := unmarshalJsonFile("./config/emoji_default.json", &emoji)
	if err != nil {
		return err
	}

	emojis = make(map[string]string, len(emoji.Emojis))
	if !emoji.NoDefault {
		for _, item := range emoji.Emojis {
			for ikey, ival := range item {
				emojis[ikey] = ival
			}
		}
	}

	emoji = emojiHolder{}
	err = unmarshalJsonFileIgnore404("./config/emoji.json", &emoji)
	if err != nil {
		return err
	}
	if emoji.NoDefault {
		emojis = make(map[string]string)
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
func tryStepForward(i, step int, runes []rune) (int, bool) {
	i += step
	if i < len(runes) {
		return i, true
	}
	return i - step, false
}

// TODO: Write a test for this
func tryStepBackward(i, step int, runes []rune) (int, bool) {
	if i == 0 {
		return i, false
	}
	return i - 1, true
}

// TODO: Preparse Markdown and normalize it into HTML?
// TODO: Use a string builder
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

	runes := []rune(msg)
	msg = ""

	// TODO: We can maybe reduce the size of this by using an offset?
	// TODO: Move some of these closures out of this function to make things a little more efficient
	allowedTags := [][]string{
		'e': {"m"},
		's': {"", "trong", "poiler", "pan"},
		'd': {"el"},
		'u': {""},
		'b': {"", "lockquote"},
		'i': {""},
		'h': {"1", "2", "3"},
		//'p': {""},
		'g': {""}, // Quick and dirty fix for Grammarly
	}
	buildLitMatch := func(tag string) func(*TagToAction, bool, int, []rune) (int, string) {
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
	tagToAction := [][]*TagToAction{
		'e': {{"m", buildLitMatch("em"), 0, false}},
		's': {
			{"", buildLitMatch("del"), 0, false},
			{"trong", buildLitMatch("strong"), 0, false},
			{"poiler", buildLitMatch("spoiler"), 0, false},
			// Hides the span tags Trumbowyg loves blasting out randomly
			{"pan", func(act *TagToAction, open bool, i int, runes []rune) (int, string) {
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
		'd': {{"el", buildLitMatch("del"), 0, false}},
		'u': {{"", buildLitMatch("u"), 0, false}},
		'b': {
			{"", buildLitMatch("strong"), 0, false},
			{"lockquote", buildLitMatch("blockquote"), 0, false},
		},
		'i': {{"", buildLitMatch("em"), 0, false}},
		'h': {
			{"1", buildLitMatch("h2"), 0, false},
			{"2", buildLitMatch("h3"), 0, false},
			{"3", buildLitMatch("h4"), 0, false},
		},
		//'p': {{"", buildLitMatch2("\n\n", ""), 0, false}},
		'g': {
			{"", func(act *TagToAction, open bool, i int, runes []rune) (int, string) {
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
	// TODO: Use a string builder
	// TODO: Implement faster emoji parser
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
func peek(cur, skip int, runes []rune) rune {
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
func AddHashLinkType(prefix string, h func(*strings.Builder, string, *int)) {
	// There can only be one hash link type starting with a specific character at the moment
	hashType := hashLinkTypes[prefix[0]]
	if hashType != "" {
		return
	}
	hashLinkMap[prefix] = h
	hashLinkTypes[prefix[0]] = prefix
}

func WriteURL(sb *strings.Builder, url, label string) {
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

// TODO: Pack multiple bit flags into an integer instead of using a struct?
var DefaultParseSettings = &ParseSettings{}

type ParseSettings struct {
	NoEmbed bool
}

func (ps *ParseSettings) CopyPtr() *ParseSettings {
	n := &ParseSettings{}
	*n = *ps
	return n
}

func ParseMessage(msg string, sectionID int, sectionType string, settings *ParseSettings, user *User) string {
	msg, _ = ParseMessage2(msg, sectionID, sectionType, settings, user)
	return msg
}

var litRepPrefix = []byte{':', ';'}

//var litRep = [][]byte{':':{')','(','D','O','o','P','p'},';':{')'}}
var litRep = [][]string{':': {')': "😀", '(': "😞", 'D': "😃", 'O': "😲", 'o': "😲", 'P': "😛", 'p': "😛"}, ';': {')': "😉"}}

// TODO: Write a test for this
// TODO: We need a lot more hooks here. E.g. To add custom media types and handlers.
// TODO: Use templates to reduce the amount of boilerplate?
func ParseMessage2(msg string, sectionID int, sectionType string, settings *ParseSettings, user *User) (string, bool) {
	if settings == nil {
		settings = DefaultParseSettings
	}
	if user == nil {
		user = &GuestUser
	}
	// TODO: Word boundary detection for these to avoid mangling code
	/*rep := func(find, replace string) {
		msg = strings.Replace(msg, find, replace, -1)
	}
	rep(":)", "😀")
	rep(":(", "😞")
	rep(":D", "😃")
	rep(":P", "😛")
	rep(":O", "😲")
	rep(":p", "😛")
	rep(":o", "😲")
	rep(";)", "😉")*/

	// Word filter list. E.g. Swear words and other things the admins don't like
	filters, err := WordFilters.GetAll()
	if err != nil {
		LogError(err)
		return "", false
	}
	for _, f := range filters {
		msg = strings.Replace(msg, f.Find, f.Replace, -1)
	}
	if len(msg) < 2 {
		msg = strings.Replace(msg, "\n", "<br>", -1)
		msg = GetHookTable().Sshook("parse_assign", msg)
		return msg, false
	}

	// Search for URLs, mentions and hashlinks in the messages...
	var sb strings.Builder
	lastItem := 0
	i := 0
	var externalHead bool
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
			ch := msg[i]

			// Very short literal matcher
			if len(litRep) > int(ch) {
				sl := litRep[ch]
				if sl != nil {
					i++
					ch := msg[i]
					if len(sl) > int(ch) {
						val := sl[ch]
						if val != "" {
							i--
							sb.WriteString(msg[lastItem:i])
							i++
							sb.WriteString(val)
							i++
							lastItem = i
							i--
							continue
						}
					}
					i--
				}
				//lastItem = i
				//i--
				//continue
			}

			switch ch {
			case '#':
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
			case '@':
				sb.WriteString(msg[lastItem:i])
				i++
				start := i
				uid, intLen := CoerceIntString(msg[start:])
				i += intLen

				var menUser *User
				if uid != 0 && user.ID == uid {
					menUser = user
				} else {
					menUser = Users.Getn(uid)
					if menUser == nil {
						sb.Write(InvalidProfile)
						lastItem = i
						i--
						continue
					}
				}

				sb.Grow(mentionSizeHint + len(menUser.Link) + len(menUser.Name))
				sb.Write(URLOpen)
				sb.WriteString(menUser.Link)
				sb.Write(urlMention)
				sb.Write(bytesGreaterThan)
				sb.WriteByte('@')
				sb.WriteString(menUser.Name)
				sb.Write(URLClose)
				lastItem = i
				i--
			case 'h', 'f', 'g', '/':
				//fmt.Println("s3")
				if len(msg) > i+5 && msg[i+1] == 't' && msg[i+2] == 't' && msg[i+3] == 'p' {
					if len(msg) > i+6 && msg[i+4] == 's' && msg[i+5] == ':' && msg[i+6] == '/' {
						// Do nothing
					} else if msg[i+4] == ':' && msg[i+5] == '/' {
						// Do nothing
					} else {
						continue
					}
				} else if len(msg) > i+4 {
					fch := msg[i+1]
					if fch == 't' && msg[i+2] == 'p' && msg[i+3] == ':' && msg[i+4] == '/' {
						// Do nothing
					} else if fch == 'i' && msg[i+2] == 't' && msg[i+3] == ':' && msg[i+4] == '/' {
						// Do nothing
					} else if fch == '/' {
						// Do nothing
					} else {
						continue
					}
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

				media, ok := parseMediaString(msg[i:i+urlLen], settings)
				if !ok {
					//fmt.Println("o3")
					sb.Write(InvalidURL)
					i += urlLen
					lastItem = i
					continue
				}
				//fmt.Println("p2")

				addImage := func(url string) {
					sb.Grow(imageSizeHint + len(url) + len(url))
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
				switch media.Type {
				case AImage:
					addImage(media.URL + "?sid=" + strconv.Itoa(sectionID) + "&amp;stype=" + sectionType)
					continue
				case AVideo:
					sb.Grow(videoSizeHint + (len(media.URL) + len(sectionType)*2))
					sb.Write(videoOpen)
					sb.WriteString(media.URL)
					sb.Write(sidParam)
					sb.WriteString(strconv.Itoa(sectionID))
					sb.Write(stypeParam)
					sb.WriteString(sectionType)
					sb.Write(videoOpen2)
					sb.WriteString(media.URL)
					sb.Write(sidParam)
					sb.WriteString(strconv.Itoa(sectionID))
					sb.Write(stypeParam)
					sb.WriteString(sectionType)
					sb.Write(videoClose)
					i += urlLen
					lastItem = i
					continue
				case AAudio:
					sb.Grow(audioSizeHint + (len(media.URL) + len(sectionType)*2))
					sb.Write(audioOpen)
					sb.WriteString(media.URL)
					sb.Write(sidParam)
					sb.WriteString(strconv.Itoa(sectionID))
					sb.Write(stypeParam)
					sb.WriteString(sectionType)
					sb.Write(audioOpen2)
					sb.WriteString(media.URL)
					sb.Write(sidParam)
					sb.WriteString(strconv.Itoa(sectionID))
					sb.Write(stypeParam)
					sb.WriteString(sectionType)
					sb.Write(audioClose)
					i += urlLen
					lastItem = i
					continue
				case EImage:
					addImage(media.URL)
					continue
				case AText:
					/*sb.Write(textOpen)
					sb.WriteString(media.URL)
					sb.Write(sidParam)
					sid := strconv.Itoa(sectionID)
					sb.WriteString(sid)
					sb.Write(stypeParam)
					sb.WriteString(sectionType)
					sb.Write(textOpen2)
					sb.WriteString(media.URL)
					sb.Write(sidParam)
					sb.WriteString(sid)
					sb.Write(stypeParam)
					sb.WriteString(sectionType)
					sb.Write(textClose)
					i += urlLen
					lastItem = i
					continue*/
					sb.Write(textOpen)
					sb.WriteString(media.URL)
					sb.Write(textOpen2)
					sb.WriteString(media.URL)
					sb.Write(sidParam)
					sid := strconv.Itoa(sectionID)
					sb.WriteString(sid)
					sb.Write(stypeParam)
					sb.WriteString(sectionType)
					sb.Write(textOpen3)
					sb.WriteString(media.URL)
					sb.Write(sidParam)
					sb.WriteString(sid)
					sb.Write(stypeParam)
					sb.WriteString(sectionType)
					sb.Write(textClose)
					i += urlLen
					lastItem = i
					continue
				case AOther:
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
				case ERaw:
					sb.WriteString(media.Body)
					i += urlLen
					lastItem = i
					continue
				case ERawExternal:
					sb.WriteString(media.Body)
					i += urlLen
					lastItem = i
					externalHead = true
					continue
				case ENone:
					// Do nothing
				// TODO: Add support for media plugins
				default:
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
				sb.WriteString(media.URL)
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
	return msg, externalHead
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
		ch := data[i]
		if ch != '\\' && ch != '_' && ch != '?' && ch != '&' && ch != '=' && ch != '@' && ch != '#' && ch != ']' && !(ch > 44 && ch < 60) && !(ch > 64 && ch < 92) && !(ch > 96 && ch < 123) { // 57 is 9, 58 is :, 59 is ;, 90 is Z, 91 is [
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
		ch := data[i]
		if ch != '\\' && ch != '_' && ch != '?' && ch != '&' && ch != '=' && ch != '@' && ch != '#' && ch != ']' && !(ch > 44 && ch < 60) && !(ch > 64 && ch < 92) && !(ch > 96 && ch < 123) { // 57 is 9, 58 is :, 59 is ;, 90 is Z, 91 is [
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
		ch := data[i]
		if ch != '\\' && ch != '_' && ch != '?' && ch != '&' && ch != '=' && ch != '@' && ch != '#' && ch != ']' && !(ch > 44 && ch < 60) && !(ch > 64 && ch < 92) && !(ch > 96 && ch < 123) { // 57 is 9, 58 is :, 59 is ;, 90 is Z, 91 is [
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
		ch := data[i] //char
		if ch < 33 {  // space and invisibles
			//fmt.Println("e2:",i)
			return i, i != f
		} else if ch != '\\' && ch != '_' && ch != '?' && ch != '&' && ch != '=' && ch != '@' && ch != '#' && ch != ']' && !(ch > 44 && ch < 60) && !(ch > 64 && ch < 92) && !(ch > 96 && ch < 123) { // 57 is 9, 58 is :, 59 is ;, 90 is Z, 91 is [
			//log.Print("Bad Character: ", ch)
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
		ch := data[i]
		if ch != '\\' && ch != '_' && ch != '?' && ch != '&' && ch != '=' && ch != '@' && ch != '#' && ch != ']' && !(ch > 44 && ch < 60) && !(ch > 64 && ch < 91) && !(ch > 96 && ch < 123) { // 57 is 9, 58 is :, 59 is ;, 90 is Z, 91 is [
			//log.Print("Bad Character: ", ch)
			return i
		}
	}
	//log.Print("Data Length: ",len(data))
	return len(data)
}

type MediaEmbed struct {
	//Type string //image
	Type int
	URL  string
	FURL string
	Body string

	Trusted bool // samesite urls
}

const (
	ENone = iota
	ERaw
	ERawExternal
	EImage
	AImage
	AVideo
	AAudio
	AText
	AOther
)

var LastEmbedID = AOther

// TODO: Write a test for this
func parseMediaString(data string, settings *ParseSettings) (media MediaEmbed, ok bool) {
	if !validateURLString(data) {
		return media, false
	}
	uurl, err := url.Parse(data)
	if err != nil {
		return media, false
	}
	host := uurl.Hostname()
	scheme := uurl.Scheme
	port := uurl.Port()
	query, err := url.ParseQuery(uurl.RawQuery)
	if err != nil {
		return media, false
	}
	//fmt.Println("host:", host)
	//log.Print("Site.URL:",Site.URL)

	samesite := host == "localhost" || host == "127.0.0.1" || host == "::1" || host == Site.URL
	if samesite {
		host = strings.Split(Site.URL, ":")[0]
		// ?- Test this as I'm not sure it'll do what it should. If someone's running SSL on port 80 or non-SSL on port 443 then... Well... They're in far worse trouble than this...
		port = Site.Port
		if Config.SslSchema {
			scheme = "https"
		}
	}
	if scheme != "" {
		scheme += ":"
	}
	media.Trusted = samesite

	path := uurl.EscapedPath()
	//fmt.Println("path:", path)
	pathFrags := strings.Split(path, "/")
	if len(pathFrags) >= 2 {
		if samesite && pathFrags[1] == "attachs" && (scheme == "http:" || scheme == "https:") {
			var sport string
			// ? - Assumes the sysadmin hasn't mixed up the two standard ports
			if port != "443" && port != "80" && port != "" {
				sport = ":" + port
			}
			media.URL = scheme + "//" + host + sport + path
			ext := strings.TrimPrefix(filepath.Ext(path), ".")
			if len(ext) == 0 {
				// TODO: Write a unit test for this
				return media, false
			}
			switch {
			case ImageFileExts.Contains(ext):
				media.Type = AImage
			case WebVideoFileExts.Contains(ext):
				media.Type = AVideo
			case WebAudioFileExts.Contains(ext):
				media.Type = AAudio
			case TextFileExts.Contains(ext):
				media.Type = AText
			default:
				media.Type = AOther
			}
			return media, true
		}
	}

	//fmt.Printf("settings.NoEmbed: %+v\n", settings.NoEmbed)
	//settings.NoEmbed = false
	if !settings.NoEmbed {
		// ? - I don't think this hostname will hit every YT domain
		// TODO: Make this a more customisable handler rather than hard-coding it in here
		ytInvalid := func(v string) bool {
			for _, ch := range v {
				if !((ch > 47 && ch < 58) || (ch > 64 && ch < 91) || (ch > 96 && ch < 123) || ch == '-' || ch == '_') {
					var sport string
					if port != "443" && port != "80" && port != "" {
						sport = ":" + port
					}
					var q string
					if len(uurl.RawQuery) > 0 {
						q = "?" + uurl.RawQuery
					}
					var frag string
					if len(uurl.Fragment) > 0 {
						frag = "#" + uurl.Fragment
					}
					media.FURL = host + sport + path + q + frag
					media.URL = scheme + "//" + media.FURL
					//fmt.Printf("ytInvalid true: %+v\n",v)
					return true
				}
			}
			return false
		}
		ytInvalid2 := func(t string) bool {
			for _, ch := range t {
				if !((ch > 47 && ch < 58) || ch == 'h' || ch == 'm' || ch == 's') {
					//fmt.Printf("ytInvalid2 true: %+v\n",t)
					return true
				}
			}
			return false
		}
		if strings.HasSuffix(host, ".youtube.com") && path == "/watch" {
			video, ok := query["v"]
			if ok && len(video) >= 1 && video[0] != "" {
				v := video[0]
				if ytInvalid(v) {
					return media, true
				}
				var t, t2 string
				tt, ok := query["t"]
				if ok && len(tt) >= 1 {
					t, t2 = tt[0], tt[0]
				}
				media.Type = ERawExternal
				if t != "" && !ytInvalid2(t) {
					s, m, h := parseDuration(t2)
					calc := s + (m * 60) + (h * 60 * 60)
					if calc > 0 {
						t = "&t=" + t
						t2 = "?start=" + strconv.Itoa(calc)
					} else {
						t, t2 = "", ""
					}
				}
				l := "https://" + host + path + "?v=" + v + t
				media.Body = "<iframe class='postIframe'src='https://www.youtube-nocookie.com/embed/" + v + t2 + "'frameborder=0 allowfullscreen></iframe><noscript><a href='" + l + "'>" + l + "</a></noscript>"
				return media, true
			}
		} else if host == "youtu.be" {
			v := strings.TrimPrefix(path, "/")
			if ytInvalid(v) {
				return media, true
			}
			l := "https://youtu.be/" + v
			media.Type = ERawExternal
			media.Body = "<iframe class='postIframe'src='https://www.youtube-nocookie.com/embed/" + v + "'frameborder=0 allowfullscreen></iframe><noscript><a href='" + l + "'>" + l + "</a></noscript>"
			return media, true
		} else if strings.HasPrefix(host, "www.nicovideo.jp") && strings.HasPrefix(path, "/watch/sm") {
			vid, err := strconv.ParseInt(strings.TrimPrefix(path, "/watch/sm"), 10, 64)
			if err == nil {
				var sport string
				if port != "443" && port != "80" && port != "" {
					sport = ":" + port
				}
				media.Type = ERawExternal
				sm := strconv.FormatInt(vid, 10)
				l := "https://" + host + sport + path
				media.Body = "<iframe class='postIframe'src='https://embed.nicovideo.jp/watch/sm" + sm + "?jsapi=1&amp;playerId=1'frameborder=0 allowfullscreen></iframe><noscript><a href='" + l + "'>" + l + "</a></noscript>"
				return media, true
			}
		}

		if lastFrag := pathFrags[len(pathFrags)-1]; lastFrag != "" {
			// TODO: Write a function for getting the file extension of a string
			ext := strings.TrimPrefix(filepath.Ext(lastFrag), ".")
			if len(ext) != 0 {
				if ImageFileExts.Contains(ext) {
					media.Type = EImage
					var sport string
					if port != "443" && port != "80" && port != "" {
						sport = ":" + port
					}
					media.URL = scheme + "//" + host + sport + path
					return media, true
				}
				// TODO: Support external videos
			}
		}
	}

	var sport string
	if port != "443" && port != "80" && port != "" {
		sport = ":" + port
	}
	var q string
	if len(uurl.RawQuery) > 0 {
		q = "?" + uurl.RawQuery
	}
	var frag string
	if len(uurl.Fragment) > 0 {
		frag = "#" + uurl.Fragment
	}
	media.FURL = host + sport + path + q + frag
	media.URL = scheme + "//" + media.FURL

	return media, true
}

func parseDuration(dur string) (s, m, h int) {
	var ibuf []byte
	for _, ch := range dur {
		switch {
		case ch > 47 && ch < 58:
			ibuf = append(ibuf, byte(ch))
		case ch == 'h':
			h, _ = strconv.Atoi(string(ibuf))
			ibuf = ibuf[:0]
		case ch == 'm':
			m, _ = strconv.Atoi(string(ibuf))
			ibuf = ibuf[:0]
		case ch == 's':
			s, _ = strconv.Atoi(string(ibuf))
			ibuf = ibuf[:0]
		}
	}
	// Stop accidental uses of timestamps
	if h == 0 && m == 0 && s < 2 {
		s = 0
	}
	return s, m, h
}

// TODO: Write a test for this
func CoerceIntString(data string) (res, length int) {
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
func Paginate(currentPage, lastPage, maxPages int) (out []int) {
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
func PageOffset(count, page, perPage int) (int, int, int) {
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
func LastPage(count, perPage int) int {
	return (count / perPage) + 1
}
