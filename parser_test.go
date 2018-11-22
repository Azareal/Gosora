package main

import (
	"testing"

	"github.com/Azareal/Gosora/common"
)

func TestPreparser(t *testing.T) {
	var msgList = &METriList{nil}

	// Note: The open tag is evaluated without knowledge of the close tag for efficiency and simplicity, so the parser autofills the associated close tag when it finds an open tag without a partner
	msgList.Add("", "")
	msgList.Add(" ", "")
	msgList.Add(" hi", "hi")
	msgList.Add("hi ", "hi")
	msgList.Add("hi", "hi")
	msgList.Add(":grinning:", "😀")
	msgList.Add("😀", "😀")
	msgList.Add("&nbsp;", "")
	msgList.Add("<p>", "")
	msgList.Add("</p>", "")
	msgList.Add("<p></p>", "")

	msgList.Add("<", "&lt;")
	msgList.Add(">", "&gt;")
	msgList.Add("<meow>", "&lt;meow&gt;")
	msgList.Add("&lt;", "&amp;lt;")
	msgList.Add("&", "&amp;")

	// Note: strings.TrimSpace strips newlines, if there's nothing before or after them
	msgList.Add("<br>", "")
	msgList.Add("<br />", "")
	msgList.Add("\\n", "\n", "")
	msgList.Add("\\n\\n", "\n\n", "")
	msgList.Add("\\n\\n\\n", "\n\n\n", "")
	msgList.Add("\\r\\n", "\r\n", "") // Windows style line ending
	msgList.Add("\\n\\r", "\n\r", "")

	msgList.Add("ho<br>ho", "ho\n\nho")
	msgList.Add("ho<br />ho", "ho\n\nho")
	msgList.Add("ho\\nho", "ho\nho", "ho\nho")
	msgList.Add("ho\\n\\nho", "ho\n\nho", "ho\n\nho")
	//msgList.Add("ho\\n\\n\\n\\nho", "ho\n\n\n\nho", "ho\n\n\nho")
	msgList.Add("ho\\r\\nho", "ho\r\nho", "ho\nho") // Windows style line ending
	msgList.Add("ho\\n\\rho", "ho\n\rho", "ho\nho")

	msgList.Add("<b></b>", "<strong></strong>")
	msgList.Add("<b>hi</b>", "<strong>hi</strong>")
	msgList.Add("<s>hi</s>", "<del>hi</del>")
	msgList.Add("<del>hi</del>", "<del>hi</del>")
	msgList.Add("<u>hi</u>", "<u>hi</u>")
	msgList.Add("<em>hi</em>", "<em>hi</em>")
	msgList.Add("<i>hi</i>", "<em>hi</em>")
	msgList.Add("<strong>hi</strong>", "<strong>hi</strong>")
	msgList.Add("<b><i>hi</i></b>", "<strong><em>hi</em></strong>")
	msgList.Add("<strong><em>hi</em></strong>", "<strong><em>hi</em></strong>")
	msgList.Add("<b><i><b>hi</b></i></b>", "<strong><em><strong>hi</strong></em></strong>")
	msgList.Add("<strong><em><strong>hi</strong></em></strong>", "<strong><em><strong>hi</strong></em></strong>")
	msgList.Add("<div>hi</div>", "&lt;div&gt;hi&lt;/div&gt;")
	msgList.Add("<span>hi</span>", "hi") // This is stripped since the editor (Trumbowyg) likes blasting useless spans
	msgList.Add("<span   >hi</span>", "hi")
	msgList.Add("<span style='background-color: yellow;'>hi</span>", "hi")
	msgList.Add("<span style='background-color: yellow;'>>hi</span>", "&gt;hi")
	msgList.Add("<b>hi", "<strong>hi</strong>")
	msgList.Add("hi</b>", "hi&lt;/b&gt;")
	msgList.Add("</b>", "&lt;/b&gt;")
	msgList.Add("</del>", "&lt;/del&gt;")
	msgList.Add("</strong>", "&lt;/strong&gt;")
	msgList.Add("<b>", "<strong></strong>")
	msgList.Add("<span style='background-color: yellow;'>hi", "hi")
	msgList.Add("hi</span>", "hi")
	msgList.Add("</span>", "")
	msgList.Add("<span></span>", "")
	msgList.Add("<span   ></span>", "")
	msgList.Add("<></>", "&lt;&gt;&lt;/&gt;")
	msgList.Add("</><>", "&lt;/&gt;&lt;&gt;")
	msgList.Add("<>", "&lt;&gt;")
	msgList.Add("</>", "&lt;/&gt;")
	msgList.Add("@", "@")
	msgList.Add("@Admin", "@1")
	msgList.Add("@Bah", "@Bah")
	msgList.Add(" @Admin", "@1")
	msgList.Add("\n@Admin", "@1")
	msgList.Add("@Admin\n", "@1")
	msgList.Add("@Admin\ndd", "@1\ndd")
	msgList.Add("d@Admin", "d@Admin")
	//msgList.Add("byte 0", string([]byte{0}), "")
	msgList.Add("byte 'a'", string([]byte{'a'}), "a")
	//msgList.Add("byte 255", string([]byte{255}), "")
	//msgList.Add("rune 0", string([]rune{0}), "")
	// TODO: Do a test with invalid UTF-8 input

	for _, item := range msgList.Items {
		res := common.PreparseMessage(item.Msg)
		if res != item.Expects {
			if item.Name != "" {
				t.Error("Name: ", item.Name)
			}
			t.Error("Testing string '" + item.Msg + "'")
			t.Error("Bad output:", "'"+res+"'")
			//t.Error("Ouput in bytes:", []byte(res))
			t.Error("Expected:", "'"+item.Expects+"'")
		}
	}
}

