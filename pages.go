package main

//import "fmt"
import "sync"
import "bytes"
import "strings"
import "strconv"
import "regexp"
import "html/template"

type HeaderVars struct
{
	NoticeList []string
	Scripts []string
	Stylesheets []string
	Widgets PageWidgets
	ExtData ExtData
}

type PageWidgets struct
{
	LeftSidebar template.HTML
	RightSidebar template.HTML
}

/*type UnsafeExtData struct
{
	items map[string]interface{} // Key: pluginname
}*/

type ExtData struct
{
	items map[string]interface{} // Key: pluginname
	sync.RWMutex
}

type Page struct
{
	Title string
	CurrentUser User
	Header HeaderVars
	ItemList []interface{}
	Something interface{}
}

type TopicPage struct
{
	Title string
	CurrentUser User
	Header HeaderVars
	ItemList []Reply
	Topic TopicUser
	Page int
	LastPage int
	ExtData ExtData
}

type TopicsPage struct
{
	Title string
	CurrentUser User
	Header HeaderVars
	ItemList []TopicsRow
	ExtData ExtData
}

type ForumPage struct
{
	Title string
	CurrentUser User
	Header HeaderVars
	ItemList []TopicUser
	Forum Forum
	Page int
	LastPage int
	ExtData ExtData
}

type ForumsPage struct
{
	Title string
	CurrentUser User
	Header HeaderVars
	ItemList []Forum
	ExtData ExtData
}

type ProfilePage struct
{
	Title string
	CurrentUser User
	Header HeaderVars
	ItemList []Reply
	ProfileOwner User
	ExtData ExtData
}

type CreateTopicPage struct
{
	Title string
	CurrentUser User
	Header HeaderVars
	ItemList []Forum
	FID int
	ExtData ExtData
}

type GridElement struct
{
	ID string
	Body string
	Order int // For future use
	Class string
	Background string
	TextColour string
	Note string
}

type PanelDashboardPage struct
{
	Title string
	CurrentUser User
	Header HeaderVars
	GridItems []GridElement
	ExtData ExtData
}

type ThemesPage struct
{
	Title string
	CurrentUser User
	Header HeaderVars
	PrimaryThemes []Theme
	VariantThemes []Theme
	ExtData ExtData
}

type EditGroupPage struct
{
	Title string
	CurrentUser User
	Header HeaderVars
	ID int
	Name string
	Tag string
	Rank string
	DisableRank bool
	ExtData ExtData
}

type GroupForumPermPreset struct
{
	Group Group
	Preset string
}

type EditForumPage struct
{
	Title string
	CurrentUser User
	Header HeaderVars
	ID int
	Name string
	Desc string
	Active bool
	Preset string
	Groups []GroupForumPermPreset
	ExtData ExtData
}

type NameLangPair struct
{
	Name string
	LangStr string
}

type NameLangToggle struct
{
	Name string
	LangStr string
	Toggle bool
}

type EditGroupPermsPage struct
{
	Title string
	CurrentUser User
	Header HeaderVars
	ID int
	Name string
	LocalPerms []NameLangToggle
	GlobalPerms []NameLangToggle
	ExtData ExtData
}

type Log struct {
	Action template.HTML
	IPAddress string
	DoneAt string
}

type LogsPage struct
{
	Title string
	CurrentUser User
	Header HeaderVars
	Logs []Log
	ExtData ExtData
}

type PageSimple struct
{
	Title string
	Something interface{}
}

type AreYouSure struct
{
	URL string
	Message string
}

var space_gap []byte = []byte("          ")
var http_prot_b []byte = []byte("http://")
var invalid_url []byte = []byte("<span style='color: red;'>[Invalid URL]</span>")
var invalid_topic []byte = []byte("<span style='color: red;'>[Invalid Topic]</span>")
var invalid_profile []byte = []byte("<span style='color: red;'>[Invalid Profile]</span>")
var invalid_forum []byte = []byte("<span style='color: red;'>[Invalid Forum]</span>")
var url_open []byte = []byte("<a href='")
var url_open2 []byte = []byte("'>")
var bytes_singlequote []byte = []byte("'")
var bytes_greaterthan []byte = []byte(">")
var url_mention []byte = []byte(" class='mention'")
var url_close []byte = []byte("</a>")
var urlpattern string = `(?s)([ {1}])((http|https|ftp|mailto)*)(:{??)\/\/([\.a-zA-Z\/]+)([ {1}])`
var url_reg *regexp.Regexp

