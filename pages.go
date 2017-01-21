package main
import "strings"
import "regexp"

type Page struct
{
	Title string
	CurrentUser User
	NoticeList []string
	ItemList []interface{}
	Something interface{}
}

type TopicPage struct
{
	Title string
	CurrentUser User
	NoticeList []string
	ItemList []Reply
	Topic TopicUser
	Page int
	LastPage int
	ExtData interface{}
}

type TopicsPage struct
{
	Title string
	CurrentUser User
	NoticeList []string
	ItemList []TopicUser
	ExtData interface{}
}

type ForumPage struct
{
	Title string
	CurrentUser User
	NoticeList []string
	ItemList []TopicUser
	ExtData interface{}
}

type ForumsPage struct
{
	Title string
	CurrentUser User
	NoticeList []string
	ItemList []Forum
	ExtData interface{}
}

type ProfilePage struct
{
	Title string
	CurrentUser User
	NoticeList []string
	ItemList []Reply
	ProfileOwner User
	ExtData interface{}
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
	return msg
}

func preparse_message(msg string) string {
	if hooks["preparse_preassign"] != nil {
		out := run_hook("preparse_preassign", msg)
		msg = out.(string)
	}
	return shortcode_to_unicode(msg)
}

func parse_message(msg string) string {
	msg = strings.Replace(msg,":)","😀",-1)
	msg = strings.Replace(msg,":D","😃",-1)
	msg = strings.Replace(msg,":P","😛",-1)
	msg = url_reg.ReplaceAllString(msg,"<a href=\"$2$3//$4\" rel=\"nofollow\">$2$3//$4</a>")
	msg = strings.Replace(msg,"\n","<br>",-1)
	if hooks["parse_assign"] != nil {
		out := run_hook("parse_assign", msg)
		msg = out.(string)
	}
	return msg
}