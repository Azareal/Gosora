package main

import (
	"errors"
	"io"
	"os/exec"
	"runtime"

	"github.com/Azareal/Gosora/common"
)

/*
	Sending emails in a way you really shouldn't be sending them.
	This method doesn't require a SMTP server, but has higher chances of an email being rejected or being seen as spam. Use at your own risk. Only for Linux as Windows doesn't have Sendmail.
*/
func init() {
	// Don't bother registering this plugin on platforms other than Linux
	if runtime.GOOS != "linux" {
		return
	}
	common.Plugins.Add(&common.Plugin{UName: "sendmail", Name: "Sendmail", Author: "Azareal", URL: "http://github.com/Azareal", Tag: "Linux Only", Init: initSendmail, Activate: activateSendmail, Deactivate: deactivateSendmail})
}

func initSendmail(plugin *common.Plugin) error {
	plugin.AddHook("email_send_intercept", sendSendmail)
	return nil
}

// /usr/sbin/sendmail is only available on Linux
func activateSendmail(plugin *common.Plugin) error {
	if !common.Site.EnableEmails {
		return errors.New("You have emails disabled in your configuration file")
	}
	if runtime.GOOS != "linux" {
		return errors.New("This plugin only supports Linux")
	}
	return nil
}

func deactivateSendmail(plugin *common.Plugin) {
	plugin.RemoveHook("email_send_intercept", sendSendmail)
}

func sendSendmail(data ...interface{}) interface{} {
	to := data[0].(string)
	subject := data[1].(string)
	body := data[2].(string)

	msg := "From: " + common.Site.Email + "\n"
	msg += "To: " + to + "\n"
	msg += "Subject: " + subject + "\n\n"
	msg += body + "\n"

	sendmail := exec.Command("/usr/sbin/sendmail", "-t", "-i")
	stdin, err := sendmail.StdinPipe()
	if err != nil {
		return false // Possibly disable the plugin and show an error to the admin on the dashboard? Plugin log file?
	}

	err = sendmail.Start()
	if err != nil {
		return false
	}
	io.WriteString(stdin, msg)

	err = stdin.Close()
	if err != nil {
		return false
	}

	return sendmail.Wait() == nil
}