func init() {
	url_reg = regexp.MustCompile(urlpattern)
}

func shortcode_to_unicode(msg string) string {
	//re := regexp.MustCompile(":(.):")
	msg = strings.Replace(msg,":grinning:","😀",-1)
	msg = strings.Replace(msg,":grin:","😁",-1)
	msg = strings.Replace(msg,":joy:","😂",-1)
	msg = strings.Replace(msg,":rofl:","🤣",-1)
	msg = strings.Replace(msg,":smiley:","😃",-1)
	msg = strings.Replace(msg,":smile:","😄",-1)
	msg = strings.Replace(msg,":sweat_smile:","😅",-1)
	msg = strings.Replace(msg,":laughing:","😆",-1)
	msg = strings.Replace(msg,":satisfied:","😆",-1)
	msg = strings.Replace(msg,":wink:","😉",-1)
	msg = strings.Replace(msg,":blush:","😊",-1)
	msg = strings.Replace(msg,":yum:","😋",-1)
	msg = strings.Replace(msg,":sunglasses:","😎",-1)
	msg = strings.Replace(msg,":heart_eyes:","😍",-1)
	msg = strings.Replace(msg,":kissing_heart:","😘",-1)
	msg = strings.Replace(msg,":kissing:","😗",-1)
	msg = strings.Replace(msg,":kissing_smiling_eyes:","😙",-1)
	msg = strings.Replace(msg,":kissing_closed_eyes:","😚",-1)
	msg = strings.Replace(msg,":relaxed:","☺️",-1)
	msg = strings.Replace(msg,":slight_smile:","🙂",-1)
	msg = strings.Replace(msg,":hugging:","🤗",-1)
	msg = strings.Replace(msg,":thinking:","🤔",-1)
	msg = strings.Replace(msg,":neutral_face:","😐",-1)
	msg = strings.Replace(msg,":expressionless:","😑",-1)
	msg = strings.Replace(msg,":no_mouth:","😶",-1)
	msg = strings.Replace(msg,":rolling_eyes:","🙄",-1)
	msg = strings.Replace(msg,":smirk:","😏",-1)
	msg = strings.Replace(msg,":persevere:","😣",-1)
	msg = strings.Replace(msg,":disappointed_relieved:","😥",-1)
	msg = strings.Replace(msg,":open_mouth:","😮",-1)
	msg = strings.Replace(msg,":zipper_mouth:","🤐",-1)
	msg = strings.Replace(msg,":hushed:","😯",-1)
	msg = strings.Replace(msg,":sleepy:","😪",-1)
	msg = strings.Replace(msg,":tired_face:","😫",-1)
	msg = strings.Replace(msg,":sleeping:","😴",-1)
	msg = strings.Replace(msg,":relieved:","😌",-1)
	msg = strings.Replace(msg,":nerd:","🤓",-1)
	msg = strings.Replace(msg,":stuck_out_tongue:","😛",-1)
	msg = strings.Replace(msg,":worried:","😟",-1)
	msg = strings.Replace(msg,":drooling_face:","🤤",-1)
	msg = strings.Replace(msg,":disappointed:","😞",-1)
	msg = strings.Replace(msg,":astonished:","😲",-1)
	msg = strings.Replace(msg,":slight_frown:","🙁",-1)
	msg = strings.Replace(msg,":skull_crossbones:","☠️",-1)
	msg = strings.Replace(msg,":skull:","💀",-1)
	msg = strings.Replace(msg,":point_up:","☝️",-1)
	msg = strings.Replace(msg,":v:","✌️️",-1)
	msg = strings.Replace(msg,":writing_hand:","✍️",-1)
	msg = strings.Replace(msg,":heart:","❤️️",-1)
	msg = strings.Replace(msg,":heart_exclamation:","❣️",-1)
	msg = strings.Replace(msg,":hotsprings:","♨️",-1)
	msg = strings.Replace(msg,":airplane:","✈️️",-1)
	msg = strings.Replace(msg,":hourglass:","⌛",-1)
	msg = strings.Replace(msg,":watch:","⌚",-1)
	msg = strings.Replace(msg,":comet:","☄️",-1)
	msg = strings.Replace(msg,":snowflake:","❄️",-1)
	msg = strings.Replace(msg,":cloud:","☁️",-1)
	msg = strings.Replace(msg,":sunny:","☀️",-1)
	msg = strings.Replace(msg,":spades:","♠️",-1)
	msg = strings.Replace(msg,":hearts:","♥️️",-1)
	msg = strings.Replace(msg,":diamonds:","♦️",-1)
	msg = strings.Replace(msg,":clubs:","♣️",-1)
	msg = strings.Replace(msg,":phone:","☎️",-1)
	msg = strings.Replace(msg,":telephone:","☎️",-1)
	msg = strings.Replace(msg,":biohazard:","☣️",-1)
	msg = strings.Replace(msg,":radioactive:","☢️",-1)
	msg = strings.Replace(msg,":scissors:","✂️",-1)
	msg = strings.Replace(msg,":arrow_upper_right:","↗️",-1)
	msg = strings.Replace(msg,":arrow_right:","➡️",-1)
	msg = strings.Replace(msg,":arrow_lower_right:","↘️",-1)
	msg = strings.Replace(msg,":arrow_lower_left:","↙️",-1)
	msg = strings.Replace(msg,":arrow_upper_left:","↖️",-1)
	msg = strings.Replace(msg,":arrow_up_down:","↕️",-1)
	msg = strings.Replace(msg,":left_right_arrow:","↔️",-1)
	msg = strings.Replace(msg,":leftwards_arrow_with_hook:","↩️",-1)
	msg = strings.Replace(msg,":arrow_right_hook:","↪️",-1)
	msg = strings.Replace(msg,":arrow_forward:","▶️",-1)
	msg = strings.Replace(msg,":arrow_backward:","◀️",-1)
	msg = strings.Replace(msg,":female:","♀️",-1)
	msg = strings.Replace(msg,":male:","♂️",-1)
	msg = strings.Replace(msg,":ballot_box_with_check:","☑️",-1)
	msg = strings.Replace(msg,":heavy_check_mark:","✔️️",-1)
	msg = strings.Replace(msg,":heavy_multiplication_x:","✖️",-1)
	msg = strings.Replace(msg,":pisces:","♓",-1)
	msg = strings.Replace(msg,":aquarius:","♒",-1)
	msg = strings.Replace(msg,":capricorn:","♑",-1)
	msg = strings.Replace(msg,":sagittarius:","♐",-1)
	msg = strings.Replace(msg,":scorpius:","♏",-1)
	msg = strings.Replace(msg,":libra:","♎",-1)
	msg = strings.Replace(msg,":virgo:","♍",-1)
	msg = strings.Replace(msg,":leo:","♌",-1)
	msg = strings.Replace(msg,":cancer:","♋",-1)
	msg = strings.Replace(msg,":gemini:","♊",-1)
	msg = strings.Replace(msg,":taurus:","♉",-1)
	msg = strings.Replace(msg,":aries:","♈",-1)
	msg = strings.Replace(msg,":peace:","☮️",-1)
	msg = strings.Replace(msg,":eight_spoked_asterisk:","✳️",-1)
	msg = strings.Replace(msg,":eight_pointed_black_star:","✴️",-1)
	msg = strings.Replace(msg,":snowman2:","☃️",-1)
	msg = strings.Replace(msg,":umbrella2:","☂️",-1)
	msg = strings.Replace(msg,":pencil2:","✏️",-1)
	msg = strings.Replace(msg,":black_nib:","✒️",-1)
	msg = strings.Replace(msg,":email:","✉️",-1)
	msg = strings.Replace(msg,":envelope:","✉️",-1)
	msg = strings.Replace(msg,":keyboard:","⌨️",-1)
	msg = strings.Replace(msg,":white_small_square:","▫️",-1)
	msg = strings.Replace(msg,":black_small_square:","▪️",-1)
	msg = strings.Replace(msg,":secret:","㊙️",-1)
	msg = strings.Replace(msg,":congratulations:","㊗️",-1)
	msg = strings.Replace(msg,":m:","Ⓜ️",-1)
	msg = strings.Replace(msg,":tm:","™️️",-1)
	msg = strings.Replace(msg,":registered:","®️",-1)
	msg = strings.Replace(msg,":copyright:","©️",-1)
	msg = strings.Replace(msg,":wavy_dash:","〰️",-1)
	msg = strings.Replace(msg,":bangbang:","‼️",-1)
	msg = strings.Replace(msg,":sparkle:","❇️",-1)
	msg = strings.Replace(msg,":star_of_david:","✡️",-1)
	msg = strings.Replace(msg,":wheel_of_dharma:","☸️",-1)
	msg = strings.Replace(msg,":yin_yang:","☯️",-1)
	msg = strings.Replace(msg,":cross:","✝️",-1)
	msg = strings.Replace(msg,":orthodox_cross:","☦️",-1)
	msg = strings.Replace(msg,":star_and_crescent:","☪️",-1)
	msg = strings.Replace(msg,":frowning2:","☹️",-1)
	msg = strings.Replace(msg,":information_source:","ℹ️",-1)
	msg = strings.Replace(msg,":interrobang:","⁉️",-1)

	return msg
}

