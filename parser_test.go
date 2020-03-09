package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	c "github.com/Azareal/Gosora/common"
)

func TestPreparser(t *testing.T) {
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}
	l := &METriList{nil}

	// Note: The open tag is evaluated without knowledge of the close tag for efficiency and simplicity, so the parser autofills the associated close tag when it finds an open tag without a partner
	l.Add("", "")
	l.Add(" ", "")
	l.Add(" hi", "hi")
	l.Add("hi ", "hi")
	l.Add("hi", "hi")
	l.Add(":grinning:", "😀")
	l.Add("😀", "😀")
	l.Add("&nbsp;", "")
	l.Add("<p>", "")
	l.Add("</p>", "")
	l.Add("<p></p>", "")

	l.Add("<", "&lt;")
	l.Add(">", "&gt;")
	l.Add("<meow>", "&lt;meow&gt;")
	l.Add("&lt;", "&amp;lt;")
	l.Add("&", "&amp;")

	// Note: strings.TrimSpace strips newlines, if there's nothing before or after them
	l.Add("<br>", "")
	l.Add("<br />", "")
	l.Add("\\n", "\n", "")
	l.Add("\\n\\n", "\n\n", "")
	l.Add("\\n\\n\\n", "\n\n\n", "")
	l.Add("\\r\\n", "\r\n", "") // Windows style line ending
	l.Add("\\n\\r", "\n\r", "")

	l.Add("ho<br>ho", "ho\n\nho")
	l.Add("ho<br />ho", "ho\n\nho")
	l.Add("ho\\nho", "ho\nho", "ho\nho")
	l.Add("ho\\n\\nho", "ho\n\nho", "ho\n\nho")
	//l.Add("ho\\n\\n\\n\\nho", "ho\n\n\n\nho", "ho\n\n\nho")
	l.Add("ho\\r\\nho", "ho\r\nho", "ho\nho") // Windows style line ending
	l.Add("ho\\n\\rho", "ho\n\rho", "ho\nho")

	l.Add("<b></b>", "<strong></strong>")
	l.Add("<b>hi</b>", "<strong>hi</strong>")
	l.Add("<b>h</b>", "<strong>h</strong>")
	l.Add("<s>hi</s>", "<del>hi</del>")
	l.Add("<del>hi</del>", "<del>hi</del>")
	l.Add("<u>hi</u>", "<u>hi</u>")
	l.Add("<em>hi</em>", "<em>hi</em>")
	l.Add("<i>hi</i>", "<em>hi</em>")
	l.Add("<strong>hi</strong>", "<strong>hi</strong>")
	l.Add("<spoiler>hi</spoiler>", "<spoiler>hi</spoiler>")
	l.Add("<g>hi</g>", "hi") // Grammarly fix
	l.Add("<b><i>hi</i></b>", "<strong><em>hi</em></strong>")
	l.Add("<strong><em>hi</em></strong>", "<strong><em>hi</em></strong>")
	l.Add("<b><i><b>hi</b></i></b>", "<strong><em><strong>hi</strong></em></strong>")
	l.Add("<strong><em><strong>hi</strong></em></strong>", "<strong><em><strong>hi</strong></em></strong>")
	l.Add("<div>hi</div>", "&lt;div&gt;hi&lt;/div&gt;")
	l.Add("<span>hi</span>", "hi") // This is stripped since the editor (Trumbowyg) likes blasting useless spans
	l.Add("<span   >hi</span>", "hi")
	l.Add("<span style='background-color: yellow;'>hi</span>", "hi")
	l.Add("<span style='background-color: yellow;'>>hi</span>", "&gt;hi")
	l.Add("<b>hi", "<strong>hi</strong>")
	l.Add("hi</b>", "hi&lt;/b&gt;")
	l.Add("</b>", "&lt;/b&gt;")
	l.Add("</del>", "&lt;/del&gt;")
	l.Add("</strong>", "&lt;/strong&gt;")
	l.Add("<b>", "<strong></strong>")
	l.Add("<span style='background-color: yellow;'>hi", "hi")
	l.Add("<span style='background-color:yellow;'>hi", "hi")
	l.Add("hi</span>", "hi")
	l.Add("</span>", "")
	l.Add("<span></span>", "")
	l.Add("<span   ></span>", "")
	l.Add("<span><span></span></span>", "")
	l.Add("<span><b></b></span>", "<strong></strong>")
	l.Add("<h1>t</h1>", "<h2>t</h2>")
	l.Add("<h2>t</h2>", "<h3>t</h3>")
	l.Add("<h3>t</h3>", "<h4>t</h4>")
	l.Add("<></>", "&lt;&gt;&lt;/&gt;")
	l.Add("</><>", "&lt;/&gt;&lt;&gt;")
	l.Add("<>", "&lt;&gt;")
	l.Add("</>", "&lt;/&gt;")
	l.Add("<p>hi</p>", "hi")
	l.Add("<p></p>", "")
	l.Add("<blockquote>hi</blockquote>", "<blockquote>hi</blockquote>")
	l.Add("<blockquote><b>hi</b></blockquote>", "<blockquote><strong>hi</strong></blockquote>")
	l.Add("<blockquote><meow>hi</meow></blockquote>", "<blockquote>&lt;meow&gt;hi&lt;/meow&gt;</blockquote>")
	l.Add("\\<blockquote>hi</blockquote>", "&lt;blockquote&gt;hi&lt;/blockquote&gt;")
	//l.Add("\\\\<blockquote><meow>hi</meow></blockquote>", "\\<blockquote>&lt;meow&gt;hi&lt;/meow&gt;</blockquote>") // TODO: Double escapes should print a literal backslash
	//l.Add("&lt;blockquote&gt;hi&lt;/blockquote&gt;", "&lt;blockquote&gt;hi&lt;/blockquote&gt;") // TODO: Stop double-entitising this
	l.Add("\\<blockquote>hi</blockquote>\\<blockquote>hi</blockquote>", "&lt;blockquote&gt;hi&lt;/blockquote&gt;&lt;blockquote&gt;hi&lt;/blockquote&gt;")
	l.Add("\\<a itemprop=\"author\">Admin</a>", "&lt;a itemprop=&#34;author&#34;&gt;Admin&lt;/a&gt;")
	l.Add("<blockquote>\\<a itemprop=\"author\">Admin</a></blockquote>", "<blockquote>&lt;a itemprop=&#34;author&#34;&gt;Admin&lt;/a&gt;</blockquote>")
	l.Add("\n<blockquote>\\<a itemprop=\"author\">Admin</a></blockquote>\n", "<blockquote>&lt;a itemprop=&#34;author&#34;&gt;Admin&lt;/a&gt;</blockquote>")
	l.Add("tt\n<blockquote>\\<a itemprop=\"author\">Admin</a></blockquote>\ntt", "tt\n<blockquote>&lt;a itemprop=&#34;author&#34;&gt;Admin&lt;/a&gt;</blockquote>\ntt")
	l.Add("@", "@")
	l.Add("@Admin", "@1")
	l.Add("@Bah", "@Bah")
	l.Add(" @Admin", "@1")
	l.Add("\n@Admin", "@1")
	l.Add("@Admin\n", "@1")
	l.Add("@Admin\ndd", "@1\ndd")
	l.Add("d@Admin", "d@Admin")
	l.Add("\\@Admin", "@Admin")
	l.Add("@元気", "@元気")
	// TODO: More tests for unicode names?
	//l.Add("\\\\@Admin", "@1")
	//l.Add("byte 0", string([]byte{0}), "")
	l.Add("byte 'a'", string([]byte{'a'}), "a")
	//l.Add("byte 255", string([]byte{255}), "")
	//l.Add("rune 0", string([]rune{0}), "")
	// TODO: Do a test with invalid UTF-8 input

	for _, item := range l.Items {
		if res := c.PreparseMessage(item.Msg); res != item.Expects {
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
	miscinit(t)
	if !c.PluginsInited {
		c.InitPlugins()
	}
	l := &METriList{nil}

	url := "github.com/Azareal/Gosora"
	eurl := "<a rel='ugc' href='//" + url + "'>" + url + "</a>"
	l.Add("", "")
	l.Add("haha", "haha")
	l.Add("<b>t</b>", "<b>t</b>")
	l.Add("//", "//")
	l.Add("http://", "<red>[Invalid URL]</red>")
	l.Add("https://", "<red>[Invalid URL]</red>")
	l.Add("ftp://", "<red>[Invalid URL]</red>")
	l.Add("git://", "<red>[Invalid URL]</red>")
	l.Add("ssh://", "ssh://")

	l.Add("// ", "// ")
	l.Add("// //", "// //")
	l.Add("// // //", "// // //")
	l.Add("http:// ", "<red>[Invalid URL]</red> ")
	l.Add("https:// ", "<red>[Invalid URL]</red> ")
	l.Add("ftp:// ", "<red>[Invalid URL]</red> ")
	l.Add("git:// ", "<red>[Invalid URL]</red> ")
	l.Add("ssh:// ", "ssh:// ")

	l.Add("// t", "// t")
	l.Add("http:// t", "<red>[Invalid URL]</red> t")

	l.Add("h", "h")
	l.Add("ht", "ht")
	l.Add("htt", "htt")
	l.Add("http", "http")
	l.Add("http:", "http:")
	//t l.Add("http:/", "http:/")
	//t l.Add("http:/d", "http:/d")
	l.Add("http:d", "http:d")
	l.Add("https:", "https:")
	l.Add("ftp:", "ftp:")
	l.Add("git:", "git:")
	l.Add("ssh:", "ssh:")

	l.Add("http", "http")
	l.Add("https", "https")
	l.Add("ftp", "ftp")
	l.Add("git", "git")
	l.Add("ssh", "ssh")

	l.Add("ht", "ht")
	l.Add("htt", "htt")
	l.Add("ft", "ft")
	l.Add("gi", "gi")
	l.Add("ss", "ss")
	l.Add("haha\nhaha\nhaha", "haha<br>haha<br>haha")
	l.Add("//"+url, eurl)
	l.Add("//a", "<a rel='ugc' href='//a'>a</a>")
	l.Add(" //a", " <a rel='ugc' href='//a'>a</a>")
	l.Add("//a ", "<a rel='ugc' href='//a'>a</a> ")
	l.Add(" //a ", " <a rel='ugc' href='//a'>a</a> ")
	l.Add("d //a ", "d <a rel='ugc' href='//a'>a</a> ")
	l.Add("ddd ddd //a ", "ddd ddd <a rel='ugc' href='//a'>a</a> ")
	l.Add("https://"+url, "<a rel='ugc' href='https://"+url+"'>"+url+"</a>")
	l.Add("https://t", "<a rel='ugc' href='https://t'>t</a>")
	l.Add("http://"+url, "<a rel='ugc' href='http://"+url+"'>"+url+"</a>")
	l.Add("#http://"+url, "#http://"+url)
	l.Add("@http://"+url, "<red>[Invalid Profile]</red>ttp://"+url)
	l.Add("//"+url+"\n", "<a rel='ugc' href='//"+url+"'>"+url+"</a><br>")
	l.Add("\n//"+url, "<br>"+eurl)
	l.Add("\n//"+url+"\n", "<br>"+eurl+"<br>")
	l.Add("\n//"+url+"\n\n", "<br>"+eurl+"<br><br>")
	l.Add("//"+url+"\n//"+url, eurl+"<br>"+eurl)
	l.Add("//"+url+"\n\n//"+url, eurl+"<br><br>"+eurl)

	pre2 := c.Config.SslSchema
	c.Config.SslSchema = true
	local := func(u string) {
		s := "//" + c.Site.URL
		fs := "http://" + c.Site.URL
		if c.Config.SslSchema {
			s = "https:" + s
			fs = "https://" + c.Site.URL
		}
		l.Add("//"+u, "<a href='"+fs+"'>"+c.Site.URL+"</a>")
		l.Add("//"+u+"\n", "<a href='"+fs+"'>"+c.Site.URL+"</a><br>")
		l.Add("//"+u+"\n//"+u, "<a href='"+fs+"'>"+c.Site.URL+"</a><br><a href='"+fs+"'>"+c.Site.URL+"</a>")
		l.Add("http://"+u, "<a href='"+fs+"'>"+c.Site.URL+"</a>")
		l.Add("https://"+u, "<a href='"+fs+"'>"+c.Site.URL+"</a>")
	}
	local("localhost")
	local("127.0.0.1")
	local("[::1]")

	l.Add("https://www.youtube.com/watch?v=lalalalala", "<iframe class='postIframe' src='https://www.youtube-nocookie.com/embed/lalalalala' frameborder=0 allowfullscreen></iframe>")
	//l.Add("https://www.youtube.com/watch?v=;","<iframe class='postIframe' src='https://www.youtube-nocookie.com/embed/;' frameborder=0 allowfullscreen></iframe>")
	l.Add("https://www.youtube.com/watch?v=d;", "<iframe class='postIframe' src='https://www.youtube-nocookie.com/embed/d' frameborder=0 allowfullscreen></iframe>")
	l.Add("https://www.youtube.com/watch?v=d;d", "<iframe class='postIframe' src='https://www.youtube-nocookie.com/embed/d' frameborder=0 allowfullscreen></iframe>")
	l.Add("https://www.youtube.com/watch?v=alert()", "<red>[Invalid URL]</red>()")
	l.Add("https://www.youtube.com/watch?v=alert()()", "<red>[Invalid URL]</red>()()")
	l.Add("https://www.youtube.com/watch?v=js:alert()", "<red>[Invalid URL]</red>()")
	l.Add("https://www.youtube.com/watch?v='+><script>alert(\"\")</script><+'", "<red>[Invalid URL]</red>'+><script>alert(\"\")</script><+'")
	l.Add("https://www.youtube.com/watch?v='+onready='alert(\"\")'+'", "<red>[Invalid URL]</red>'+onready='alert(\"\")'+'")
	l.Add(" https://www.youtube.com/watch?v=lalalalala", " <iframe class='postIframe' src='https://www.youtube-nocookie.com/embed/lalalalala' frameborder=0 allowfullscreen></iframe>")
	l.Add("https://www.youtube.com/watch?v=lalalalala tt", "<iframe class='postIframe' src='https://www.youtube-nocookie.com/embed/lalalalala' frameborder=0 allowfullscreen></iframe> tt")
	l.Add("https://www.youtube.com/watch?v=lalalalala&d=haha", "<iframe class='postIframe' src='https://www.youtube-nocookie.com/embed/lalalalala' frameborder=0 allowfullscreen></iframe>")
	l.Add("https://gaming.youtube.com/watch?v=lalalalala", "<iframe class='postIframe' src='https://www.youtube-nocookie.com/embed/lalalalala' frameborder=0 allowfullscreen></iframe>")
	l.Add("https://gaming.youtube.com/watch?v=lalalalala&d=haha", "<iframe class='postIframe' src='https://www.youtube-nocookie.com/embed/lalalalala' frameborder=0 allowfullscreen></iframe>")
	l.Add("https://m.youtube.com/watch?v=lalalalala", "<iframe class='postIframe' src='https://www.youtube-nocookie.com/embed/lalalalala' frameborder=0 allowfullscreen></iframe>")
	l.Add("https://m.youtube.com/watch?v=lalalalala&d=haha", "<iframe class='postIframe' src='https://www.youtube-nocookie.com/embed/lalalalala' frameborder=0 allowfullscreen></iframe>")
	l.Add("http://www.youtube.com/watch?v=lalalalala", "<iframe class='postIframe' src='https://www.youtube-nocookie.com/embed/lalalalala' frameborder=0 allowfullscreen></iframe>")
	l.Add("//www.youtube.com/watch?v=lalalalala", "<iframe class='postIframe' src='https://www.youtube-nocookie.com/embed/lalalalala' frameborder=0 allowfullscreen></iframe>")
	//l.Add("www.youtube.com/watch?v=lalalalala","<iframe class='postIframe' src='https://www.youtube-nocookie.com/embed/lalalalala' frameborder=0 allowfullscreen></iframe>")

	l.Add("#tid-1", "<a href='/topic/1'>#tid-1</a>")
	l.Add("##tid-1", "##tid-1")
	l.Add("# #tid-1", "# #tid-1")
	l.Add("@ #tid-1", "<red>[Invalid Profile]</red>#tid-1")
	l.Add("@#tid-1", "<red>[Invalid Profile]</red>tid-1")
	l.Add("@ #tid-@", "<red>[Invalid Profile]</red>#tid-@")
	l.Add("#tid-1 #tid-1", "<a href='/topic/1'>#tid-1</a> <a href='/topic/1'>#tid-1</a>")
	l.Add("#tid-0", "<red>[Invalid Topic]</red>")
	l.Add("https://"+url+"/#tid-1", "<a rel='ugc' href='https://"+url+"/#tid-1'>"+url+"/#tid-1</a>")
	l.Add("https://"+url+"/?hi=2", "<a rel='ugc' href='https://"+url+"/?hi=2'>"+url+"/?hi=2</a>")
	l.Add("https://"+url+"/?hi=2#t=1", "<a rel='ugc' href='https://"+url+"/?hi=2#t=1'>"+url+"/?hi=2#t=1</a>")
	l.Add("#fid-1", "<a href='/forum/1'>#fid-1</a>")
	l.Add(" #fid-1", " <a href='/forum/1'>#fid-1</a>")
	l.Add("#fid-0", "<red>[Invalid Forum]</red>")
	l.Add(" #fid-0", " <red>[Invalid Forum]</red>")
	l.Add("#", "#")
	l.Add("# ", "# ")
	l.Add(" @", " @")
	l.Add(" #", " #")
	l.Add("#@", "#@")
	l.Add("#@ ", "#@ ")
	l.Add("#@1", "#@1")
	l.Add("#f", "#f")
	l.Add("f#f", "f#f")
	l.Add("f#", "f#")
	l.Add("#ff", "#ff")
	l.Add("#ffffid-0", "#ffffid-0")
	//l.Add("#ffffid-0", "#ffffid-0")
	l.Add("#nid-0", "#nid-0")
	l.Add("#nnid-0", "#nnid-0")
	l.Add("@@", "<red>[Invalid Profile]</red>")
	l.Add("@@ @@", "<red>[Invalid Profile]</red> <red>[Invalid Profile]</red>")
	l.Add("@@1", "<red>[Invalid Profile]</red>1")
	l.Add("@#1", "<red>[Invalid Profile]</red>1")
	l.Add("@##1", "<red>[Invalid Profile]</red>#1")
	l.Add("@2", "<red>[Invalid Profile]</red>")
	l.Add("@2t", "<red>[Invalid Profile]</red>t")
	l.Add("@2 t", "<red>[Invalid Profile]</red> t")
	l.Add("@2 ", "<red>[Invalid Profile]</red> ")
	l.Add("@2 @2", "<red>[Invalid Profile]</red> <red>[Invalid Profile]</red>")
	l.Add("@1", "<a href='/user/admin.1' class='mention'>@Admin</a>")
	l.Add(" @1", " <a href='/user/admin.1' class='mention'>@Admin</a>")
	l.Add("@1t", "<a href='/user/admin.1' class='mention'>@Admin</a>t")
	l.Add("@1 ", "<a href='/user/admin.1' class='mention'>@Admin</a> ")
	l.Add("@1 @1", "<a href='/user/admin.1' class='mention'>@Admin</a> <a href='/user/admin.1' class='mention'>@Admin</a>")
	l.Add("@0", "<red>[Invalid Profile]</red>")
	l.Add("@-1", "<red>[Invalid Profile]</red>1")

	// TODO: Fix this hack and make the results a bit more reproducible, push the tests further in the process.
	for _, item := range l.Items {
		if res := c.ParseMessage(item.Msg, 1, "forums", nil, nil); res != item.Expects {
			if item.Name != "" {
				t.Error("Name: ", item.Name)
			}
			t.Error("Testing string '" + item.Msg + "'")
			t.Error("Bad output:", "'"+res+"'")
			t.Error("Expected:", "'"+item.Expects+"'")
			break
		}
	}
	c.Config.SslSchema = pre2

	l = &METriList{nil}
	pre := c.Site.URL // Just in case this is localhost...
	pre2 = c.Config.SslSchema
	c.Site.URL = "example.com"
	c.Config.SslSchema = true
	l.Add("//"+c.Site.URL, "<a href='https://"+c.Site.URL+"'>"+c.Site.URL+"</a>")
	l.Add("//"+c.Site.URL+"\n", "<a href='https://"+c.Site.URL+"'>"+c.Site.URL+"</a><br>")
	l.Add("//"+c.Site.URL+"\n//"+c.Site.URL, "<a href='https://"+c.Site.URL+"'>"+c.Site.URL+"</a><br><a href='https://"+c.Site.URL+"'>"+c.Site.URL+"</a>")
	for _, item := range l.Items {
		if res := c.ParseMessage(item.Msg, 1, "forums", nil, nil); res != item.Expects {
			if item.Name != "" {
				t.Error("Name: ", item.Name)
			}
			t.Error("Testing string '" + item.Msg + "'")
			t.Error("Bad output:", "'"+res+"'")
			t.Error("Expected:", "'"+item.Expects+"'")
			break
		}
	}
	c.Site.URL = pre
	c.Config.SslSchema = pre2

	c.AddHashLinkType("nnid-", func(sb *strings.Builder, msg string, i *int) {
		tid, intLen := c.CoerceIntString(msg[*i:])
		*i += intLen

		topic, err := c.Topics.Get(tid)
		if err != nil || !c.Forums.Exists(topic.ParentID) {
			sb.Write(c.InvalidTopic)
			return
		}
		c.WriteURL(sb, c.BuildTopicURL("", tid), "#nnid-"+strconv.Itoa(tid))
	})
	res := c.ParseMessage("#nnid-1", 1, "forums", nil, nil)
	expect := "<a href='/topic/1'>#nnid-1</a>"
	if res != expect {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected:", "'"+expect+"'")
	}

	c.AddHashLinkType("longidnameneedtooverflowhack-", func(sb *strings.Builder, msg string, i *int) {
		tid, intLen := c.CoerceIntString(msg[*i:])
		*i += intLen

		topic, err := c.Topics.Get(tid)
		if err != nil || !c.Forums.Exists(topic.ParentID) {
			sb.Write(c.InvalidTopic)
			return
		}
		c.WriteURL(sb, c.BuildTopicURL("", tid), "#longidnameneedtooverflowhack-"+strconv.Itoa(tid))
	})
	res = c.ParseMessage("#longidnameneedtooverflowhack-1", 1, "forums", nil, nil)
	expect = "<a href='/topic/1'>#longidnameneedtooverflowhack-1</a>"
	if res != expect {
		t.Error("Bad output:", "'"+res+"'")
		t.Error("Expected:", "'"+expect+"'")
	}
}

func TestPaginate(t *testing.T) {
	var plist []int
	f := func(i, want int) {
		expect(t, plist[i] == want, fmt.Sprintf("plist[%d] should be %d not %d", i, want, plist[i]))
	}

	plist = c.Paginate(1, 1, 5)
	expect(t, len(plist) == 1, fmt.Sprintf("len of plist should be 1 not %d", len(plist)))
	f(0, 1)

	plist = c.Paginate(1, 5, 5)
	expect(t, len(plist) == 5, fmt.Sprintf("len of plist should be 5 not %d", len(plist)))
	f(0, 1)
	f(1, 2)
	f(2, 3)
	f(3, 4)
	f(4, 5)

	// TODO: More Paginate() tests
	// TODO: Tests for other paginator functions
}
