package tmpl

import (
	"reflect"
)

// For use in generated code
type FragLite struct {
	Body string
}

type Fragment struct {
	Body         string
	TemplateName string
	Index        int
	Seen         bool
}

type OutBufferFrame struct {
	Body         string
	Type         string
	TemplateName string
	Extra        interface{}
	Extra2       interface{}
}

type CContext struct {
	RootHolder   string
	VarHolder    string
	HoldReflect  reflect.Value
	RootTemplateName string
	TemplateName string
	LoopDepth    int
	OutBuf       *[]OutBufferFrame
}

func (con *CContext) Push(nType string, body string) (index int) {
	*con.OutBuf = append(*con.OutBuf, OutBufferFrame{body, nType, con.TemplateName, nil, nil})
	return con.LastBufIndex()
}

func (con *CContext) PushText(body string, fragIndex int, fragOutIndex int) (index int) {
	*con.OutBuf = append(*con.OutBuf, OutBufferFrame{body, "text", con.TemplateName, fragIndex, fragOutIndex})
	return con.LastBufIndex()
}

func (con *CContext) PushPhrase(langIndex int) (index int) {
	*con.OutBuf = append(*con.OutBuf, OutBufferFrame{"", "lang", con.TemplateName, langIndex, nil})
	return con.LastBufIndex()
}

func (con *CContext) PushPhrasef(langIndex int, args string) (index int) {
	*con.OutBuf = append(*con.OutBuf, OutBufferFrame{args, "langf", con.TemplateName, langIndex, nil})
	return con.LastBufIndex()
}

func (con *CContext) StartLoop(body string) (index int) {
	con.LoopDepth++
	return con.Push("startloop", body)
}

func (con *CContext) EndLoop(body string) (index int) {
	return con.Push("endloop", body)
}

func (con *CContext) StartTemplate(body string) (index int) {
	return con.addFrame(body, "starttemplate", nil, nil)
}

func (con *CContext) EndTemplate(body string) (index int) {
	return con.Push("endtemplate", body)
}

func (con *CContext) AttachVars(vars string, index int) {
	outBuf := *con.OutBuf
	node := outBuf[index]
	if node.Type != "starttemplate" && node.Type != "startloop" {
		panic("not a starttemplate node")
	}
	node.Body += vars
	outBuf[index] = node
}

func (con *CContext) addFrame(body string, ftype string, extra1 interface{}, extra2 interface{}) (index int) {
	*con.OutBuf = append(*con.OutBuf, OutBufferFrame{body, ftype, con.TemplateName, extra1, extra2})
	return con.LastBufIndex()
}

func (con *CContext) LastBufIndex() int {
	return len(*con.OutBuf) - 1
}

func (con *CContext) DiscardAndAfter(index int) {
	outBuf := *con.OutBuf
	if len(outBuf) <= index {
		return
	}
	if index == 0 {
		outBuf = nil
	} else {
		outBuf = outBuf[:index]
	}
	*con.OutBuf = outBuf
}
