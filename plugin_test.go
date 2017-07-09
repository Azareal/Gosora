package main

import "strconv"
import "testing"

// go test -v

type ME_Pair struct
{
  Msg string
  Expects string
}

func addMEPair(msgList []ME_Pair, msg string, expects string) []ME_Pair {
  return append(msgList,ME_Pair{msg,expects})
}

func TestBBCodeRender(t *testing.T) {
  var res string
  var msgList []ME_Pair
  msgList = addMEPair(msgList,"hi","hi")
  msgList = addMEPair(msgList,"😀","😀")
  msgList = addMEPair(msgList,"[b]😀[/b]","<b>😀</b>")
  msgList = addMEPair(msgList,"[b]😀😀😀[/b]","<b>😀😀😀</b>")
  msgList = addMEPair(msgList,"[b]hi[/b]","<b>hi</b>")
  msgList = addMEPair(msgList,"[u]hi[/u]","<u>hi</u>")
  msgList = addMEPair(msgList,"[i]hi[/i]","<i>hi</i>")
  msgList = addMEPair(msgList,"[s]hi[/s]","<s>hi</s>")
  msgList = addMEPair(msgList,"[c]hi[/c]","[c]hi[/c]")
  msgList = addMEPair(msgList,"[b]hi[/i]","[b]hi[/i]")
  msgList = addMEPair(msgList,"[/b]hi[b]","[/b]hi[b]")
  msgList = addMEPair(msgList,"[/b]hi[/b]","[/b]hi[/b]")
  msgList = addMEPair(msgList,"[/b]hi","[/b]hi")
  msgList = addMEPair(msgList,"[code]hi[/code]","<span class='codequotes'>hi</span>")
  msgList = addMEPair(msgList,"[code][b]hi[/b][/code]","<span class='codequotes'>[b]hi[/b]</span>")
  msgList = addMEPair(msgList,"[quote]hi[/quote]","<span class='postQuote'>hi</span>")
  msgList = addMEPair(msgList,"[quote][b]hi[/b][/quote]","<span class='postQuote'><b>hi</b></span>")
  msgList = addMEPair(msgList,"[quote][b]h[/b][/quote]","<span class='postQuote'><b>h</b></span>")
  msgList = addMEPair(msgList,"[quote][b][/b][/quote]","<span class='postQuote'><b></b></span>")

  t.Log("Testing bbcode_full_parse")
  for _, item := range msgList {
    t.Log("Testing string '"+item.Msg+"'")
    res = bbcode_full_parse(item.Msg).(string)
    if res != item.Expects {
      t.Error("Bad output:","'"+res+"'")
      t.Error("Expected:",item.Expects)
    }
  }

  var msg, expects string
  var err error

  msg = "[rand][/rand]"
  expects = "<span style='color: red;'>[Invalid Number]</span>[rand][/rand]"
  t.Log("Testing string '"+msg+"'")
  res = bbcode_full_parse(msg).(string)
  if res != expects {
    t.Error("Bad output:","'"+res+"'")
    t.Error("Expected:",expects)
  }

  msg = "[rand]-1[/rand]"
  expects = "<span style='color: red;'>[No Negative Numbers]</span>[rand]-1[/rand]"
  t.Log("Testing string '"+msg+"'")
  res = bbcode_full_parse(msg).(string)
  if res != expects {
    t.Error("Bad output:","'"+res+"'")
    t.Error("Expected:",expects)
  }

  var conv int
  msg = "[rand]1[/rand]"
  t.Log("Testing string '"+msg+"'")
  res = bbcode_full_parse(msg).(string)
  conv, err = strconv.Atoi(res)
  if err != nil && (conv > 1 || conv < 0) {
    t.Error("Bad output:","'"+res+"'")
    t.Error("Expected a number in the range 0-1")
  }

  t.Log("Testing bbcode_regex_parse")
  for _, item := range msgList {
    t.Log("Testing string '"+item.Msg+"'")
    res = bbcode_regex_parse(item.Msg).(string)
    if res != item.Expects {
      t.Error("Bad output:","'"+res+"'")
      t.Error("Expected:",item.Expects)
    }
  }
}

func TestMarkdownRender(t *testing.T) {
  var res string
  var msgList []ME_Pair
  msgList = addMEPair(msgList,"hi","hi")
  msgList = addMEPair(msgList,"**hi**","<b>hi</b>")
  msgList = addMEPair(msgList,"_hi_","<u>hi</u>")
  msgList = addMEPair(msgList,"*hi*","<i>hi</i>")
  msgList = addMEPair(msgList,"~hi~","<s>hi</s>")
  msgList = addMEPair(msgList,"*hi**","<i>hi</i>*")
  msgList = addMEPair(msgList,"**hi***","<b>hi</b>*")
  msgList = addMEPair(msgList,"**hi*","*<i>hi</i>")
  msgList = addMEPair(msgList,"***hi***","*<b><i>hi</i></b>")
  msgList = addMEPair(msgList,"\\*hi\\*","*hi*")
  msgList = addMEPair(msgList,"*~hi~*","<i><s>hi</s></i>")
  msgList = addMEPair(msgList,"**","**")
  msgList = addMEPair(msgList,"***","***")
  msgList = addMEPair(msgList,"****","****")
  msgList = addMEPair(msgList,"*****","*****")
  msgList = addMEPair(msgList,"******","******")
  msgList = addMEPair(msgList,"*******","*******")
  msgList = addMEPair(msgList,"~~","~~")
  msgList = addMEPair(msgList,"~~~","~~~")
  msgList = addMEPair(msgList,"~~~~","~~~~")
  msgList = addMEPair(msgList,"~~~~~","~~~~~")
  msgList = addMEPair(msgList,"__","__")
  msgList = addMEPair(msgList,"___","___")
  msgList = addMEPair(msgList,"_ _","<u> </u>")
  msgList = addMEPair(msgList,"* *","<i> </i>")
  msgList = addMEPair(msgList,"** **","<b> </b>")
  msgList = addMEPair(msgList,"*** ***","<b><i> </i></b>")

  for _, item := range msgList {
    t.Log("Testing string '"+item.Msg+"'")
    res = markdown_parse(item.Msg).(string)
    if res != item.Expects {
      t.Error("Bad output:","'"+res+"'")
      t.Error("Expected:",item.Expects)
    }
  }
}
