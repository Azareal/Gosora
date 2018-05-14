package main

import (
	"bytes"

	"../common"
	"../common/alerts"
	"../tmpl_client"
	"github.com/gopherjs/gopherjs/js"
)

func main() {
	js.Global.Set("weakPassword", func(password string, username string, email string) string {
		err := common.WeakPassword(password, username, email)
		if err != nil {
			return err.Error()
		}
		return ""
	})

	js.Global.Set("renderAlert", func(asid int, path string, msg string, avatar string) string {
		var buf bytes.Buffer
		alertItem := alerts.AlertItem{asid, path, msg, avatar}
		err := tmpl.Template_alert(alertItem, &buf)
		if err != nil {
			println(err.Error())
		}
		return string(buf.Bytes())
	})
}