func preparse_message(msg string) string {
	if sshooks["preparse_preassign"] != nil {
		msg = run_sshook("preparse_preassign", msg)
	}
	return shortcode_to_unicode(msg)
}

func parse_message(msg string/*, user User*/) string {
	msg = strings.Replace(msg,":)","😀",-1)
	msg = strings.Replace(msg,":(","😞",-1)
	msg = strings.Replace(msg,":D","😃",-1)
	msg = strings.Replace(msg,":P","😛",-1)
	msg = strings.Replace(msg,":O","😲",-1)
	msg = strings.Replace(msg,":p","😛",-1)
	msg = strings.Replace(msg,":o","😲",-1)
	//msg = url_reg.ReplaceAllString(msg,"<a href=\"$2$3//$4\" rel=\"nofollow\">$2$3//$4</a>")

	// Search for URLs, mentions and hashlinks in the messages...
	//fmt.Println("Parser Loop!")
	var msgbytes = []byte(msg)
	var outbytes []byte
	msgbytes = append(msgbytes,space_gap...)
	//fmt.Println(`"`+string(msgbytes)+`"`)
	lastItem := 0
	i := 0
	for ; len(msgbytes) > (i + 1); i++ {
		//fmt.Println("Index:")
		//fmt.Println(i)
		//fmt.Println("Index Item:")
		//fmt.Println(msgbytes[i])
		//if msgbytes[i] == 10 {
		//	fmt.Println("NEWLINE")
		//} else if msgbytes[i] == 32 {
		//	fmt.Println("SPACE")
		//} else {
		//	fmt.Println(string(msgbytes[i]))
		//}
		//fmt.Println("End Index")
		if (i==0 && (msgbytes[0] > 32)) || ((msgbytes[i] < 33) && (msgbytes[i + 1] > 32)) {
			//fmt.Println("IN")
			//fmt.Println(msgbytes[i])
			if (i != 0) || msgbytes[i] < 33 {
				i++
			}

			if msgbytes[i]=='#' {
				//fmt.Println("IN #")
				if bytes.Equal(msgbytes[i+1:i+5],[]byte("tid-")) {
					outbytes = append(outbytes,msgbytes[lastItem:i]...)
					i += 5
					start := i
					tid, int_len := coerce_int_bytes(msgbytes[start:])
					i += int_len

					topic, err := topics.CascadeGet(tid)
					if err != nil || !fstore.Exists(topic.ParentID) {
						outbytes = append(outbytes,invalid_topic...)
						lastItem = i
						continue
					}

					outbytes = append(outbytes, url_open...)
					var url_bit []byte = []byte(build_topic_url("",tid))
					outbytes = append(outbytes, url_bit...)
					outbytes = append(outbytes, url_open2...)
					var tid_bit []byte = []byte("#tid-" + strconv.Itoa(tid))
					outbytes = append(outbytes, tid_bit...)
					outbytes = append(outbytes, url_close...)
					lastItem = i

					//fmt.Println(string(msgbytes))
					//fmt.Println(msgbytes)
					//fmt.Println(msgbytes[lastItem - 1])
					//fmt.Println(lastItem - 1)
					//fmt.Println(msgbytes[lastItem])
					//fmt.Println(lastItem)
				} else if bytes.Equal(msgbytes[i+1:i+5],[]byte("rid-")) {
					outbytes = append(outbytes,msgbytes[lastItem:i]...)
					i += 5
					start := i
					rid, int_len := coerce_int_bytes(msgbytes[start:])
					i += int_len

					topic, err := get_topic_by_reply(rid)
					if err != nil || !fstore.Exists(topic.ParentID) {
						outbytes = append(outbytes,invalid_topic...)
						lastItem = i
						continue
					}

					outbytes = append(outbytes, url_open...)
					var url_bit []byte = []byte(build_topic_url("",topic.ID))
					outbytes = append(outbytes, url_bit...)
					outbytes = append(outbytes, url_open2...)
					var rid_bit []byte = []byte("#rid-" + strconv.Itoa(rid))
					outbytes = append(outbytes, rid_bit...)
					outbytes = append(outbytes, url_close...)
					lastItem = i
				} else if bytes.Equal(msgbytes[i+1:i+5],[]byte("fid-")) {
					outbytes = append(outbytes,msgbytes[lastItem:i]...)
					i += 5
					start := i
					fid, int_len := coerce_int_bytes(msgbytes[start:])
					i += int_len

					if !fstore.Exists(fid) {
						outbytes = append(outbytes,invalid_forum...)
						lastItem = i
						continue
					}

					outbytes = append(outbytes, url_open...)
					var url_bit []byte = []byte(build_forum_url("",fid))
					outbytes = append(outbytes, url_bit...)
					outbytes = append(outbytes, url_open2...)
					var fid_bit []byte = []byte("#fid-" + strconv.Itoa(fid))
					outbytes = append(outbytes, fid_bit...)
					outbytes = append(outbytes, url_close...)
					lastItem = i
				} else {
					// TO-DO: Forum Shortcode Link
				}
			} else if msgbytes[i]=='@' {
				//fmt.Println("IN @")
				outbytes = append(outbytes,msgbytes[lastItem:i]...)
				i++
				start := i
				uid, int_len := coerce_int_bytes(msgbytes[start:])
				i += int_len

				menUser, err := users.CascadeGet(uid)
				if err != nil {
					outbytes = append(outbytes,invalid_profile...)
					lastItem = i
					continue
				}

				outbytes = append(outbytes, url_open...)
				var url_bit []byte = []byte(build_profile_url(menUser.Slug,uid))
				outbytes = append(outbytes, url_bit...)
				outbytes = append(outbytes, bytes_singlequote...)
				outbytes = append(outbytes, url_mention...)
				outbytes = append(outbytes, bytes_greaterthan...)
				var uid_bit []byte = []byte("@" + menUser.Name)
				outbytes = append(outbytes, uid_bit...)
				outbytes = append(outbytes, url_close...)
				lastItem = i

				//fmt.Println(string(msgbytes))
				//fmt.Println(msgbytes)
				//fmt.Println(msgbytes[lastItem - 1])
				//fmt.Println("lastItem - 1",lastItem - 1)
				//fmt.Println("msgbytes[lastItem]",msgbytes[lastItem])
				//fmt.Println("lastItem",lastItem)
			} else if msgbytes[i]=='h' || msgbytes[i]=='f' || msgbytes[i]=='g' {
				//fmt.Println("IN hfg")
				if msgbytes[i + 1]=='t' && msgbytes[i + 2]=='t' && msgbytes[i + 3]=='p' {
					if msgbytes[i + 4] == 's' && msgbytes[i + 5] == ':' && msgbytes[i + 6] == '/' && msgbytes[i + 7] == '/' {
						// Do nothing
					} else if msgbytes[i + 4] == ':' && msgbytes[i + 5] == '/' && msgbytes[i + 6] == '/' {
						// Do nothing
					} else {
						continue
					}
				} else if msgbytes[i + 1] == 't' && msgbytes[i + 2] == 'p' && msgbytes[i + 3] == ':' && msgbytes[i + 4] == '/' && msgbytes[i + 5] == '/' {
					// Do nothing
				} else if msgbytes[i + 1] == 'i' && msgbytes[i + 2] == 't' && msgbytes[i + 3] == ':' && msgbytes[i + 4] == '/' && msgbytes[i + 5] == '/' {
					// Do nothing
				} else {
					continue
				}

				outbytes = append(outbytes,msgbytes[lastItem:i]...)
				url_len := partial_url_bytes_len(msgbytes[i:])
				if msgbytes[i + url_len] != ' ' && msgbytes[i + url_len] != 10 {
					outbytes = append(outbytes,invalid_url...)
					i += url_len
					continue
				}
				outbytes = append(outbytes, url_open...)
				outbytes = append(outbytes, msgbytes[i:i + url_len]...)
				outbytes = append(outbytes, url_open2...)
				outbytes = append(outbytes, msgbytes[i:i + url_len]...)
				outbytes = append(outbytes, url_close...)
				i += url_len
				lastItem = i
			}
		}
	}

	if lastItem != i && len(outbytes) != 0 {
		//fmt.Println("lastItem:",msgbytes[lastItem])
		//fmt.Println("lastItem index:")
		//fmt.Println(lastItem)
		//fmt.Println("i:")
		//fmt.Println(i)
		//fmt.Println("lastItem to end:")
		//fmt.Println(msgbytes[lastItem:])
		//fmt.Println("-----")
		calclen := len(msgbytes) - 10
		if calclen <= lastItem {
			calclen = lastItem
		}
		outbytes = append(outbytes, msgbytes[lastItem:calclen]...)
		msg = string(outbytes)
	}
	//fmt.Println(`"`+string(outbytes)+`"`)
	//fmt.Println(`"`+msg+`"`)

	msg = strings.Replace(msg,"\n","<br>",-1)
	if sshooks["parse_assign"] != nil {
		msg = run_sshook("parse_assign", msg)
	}
	return msg
}

