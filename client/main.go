package main

import (
	"../common"
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
}
