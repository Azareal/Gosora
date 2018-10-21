package main

import "strconv"
import "testing"

// go test -v

// TODO: Write a test for Hello World?

type MEPair struct {
	Msg     string
	Expects string
}

type MEPairList struct {
	Items []MEPair
}

func (tlist *MEPairList) Add(msg string, expects string) {
	tlist.Items = append(tlist.Items, MEPair{msg, expects})
}

func TestBBCodeRender(t *testing.T) {
	//t.Skip()
	err := initBbcode()
	if err != nil {
		t.Fatal(err)
	}

	var res string
	var msgList = &MEPairList{nil}
	msgList.Add("", "")
	msgList.Add(" ", " ")
	msgList.Add("  ", "  ")
	msgList.Add("   ", "   ")
	msgList.Add("[b]", "<b></b>")
	msgList.Add("[b][/b]", "<b></b>")
	msgList.Add("hi", "hi")
	msgList.Add("😀", "😀")
	msgList.Add("[b]😀[/b]", "<b>😀</b>")
	msgList.Add("[b]😀😀😀[/b]", "<b>😀😀😀</b>")
	msgList.Add("[b]hi[/b]", "<b>hi</b>")
	msgList.Add("[u]hi[/u]", "<u>hi</u>")
	msgList.Add("[i]hi[/i]", "<i>hi</i>")
	msgList.Add("[s]hi[/s]", "<s>hi</s>")
	msgList.Add("[c]hi[/c]", "[c]hi[/c]")
	if !testing.Short() {
		//msgList.Add("[b]hi[/i]", "[b]hi[/i]")
		//msgList.Add("[/b]hi[b]", "[/b]hi[b]")
		//msgList.Add("[/b]hi[/b]", "[/b]hi[/b]")
		//msgList.Add("[b][b]hi[/b]", "<b>hi</b>")
		//msgList.Add("[b][b]hi", "[b][b]hi")
		//msgList.Add("[b][b][b]hi", "[b][b][b]hi")
		//msgList.Add("[/b]hi", "[/b]hi")
	}
	msgList.Add("[code]hi[/code]", "<span class='codequotes'>hi</span>")
	msgList.Add("[code][b]hi[/b][/code]", "<span class='codequotes'>[b]hi[/b]</span>")
	msgList.Add("[code][b]hi[/code][/b]", "<span class='codequotes'>[b]hi</span>[/b]")
	msgList.Add("[quote]hi[/quote]", "<span class='postQuote'>hi</span>")
	msgList.Add("[quote][b]hi[/b][/quote]", "<span class='postQuote'><b>hi</b></span>")
	msgList.Add("[quote][b]h[/b][/quote]", "<span class='postQuote'><b>h</b></span>")
	msgList.Add("[quote][b][/b][/quote]", "<span class='postQuote'><b></b></span>")
	msgList.Add("[url][/url]", "<a href=''></a>")
	msgList.Add("[url]https://github.com/Azareal/Gosora[/url]", "<a href='https://github.com/Azareal/Gosora'>https://github.com/Azareal/Gosora</a>")
	msgList.Add("[url]http://github.com/Azareal/Gosora[/url]", "<a href='http://github.com/Azareal/Gosora'>http://github.com/Azareal/Gosora</a>")
	msgList.Add("[url]//github.com/Azareal/Gosora[/url]", "<a href='//github.com/Azareal/Gosora'>//github.com/Azareal/Gosora</a>")
	msgList.Add("-你好-", "-你好-")
	msgList.Add("[i]-你好-[/i]", "<i>-你好-</i>") // TODO: More of these Unicode tests? Emoji, Chinese, etc.?

	t.Log("Testing bbcodeFullParse")
	for _, item := range msgList.Items {
		res = bbcodeFullParse(item.Msg)
		if res != item.Expects {
			t.Error("Testing string '" + item.Msg + "'")
			t.Error("Bad output:", "'"+res+"'")
			t.Error("Expected:", "'"+item.Expects+"'")
		}
	}

	var msg string
	var expects string

	msg = "[rand][/rand]"
	expects = "<span style='color: red;'>[Invalid Number]</span>[rand][/rand]"
	t.Log("Testing string '" + msg + "'")
	res = bbcodeFullParse(msg)
	if res != expects {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected:", "'"+expects+"'")
	}

	msg = "[rand]-1[/rand]"
	expects = "<span style='color: red;'>[No Negative Numbers]</span>[rand]-1[/rand]"
	t.Log("Testing string '" + msg + "'")
	res = bbcodeFullParse(msg)
	if res != expects {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected:", "'"+expects+"'")
	}

	msg = "[rand]-01[/rand]"
	expects = "<span style='color: red;'>[No Negative Numbers]</span>[rand]-01[/rand]"
	t.Log("Testing string '" + msg + "'")
	res = bbcodeFullParse(msg)
	if res != expects {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected:", "'"+expects+"'")
	}

	msg = "[rand]NaN[/rand]"
	expects = "<span style='color: red;'>[Invalid Number]</span>[rand]NaN[/rand]"
	t.Log("Testing string '" + msg + "'")
	res = bbcodeFullParse(msg)
	if res != expects {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected:", "'"+expects+"'")
	}

	msg = "[rand]Inf[/rand]"
	expects = "<span style='color: red;'>[Invalid Number]</span>[rand]Inf[/rand]"
	t.Log("Testing string '" + msg + "'")
	res = bbcodeFullParse(msg)
	if res != expects {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected:", "'"+expects+"'")
	}

	msg = "[rand]+[/rand]"
	expects = "<span style='color: red;'>[Invalid Number]</span>[rand]+[/rand]"
	t.Log("Testing string '" + msg + "'")
	res = bbcodeFullParse(msg)
	if res != expects {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected:", "'"+expects+"'")
	}

	msg = "[rand]1+1[/rand]"
	expects = "<span style='color: red;'>[Invalid Number]</span>[rand]1+1[/rand]"
	t.Log("Testing string '" + msg + "'")
	res = bbcodeFullParse(msg)
	if res != expects {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected:", "'"+expects+"'")
	}

	var conv int
	msg = "[rand]1[/rand]"
	t.Log("Testing string '" + msg + "'")
	res = bbcodeFullParse(msg)
	conv, err = strconv.Atoi(res)
	if err != nil || (conv > 1 || conv < 0) {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected a number in the range 0-1")
	}

	msg = "[rand]0[/rand]"
	t.Log("Testing string '" + msg + "'")
	res = bbcodeFullParse(msg)
	conv, err = strconv.Atoi(res)
	if err != nil || conv != 0 {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected the number 0")
	}

	msg = "[rand]2147483647[/rand]" // Signed 32-bit MAX
	t.Log("Testing string '" + msg + "'")
	res = bbcodeFullParse(msg)
	conv, err = strconv.Atoi(res)
	if err != nil || (conv > 2147483647 || conv < 0) {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected a number between 0 and 2147483647")
	}

	msg = "[rand]9223372036854775807[/rand]" // Signed 64-bit MAX
	t.Log("Testing string '" + msg + "'")
	res = bbcodeFullParse(msg)
	conv, err = strconv.Atoi(res)
	if err != nil || (conv > 9223372036854775807 || conv < 0) {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected a number between 0 and 9223372036854775807")
	}

	// Note: conv is commented out in these two, as these numbers overflow int
	msg = "[rand]18446744073709551615[/rand]" // Unsigned 64-bit MAX
	t.Log("Testing string '" + msg + "'")
	res = bbcodeFullParse(msg)
	_, err = strconv.Atoi(res)
	if err != nil && res != "<span style='color: red;'>[Invalid Number]</span>[rand]18446744073709551615[/rand]" {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected a number between 0 and 18446744073709551615")
	}
	msg = "[rand]170141183460469231731687303715884105727[/rand]" // Signed 128-bit MAX
	t.Log("Testing string '" + msg + "'")
	res = bbcodeFullParse(msg)
	_, err = strconv.Atoi(res)
	if err != nil && res != "<span style='color: red;'>[Invalid Number]</span>[rand]170141183460469231731687303715884105727[/rand]" {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected a number between 0 and 170141183460469231731687303715884105727")
	}

	/*t.Log("Testing bbcode_regex_parse")
	for _, item := range msgList {
		t.Log("Testing string '" + item.Msg + "'")
		res = bbcodeRegexParse(item.Msg)
		if res != item.Expects {
			t.Error("Bad output:", "'"+res+"'")
			t.Error("Expected:", item.Expects)
		}
	}*/
}