func regex_parse_message(msg string) string {
	msg = strings.Replace(msg,":)","😀",-1)
	msg = strings.Replace(msg,":D","😃",-1)
	msg = strings.Replace(msg,":P","😛",-1)
	msg = url_reg.ReplaceAllString(msg,"<a href=\"$2$3//$4\" rel=\"nofollow\">$2$3//$4</a>")
	msg = strings.Replace(msg,"\n","<br>",-1)
	if sshooks["parse_assign"] != nil {
		msg = run_sshook("parse_assign", msg)
	}
	return msg
}

// 6, 7, 8, 6, 7
// ftp://, http://, https:// git://, mailto: (not a URL, just here for length comparison purposes)
func validate_url_bytes(data []byte) bool {
	datalen := len(data)
	i := 0

	if datalen >= 6 {
		if bytes.Equal(data[0:6],[]byte("ftp://")) || bytes.Equal(data[0:6],[]byte("git://")) {
			i = 6
		} else if datalen >= 7 && bytes.Equal(data[0:7],http_prot_b) {
			i = 7
		} else if datalen >= 8 && bytes.Equal(data[0:8],[]byte("https://")) {
			i = 8
		}
	}

	for ;datalen > i; i++ {
		if data[i] != '\\' && data[i] != '_' && !(data[i] > 44 && data[i] < 58) && !(data[i] > 64 && data[i] < 91) && !(data[i] > 96 && data[i] < 123) {
			return false
		}
	}
	return true
}

