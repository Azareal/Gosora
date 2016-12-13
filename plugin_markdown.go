package main

import "regexp"

var bold_italic *regexp.Regexp
var bold *regexp.Regexp
var italic *regexp.Regexp

func init() {
	plugins["markdown"] = Plugin{"markdown","Markdown","Azareal","http://github.com/Azareal","",false,"",init_markdown,nil,deactivate_markdown}
}

func init_markdown() {
	add_hook("parse_assign", markdown_parse)
	bold_italic = regexp.MustCompile(`\*\*\*(.*)\*\*\*`)
	bold = regexp.MustCompile(`\*\*(.*)\*\*`)
	italic = regexp.MustCompile(`\*(.*)\*`)
}

func deactivate_markdown() {
	remove_hook("parse_assign")
}

func markdown_parse(data interface{}) interface{} {
	msg := data.(string)
	msg = bold_italic.ReplaceAllString(msg,"<i><b>$1</b></i>")
	msg = bold.ReplaceAllString(msg,"<b>$1</b>")
	msg = italic.ReplaceAllString(msg,"<i>$1</i>")
	return msg
}