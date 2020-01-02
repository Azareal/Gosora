package qgen

import (
	"strings"
	"testing"
)

type MT struct {
	Type     int
	Contents string
}

func expectTokens(t *testing.T, whs []DBWhere, tokens ...MT) {
	i := 0
	for _, wh := range whs {
		for _, expr := range wh.Expr {
			if expr.Type != tokens[i].Type || expr.Contents != tokens[i].Contents {
				t.Fatalf("token mismatch: %+v - %+v\n", expr, tokens[i])
			}
			i++
		}
	}
}

func TestProcessWhere(t *testing.T) {
	whs := processWhere("uid = ?")
	expectTokens(t, whs, MT{TokenColumn, "uid"}, MT{TokenOp, "="}, MT{TokenSub, "?"})
	whs = processWhere("uid = 1")
	expectTokens(t, whs, MT{TokenColumn, "uid"}, MT{TokenOp, "="}, MT{TokenNumber, "1"})
	whs = processWhere("uid = 0")
	expectTokens(t, whs, MT{TokenColumn, "uid"}, MT{TokenOp, "="}, MT{TokenNumber, "0"})
	whs = processWhere("uid = '1'")
	expectTokens(t, whs, MT{TokenColumn, "uid"}, MT{TokenOp, "="}, MT{TokenString, "1"})
	whs = processWhere("uid = ''")
	expectTokens(t, whs, MT{TokenColumn, "uid"}, MT{TokenOp, "="}, MT{TokenString, ""})
	whs = processWhere("uid = '")
	expectTokens(t, whs, MT{TokenColumn, "uid"}, MT{TokenOp, "="}, MT{TokenString, ""})

	whs = processWhere("uid=?")
	expectTokens(t, whs, MT{TokenColumn, "uid"}, MT{TokenOp, "="}, MT{TokenSub, "?"})
	whs = processWhere("uid=1")
	expectTokens(t, whs, MT{TokenColumn, "uid"}, MT{TokenOp, "="}, MT{TokenNumber, "1"})
	whs = processWhere("uid=0")
	expectTokens(t, whs, MT{TokenColumn, "uid"}, MT{TokenOp, "="}, MT{TokenNumber, "0"})
	whs = processWhere("uid='1'")
	expectTokens(t, whs, MT{TokenColumn, "uid"}, MT{TokenOp, "="}, MT{TokenString, "1"})

	whs = processWhere("uid")
	expectTokens(t, whs, MT{TokenColumn, "uid"})
}

func TestMySQLBuildWhere(t *testing.T) {
	a := &MysqlAdapter{Name: "mysql", Buffer: make(map[string]DBStmt)}
	reap := func(wh, ex string) {
		sb := &strings.Builder{}
		a.buildWhere(wh, sb)
		res := sb.String()
		if res != ex {
			t.Fatalf("build where mismatch: '%+v' - '%+v'\n", ex, res)
		}
	}
	reap("uid = 0", " WHERE `uid`= 0 ")
	reap("uid = '0'", " WHERE `uid`= '0'")
	reap("uid=0", " WHERE `uid`= 0 ")
}
