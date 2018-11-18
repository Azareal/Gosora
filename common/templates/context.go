package tmpl

import (
	"errors"
	"reflect"
)

type OutBufferFrame struct {
	Body         string
	Type         string
	TemplateName string
}

type CContext struct {
	VarHolder    string
	HoldReflect  reflect.Value
	TemplateName string
	OutBuf       *[]OutBufferFrame
}

func (con *CContext) Push(nType string, body string) {
	*con.OutBuf = append(*con.OutBuf, OutBufferFrame{body, nType, con.TemplateName})
}

func (con *CContext) GetLastType() string {
	outBuf := *con.OutBuf
	if len(outBuf) == 0 {
		return ""
	}
	return outBuf[len(outBuf)-1].Type
}

func (con *CContext) GetLastBody() string {
	outBuf := *con.OutBuf
	if len(outBuf) == 0 {
		return ""
	}
	return outBuf[len(outBuf)-1].Body
}

func (con *CContext) SetLastBody(newBody string) error {
	outBuf := *con.OutBuf
	if len(outBuf) == 0 {
		return errors.New("outbuf is empty")
	}
	outBuf[len(outBuf)-1].Body = newBody
	return nil
}

func (con *CContext) GetLastTemplate() string {
	outBuf := *con.OutBuf
	if len(outBuf) == 0 {
		return ""
	}
	return outBuf[len(outBuf)-1].TemplateName
}
