package main

import (
	"strconv"
	"testing"

	c "github.com/Azareal/Gosora/common"
	e "github.com/Azareal/Gosora/extend"
)

// go test -v

// TODO: Write a test for Hello World?

type MEPair struct {
	Msg     string
	Expects string
}

type MEPairList struct {
	Items []MEPair
}

func (l *MEPairList) Add(msg, expects string) {
	l.Items = append(l.Items, MEPair{msg, expects})
}

func TestBBCodeRender(t *testing.T) {
	//t.Skip()
	if e := e.InitBbcode(c.Plugins["bbcode"]); e != nil {
		t.Fatal(e)
	}

	var res string
	l := &MEPairList{nil}
	l.Add("", "")
	l.Add(" ", " ")
	l.Add("  ", "  ")
	l.Add("   ", "   ")
	l.Add("[b]", "<b></b>")
	l.Add("[b][/b]", "<b></b>")
	l.Add("hi", "hi")
	l.Add("😀", "😀")
	l.Add("[b]😀[/b]", "<b>😀</b>")
	l.Add("[b]😀😀😀[/b]", "<b>😀😀😀</b>")
	l.Add("[b]hi[/b]", "<b>hi</b>")
	l.Add("[u]hi[/u]", "<u>hi</u>")
	l.Add("[i]hi[/i]", "<i>hi</i>")
	l.Add("[s]hi[/s]", "<s>hi</s>")
	l.Add("[c]hi[/c]", "[c]hi[/c]")
	l.Add("[h1]hi", "[h1]hi")
	l.Add("[h1]hi[/h1]", "<h2>hi</h2>")
	if !testing.Short() {
		//l.Add("[b]hi[/i]", "[b]hi[/i]")
		//l.Add("[/b]hi[b]", "[/b]hi[b]")
		//l.Add("[/b]hi[/b]", "[/b]hi[/b]")
		//l.Add("[b][b]hi[/b]", "<b>hi</b>")
		//l.Add("[b][b]hi", "[b][b]hi")
		//l.Add("[b][b][b]hi", "[b][b][b]hi")
		//l.Add("[/b]hi", "[/b]hi")
	}
	l.Add("[spoiler]hi[/spoiler]", "<spoiler>hi</spoiler>")
	l.Add("[code]hi[/code]", "<span class='codequotes'>hi</span>")
	l.Add("[code][b]hi[/b][/code]", "<span class='codequotes'>[b]hi[/b]</span>")
	l.Add("[code][b]hi[/code][/b]", "<span class='codequotes'>[b]hi</span>[/b]")
	l.Add("[quote]hi[/quote]", "<blockquote>hi</blockquote>")
	l.Add("[quote][b]hi[/b][/quote]", "<blockquote><b>hi</b></blockquote>")
	l.Add("[quote][b]h[/b][/quote]", "<blockquote><b>h</b></blockquote>")
	l.Add("[quote][b][/b][/quote]", "<blockquote><b></b></blockquote>")
	l.Add("[url][/url]", "<a href=''></a>")
	l.Add("[url]https://github.com/Azareal/Gosora[/url]", "<a href='https://github.com/Azareal/Gosora'>https://github.com/Azareal/Gosora</a>")
	l.Add("[url]http://github.com/Azareal/Gosora[/url]", "<a href='http://github.com/Azareal/Gosora'>http://github.com/Azareal/Gosora</a>")
	l.Add("[url]//github.com/Azareal/Gosora[/url]", "<a href='//github.com/Azareal/Gosora'>//github.com/Azareal/Gosora</a>")
	l.Add("-你好-", "-你好-")
	l.Add("[i]-你好-[/i]", "<i>-你好-</i>") // TODO: More of these Unicode tests? Emoji, Chinese, etc.?

	t.Log("Testing BbcodeFullParse")
	for _, item := range l.Items {
		res = e.BbcodeFullParse(item.Msg)
		if res != item.Expects {
			t.Error("Testing string '" + item.Msg + "'")
			t.Error("Bad output:", "'"+res+"'")
			t.Error("Expected:", "'"+item.Expects+"'")
		}
	}

	f := func(msg, expects string) {
		t.Log("Testing string '" + msg + "'")
		res := e.BbcodeFullParse(msg)
		if res != expects {
			t.Error("Bad output:", "'"+res+"'")
			t.Error("Expected:", "'"+expects+"'")
		}
	}
	f("[rand][/rand]", "<red>[Invalid Number]</red>[rand][/rand]")
	f("[rand]-1[/rand]", "<red>[No Negative Numbers]</red>[rand]-1[/rand]")
	f("[rand]-01[/rand]", "<red>[No Negative Numbers]</red>[rand]-01[/rand]")
	f("[rand]NaN[/rand]", "<red>[Invalid Number]</red>[rand]NaN[/rand]")
	f("[rand]Inf[/rand]", "<red>[Invalid Number]</red>[rand]Inf[/rand]")
	f("[rand]+[/rand]", "<red>[Invalid Number]</red>[rand]+[/rand]")
	f("[rand]1+1[/rand]", "<red>[Invalid Number]</red>[rand]1+1[/rand]")

	msg := "[rand]1[/rand]"
	t.Log("Testing string '" + msg + "'")
	res = e.BbcodeFullParse(msg)
	conv, err := strconv.Atoi(res)
	if err != nil || (conv > 1 || conv < 0) {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected a number in the range 0-1")
	}

	msg = "[rand]0[/rand]"
	t.Log("Testing string '" + msg + "'")
	res = e.BbcodeFullParse(msg)
	conv, err = strconv.Atoi(res)
	if err != nil || conv != 0 {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected the number 0")
	}

	msg = "[rand]2147483647[/rand]" // Signed 32-bit MAX
	t.Log("Testing string '" + msg + "'")
	res = e.BbcodeFullParse(msg)
	conv, err = strconv.Atoi(res)
	if err != nil || (conv > 2147483647 || conv < 0) {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected a number between 0 and 2147483647")
	}

	msg = "[rand]9223372036854775807[/rand]" // Signed 64-bit MAX
	t.Log("Testing string '" + msg + "'")
	res = e.BbcodeFullParse(msg)
	conv, err = strconv.Atoi(res)
	if err != nil || (conv > 9223372036854775807 || conv < 0) {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected a number between 0 and 9223372036854775807")
	}

	// Note: conv is commented out in these two, as these numbers overflow int
	msg = "[rand]18446744073709551615[/rand]" // Unsigned 64-bit MAX
	t.Log("Testing string '" + msg + "'")
	res = e.BbcodeFullParse(msg)
	_, err = strconv.Atoi(res)
	if err != nil && res != "<red>[Invalid Number]</red>[rand]18446744073709551615[/rand]" {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected a number between 0 and 18446744073709551615")
	}
	msg = "[rand]170141183460469231731687303715884105727[/rand]" // Signed 128-bit MAX
	t.Log("Testing string '" + msg + "'")
	res = e.BbcodeFullParse(msg)
	_, err = strconv.Atoi(res)
	if err != nil && res != "<red>[Invalid Number]</red>[rand]170141183460469231731687303715884105727[/rand]" {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected a number between 0 and 170141183460469231731687303715884105727")
	}

	/*t.Log("Testing bbcode_regex_parse")
	for _, item := range l.Items {
		t.Log("Testing string '" + item.Msg + "'")
		res = bbcodeRegexParse(item.Msg)
		if res != item.Expects {
			t.Error("Bad output:", "'"+res+"'")
			t.Error("Expected:", item.Expects)
		}
	}*/
	
	l = &MEPairList{nil}
	l.Add("", "")
	l.Add("ddd", "ddd")
	l.Add("[b][/b]", "")
	l.Add("[b]ddd[/b]", "ddd")
	l.Add("ddd[b]ddd[/b]ddd", "ddddddddd")
	l.Add("ddd\n[b]ddd[/b]\nddd", "ddd\nddd\nddd")
	t.Log("Testing BbcodeStripTags")
	for _, item := range l.Items {
		res = e.BbcodeStripTags(item.Msg)
		if res != item.Expects {
			t.Error("Testing string '" + item.Msg + "'")
			t.Error("Bad output:", "'"+res+"'")
			t.Error("Expected:", "'"+item.Expects+"'")
		}
	}
}