func TestMarkdownRender(t *testing.T) {
	//t.Skip()
	err := initMarkdown()
	if err != nil {
		t.Fatal(err)
	}

	var res string
	var msgList = &MEPairList{nil}
	// TODO: Fix more of these odd cases
	msgList.Add("", "")
	msgList.Add(" ", " ")
	msgList.Add("  ", "  ")
	msgList.Add("   ", "   ")
	msgList.Add("\t", "\t")
	msgList.Add("\n", "\n")
	msgList.Add("*", "*")
	//msgList.Add("**", "<i></i>")
	msgList.Add("h", "h")
	msgList.Add("hi", "hi")
	msgList.Add("**h**", "<b>h</b>")
	msgList.Add("**hi**", "<b>hi</b>")
	msgList.Add("_h_", "<u>h</u>")
	msgList.Add("_hi_", "<u>hi</u>")
	msgList.Add("*h*", "<i>h</i>")
	msgList.Add("*hi*", "<i>hi</i>")
	msgList.Add("~h~", "<s>h</s>")
	msgList.Add("~hi~", "<s>hi</s>")
	msgList.Add("*hi**", "<i>hi</i>*")
	msgList.Add("**hi***", "<b>hi</b>*")
	//msgList.Add("**hi*", "*<i>hi</i>")
	msgList.Add("***hi***", "<b><i>hi</i></b>")
	msgList.Add("***h***", "<b><i>h</i></b>")
	msgList.Add("\\***h**\\*", "*<b>h</b>*")
	msgList.Add("\\*\\**h*\\*\\*", "**<i>h</i>**")
	msgList.Add("\\*hi\\*", "*hi*")
	msgList.Add("d\\*hi\\*", "d*hi*")
	msgList.Add("\\*hi\\*d", "*hi*d")
	msgList.Add("d\\*hi\\*d", "d*hi*d")
	msgList.Add("\\", "\\")
	msgList.Add("\\\\", "\\\\")
	msgList.Add("\\d", "\\d")
	msgList.Add("\\\\d", "\\\\d")
	msgList.Add("\\\\\\d", "\\\\\\d")
	msgList.Add("d\\", "d\\")
	msgList.Add("\\d\\", "\\d\\")
	msgList.Add("*_hi_*", "<i><u>hi</u></i>")
	msgList.Add("*~hi~*", "<i><s>hi</s></i>")
	//msgList.Add("~*hi*~", "<s><i>hi</i></s>")
	//msgList.Add("~ *hi* ~", "<s> <i>hi</i> </s>")
	msgList.Add("_~hi~_", "<u><s>hi</s></u>")
	msgList.Add("***~hi~***", "<b><i><s>hi</s></i></b>")
	msgList.Add("**", "**")
	msgList.Add("***", "***")
	msgList.Add("****", "****")
	msgList.Add("*****", "*****")
	msgList.Add("******", "******")
	msgList.Add("*******", "*******")
	msgList.Add("~~", "~~")
	msgList.Add("~~~", "~~~")
	msgList.Add("~~~~", "~~~~")
	msgList.Add("~~~~~", "~~~~~")
	msgList.Add("__", "__")
	msgList.Add("___", "___")
	msgList.Add("_ _", "<u> </u>")
	msgList.Add("* *", "<i> </i>")
	msgList.Add("** **", "<b> </b>")
	msgList.Add("*** ***", "<b><i> </i></b>")
	msgList.Add("-你好-", "-你好-")
	msgList.Add("*-你好-*", "<i>-你好-</i>") // TODO: More of these Unicode tests? Emoji, Chinese, etc.?

	for _, item := range msgList.Items {
		res = markdownParse(item.Msg)
		if res != item.Expects {
			t.Error("Testing string '" + item.Msg + "'")
			t.Error("Bad output:", "'"+res+"'")
			//t.Error("Ouput in bytes:", []byte(res))
			t.Error("Expected:", "'"+item.Expects+"'")
		}
	}

	/*for _, item := range msgList.Items {
		res = markdownParse("\n" + item.Msg)
		if res != "\n"+item.Expects {
			t.Error("Testing string '\n" + item.Msg + "'")
			t.Error("Bad output:", "'"+res+"'")
			//t.Error("Ouput in bytes:", []byte(res))
			t.Error("Expected:", "'\n"+item.Expects+"'")
		}
	}

	for _, item := range msgList.Items {
		res = markdownParse("\t" + item.Msg)
		if res != "\t"+item.Expects {
			t.Error("Testing string '\t" + item.Msg + "'")
			t.Error("Bad output:", "'"+res+"'")
			//t.Error("Ouput in bytes:", []byte(res))
			t.Error("Expected:", "'\t"+item.Expects+"'")
		}
	}*/

	for _, item := range msgList.Items {
		res = markdownParse("d" + item.Msg)
		if res != "d"+item.Expects {
			t.Error("Testing string 'd" + item.Msg + "'")
			t.Error("Bad output:", "'"+res+"'")
			//t.Error("Ouput in bytes:", []byte(res))
			t.Error("Expected:", "'d"+item.Expects+"'")
		}
	}

	// TODO: Write suffix tests and double string tests
	// TODO: Write similar prefix, suffix, and double string tests for plugin_bbcode. Ditto for the outer parser along with suitable tests for that like making sure the URL parser and media embedder works.
}