func TestParser(t *testing.T) {
	var msgList = &METriList{nil}

	url := "github.com/Azareal/Gosora"
	msgList.Add("//"+url, "<a href='//"+url+"'>//"+url+"</a>")
	msgList.Add("https://"+url, "<a href='https://"+url+"'>https://"+url+"</a>")
	msgList.Add("http://"+url, "<a href='http://"+url+"'>http://"+url+"</a>")
	msgList.Add("//"+url+"\n", "<a href='//"+url+"'>//"+url+"</a><br>")
	msgList.Add("\n//"+url, "<br><a href='//"+url+"'>//"+url+"</a>")
	msgList.Add("\n//"+url+"\n", "<br><a href='//"+url+"'>//"+url+"</a><br>")
	msgList.Add("//"+url+"\n//"+url, "<a href='//"+url+"'>//"+url+"</a><br><a href='//"+url+"'>//"+url+"</a>")
	msgList.Add("//"+url+"\n\n//"+url, "<a href='//"+url+"'>//"+url+"</a><br><br><a href='//"+url+"'>//"+url+"</a>")
	msgList.Add("//"+common.Site.URL, "<a href='//"+common.Site.URL+"'>//"+common.Site.URL+"</a>")
	msgList.Add("//"+common.Site.URL+"\n", "<a href='//"+common.Site.URL+"'>//"+common.Site.URL+"</a><br>")
	msgList.Add("//"+common.Site.URL+"\n//"+common.Site.URL, "<a href='//"+common.Site.URL+"'>//"+common.Site.URL+"</a><br><a href='//"+common.Site.URL+"'>//"+common.Site.URL+"</a>")

	msgList.Add("#tid-1", "<a href='/topic/1'>#tid-1</a>")
	msgList.Add("https://"+url+"/#tid-1", "<a href='https://"+url+"/#tid-1'>https://"+url+"/#tid-1</a>")
	msgList.Add("#fid-1", "<a href='/forum/1'>#fid-1</a>")
	msgList.Add("@1", "<a href='/user/admin.1' class='mention'>@Admin</a>")
	msgList.Add("@0", "<span style='color: red;'>[Invalid Profile]</span>")
	msgList.Add("@-1", "<span style='color: red;'>[Invalid Profile]</span>1")

	for _, item := range msgList.Items {
		res := common.ParseMessage(item.Msg, 1, "forums")
		if res != item.Expects {
			if item.Name != "" {
				t.Error("Name: ", item.Name)
			}
			t.Error("Testing string '" + item.Msg + "'")
			t.Error("Bad output:", "'"+res+"'")
			t.Error("Expected:", "'"+item.Expects+"'")
		}
	}
}
