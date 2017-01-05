package main
import "regexp"

var markdown_bold_italic *regexp.Regexp
var markdown_bold *regexp.Regexp
var markdown_italic *regexp.Regexp

func init() {
	plugins["markdown"] = NewPlugin("markdown","Markdown","Azareal","http://github.com/Azareal","","","",init_markdown,nil,deactivate_markdown)
}

func init_markdown() {
	plugins["markdown"].AddHook("parse_assign", markdown_parse)
	markdown_bold_italic = regexp.MustCompile(`\*\*\*(.*)\*\*\*`)
	markdown_bold = regexp.MustCompile(`\*\*(.*)\*\*`)
	markdown_italic = regexp.MustCompile(`\*(.*)\*`)
}

func deactivate_markdown() {
	plugins["markdown"].RemoveHook("parse_assign", markdown_parse)
}

func markdown_parse(data interface{}) interface{} {
	msg := data.(string)
	msg = markdown_bold_italic.ReplaceAllString(msg,"<i><b>$1</b></i>")
	msg = markdown_bold.ReplaceAllString(msg,"<b>$1</b>")
	msg = markdown_italic.ReplaceAllString(msg,"<i>$1</i>")
	return msg
}