func TestMarkdownRender(t *testing.T) {
	//t.Skip()
	if err := e.InitMarkdown(c.Plugins["markdown"]); err != nil {
		t.Fatal(err)
	}

	l := &MEPairList{nil}
	l2 := &MEPairList{nil}
	// TODO: Fix more of these odd cases
	l.Add("", "")
	l.Add(" ", " ")
	l.Add("  ", "  ")
	l.Add("   ", "   ")
	l.Add("\t", "\t")
	l.Add("\n", "\n")
	l.Add("*", "*")
	l.Add("`", "`")
	//l.Add("**", "<i></i>")
	l.Add("h", "h")
	l.Add("hi", "hi")
	l.Add("**h**", "<b>h</b>")
	l.Add("**hi**", "<b>hi</b>")
	l.Add("_h_", "<u>h</u>")
	l.Add("_hi_", "<u>hi</u>")
	l.Add(" _hi_", " <u>hi</u>")
	l.Add("h_hi_h", "h_hi_h")
	l.Add("h _hi_ h", "h <u>hi</u> h")
	l.Add("h _hi_h", "h <u>hi</u>h")
	l.Add("*h*", "<i>h</i>")
	l.Add("*hi*", "<i>hi</i>")
	l.Add("~h~", "<s>h</s>")
	l.Add("~hi~", "<s>hi</s>")
	l.Add("`hi`", "<blockquote>hi</blockquote>")
	// TODO: Hide the backslash after escaping the item
	// TODO: Doesn't behave consistently with d in-front of it
	l2.Add("\\`hi`", "\\`hi`")
	l2.Add("#", "#")
	l2.Add("#h", "<h2>h</h2>")
	l2.Add("#hi", "<h2>hi</h2>")
	l2.Add("# hi", "<h2>hi</h2>")
	l2.Add("#      hi", "<h2>hi</h2>")
	l.Add("\n#", "\n#")
	l.Add("\n#h", "\n<h2>h</h2>")
	l.Add("\n#hi", "\n<h2>hi</h2>")
	l.Add("\n#h\n", "\n<h2>h</h2>")
	l.Add("\n#hi\n", "\n<h2>hi</h2>")
	l.Add("*hi**", "<i>hi</i>*")
	l.Add("**hi***", "<b>hi</b>*")
	//l.Add("**hi*", "*<i>hi</i>")
	l.Add("***hi***", "<b><i>hi</i></b>")
	l.Add("***h***", "<b><i>h</i></b>")
	l.Add("\\***h**\\*", "*<b>h</b>*")
	l.Add("\\*\\**h*\\*\\*", "**<i>h</i>**")
	l.Add("\\*hi\\*", "*hi*")
	l.Add("d\\*hi\\*", "d*hi*")
	l.Add("\\*hi\\*d", "*hi*d")
	l.Add("d\\*hi\\*d", "d*hi*d")
	l.Add("\\", "\\")
	l.Add("\\\\", "\\\\")
	l.Add("\\d", "\\d")
	l.Add("\\\\d", "\\\\d")
	l.Add("\\\\\\d", "\\\\\\d")
	l.Add("d\\", "d\\")
	l.Add("\\d\\", "\\d\\")
	l.Add("*_hi_*", "<i><u>hi</u></i>")
	l.Add("*~hi~*", "<i><s>hi</s></i>")
	//l.Add("~*hi*~", "<s><i>hi</i></s>")
	//l.Add("~ *hi* ~", "<s> <i>hi</i> </s>")
	l.Add("_~hi~_", "<u><s>hi</s></u>")
	l.Add("***~hi~***", "<b><i><s>hi</s></i></b>")
	l.Add("**", "**")
	l.Add("***", "***")
	l.Add("****", "****")
	l.Add("*****", "*****")
	l.Add("******", "******")
	l.Add("*******", "*******")
	l.Add("~~", "~~")
	l.Add("~~~", "~~~")
	l.Add("~~~~", "~~~~")
	l.Add("~~~~~", "~~~~~")
	l.Add("|hi|", "<spoiler>hi</spoiler>")
	l.Add("__", "__")
	l.Add("___", "___")
	l.Add("_ _", "<u> </u>")
	l.Add("* *", "<i> </i>")
	l.Add("** **", "<b> </b>")
	l.Add("*** ***", "<b><i> </i></b>")
	l.Add("-你好-", "-你好-")
	l.Add("*-你好-*", "<i>-你好-</i>") // TODO: More of these Unicode tests? Emoji, Chinese, etc.?

	for _, item := range l.Items {
		if res := e.MarkdownParse(item.Msg); res != item.Expects {
			t.Error("Testing string '" + item.Msg + "'")
			t.Error("Bad output:", "'"+res+"'")
			//t.Error("Ouput in bytes:", []byte(res))
			t.Error("Expected:", "'"+item.Expects+"'")
		}
	}

	for _, item := range l2.Items {
		if res := e.MarkdownParse(item.Msg); res != item.Expects {
			t.Error("Testing string '" + item.Msg + "'")
			t.Error("Bad output:", "'"+res+"'")
			//t.Error("Ouput in bytes:", []byte(res))
			t.Error("Expected:", "'"+item.Expects+"'")
		}
	}

	/*for _, item := range l.Items {
		if res := e.MarkdownParse("d" + item.Msg); res != "d"+item.Expects {
			t.Error("Testing string 'd" + item.Msg + "'")
			t.Error("Bad output:", "'"+res+"'")
			//t.Error("Ouput in bytes:", []byte(res))
			t.Error("Expected:", "'d"+item.Expects+"'")
		}
	}*/

	// TODO: Write suffix tests and double string tests
	// TODO: Write similar prefix, suffix, and double string tests for plugin_bbcode. Ditto for the outer parser along with suitable tests for that like making sure the URL parser and media embedder works.
}
