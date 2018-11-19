package tmpl

import (
	"reflect"
)

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