func validated_url_bytes(data []byte) (url []byte) {
	datalen := len(data)
	i := 0

	if datalen >= 6 {
		if bytes.Equal(data[0:6],[]byte("ftp://")) || bytes.Equal(data[0:6],[]byte("git://")) {
			i = 6
		} else if datalen >= 7 && bytes.Equal(data[0:7],http_prot_b) {
			i = 7
		} else if datalen >= 8 && bytes.Equal(data[0:8],[]byte("https://")) {
			i = 8
		}
	}

	for ;datalen > i; i++ {
		if data[i] != '\\' && data[i] != '_' && !(data[i] > 44 && data[i] < 58) && !(data[i] > 64 && data[i] < 91) && !(data[i] > 96 && data[i] < 123) {
			return invalid_url
		}
	}

	url = append(url, data...)
	return url
}

func partial_url_bytes(data []byte) (url []byte) {
	datalen := len(data)
	i := 0
	end := datalen - 1

	if datalen >= 6 {
		if bytes.Equal(data[0:6],[]byte("ftp://")) || bytes.Equal(data[0:6],[]byte("git://")) {
			i = 6
		} else if datalen >= 7 && bytes.Equal(data[0:7],http_prot_b) {
			i = 7
		} else if datalen >= 8 && bytes.Equal(data[0:8],[]byte("https://")) {
			i = 8
		}
	}

	for ;end >= i; i++ {
		if data[i] != '\\' && data[i] != '_' && !(data[i] > 44 && data[i] < 58) && !(data[i] > 64 && data[i] < 91) && !(data[i] > 96 && data[i] < 123) {
			end = i
		}
	}

	url = append(url, data[0:end]...)
	return url
}

