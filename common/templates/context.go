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
	VarHolder    string
	HoldReflect  reflect.Value
	TemplateName string
	LoopDepth    int
	OutBuf       *[]OutBufferFrame
}

func (con *CContext) Push(nType string, body string) (index int) {
	*con.OutBuf = append(*con.OutBuf, OutBufferFrame{body, nType, con.TemplateName, nil, nil})
	return len(*con.OutBuf) - 1
}

func (con *CContext) PushText(body string, fragIndex int, fragOutIndex int) (index int) {
	*con.OutBuf = append(*con.OutBuf, OutBufferFrame{body, "text", con.TemplateName, fragIndex, fragOutIndex})
	return len(*con.OutBuf) - 1
}

func (con *CContext) PushPhrase(body string, langIndex int) (index int) {
	*con.OutBuf = append(*con.OutBuf, OutBufferFrame{body, "lang", con.TemplateName, langIndex, nil})
	return len(*con.OutBuf) - 1
}

func (con *CContext) StartLoop(body string) (index int) {
	con.LoopDepth++
	return con.Push("startloop", body)
}

func (con *CContext) EndLoop(body string) (index int) {
	return con.Push("endloop", body)
}

func (con *CContext) StartTemplate(body string) (index int) {
	*con.OutBuf = append(*con.OutBuf, OutBufferFrame{body, "starttemplate", con.TemplateName, nil, nil})
	return len(*con.OutBuf) - 1
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
