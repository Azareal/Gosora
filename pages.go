package main
import "strings"
//import "regexp"

type Page struct
{
	Title string
	Name string
	CurrentUser User
	ItemList map[int]interface{}
	Something interface{}
}

type PageSimple struct
{
	Title string
	Name string
	Something interface{}
}

type AreYouSure struct
{
	URL string
	Message string
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
	msg = strings.Replace(msg,"\n","<br>",-1)
	if hooks["parse_assign"] != nil {
		out := run_hook("parse_assign", msg)
		msg = out.(string)
	}
	return msg
}