func partial_url_bytes_len(data []byte) int {
	datalen := len(data)
	i := 0

	if datalen >= 6 {
		//fmt.Println(string(data[0:5]))
		if bytes.Equal(data[0:6],[]byte("ftp://")) || bytes.Equal(data[0:6],[]byte("git://")) {
			i = 6
		} else if datalen >= 7 && bytes.Equal(data[0:7],http_prot_b) {
			i = 7
		} else if datalen >= 8 && bytes.Equal(data[0:8],[]byte("https://")) {
			i = 8
		}
	}

	for ;datalen > i; i++ {
		if data[i] != '\\' && data[i] != '_' && !(data[i] > 44 && data[i] < 58) && !(data[i] > 64 && data[i] < 91) && !(data[i] > 96 && data[i] < 123) {
			//fmt.Println("Bad Character:")
			//fmt.Println(data[i])
			return i
		}
	}

	//fmt.Println("Data Length:")
	//fmt.Println(datalen)
	return datalen
}

func parse_media_bytes(data []byte) (protocol []byte, url []byte) {
	datalen := len(data)
	i := 0

	if datalen >= 6 {
		if bytes.Equal(data[0:6],[]byte("ftp://")) || bytes.Equal(data[0:6],[]byte("git://")) {
			i = 6
			protocol = data[0:2]
		} else if datalen >= 7 && bytes.Equal(data[0:7],http_prot_b) {
			i = 7
			protocol = []byte("http")
		} else if datalen >= 8 && bytes.Equal(data[0:8],[]byte("https://")) {
			i = 8
			protocol = []byte("https")
		}
	}

	for ;datalen > i; i++ {
		if data[i] != '\\' && data[i] != '_' && !(data[i] > 44 && data[i] < 58) && !(data[i] > 64 && data[i] < 91) && !(data[i] > 96 && data[i] < 123) {
			return []byte(""), invalid_url
		}
	}

	if len(protocol) == 0 {
		protocol = []byte("http")
	}
	return protocol, data[i:]
}

func coerce_int_bytes(data []byte) (res int, length int) {
	if !(data[0] > 47 && data[0] < 58) {
		return 0, 1
	}

	i := 0
	for ;len(data) > i; i++ {
		if !(data[i] > 47 && data[i] < 58) {
			conv, err := strconv.Atoi(string(data[0:i]))
			if err != nil {
				return 0, i
			}
			return conv, i
		}
	}

	conv, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, i
	}
	return conv, i
}
