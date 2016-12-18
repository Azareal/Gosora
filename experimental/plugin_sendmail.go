package main

import "io"
import "os/exec"
import "errors"
import "runtime"

/*
	Sending emails in a way you really shouldn't be sending them.
	This method doesn't require a SMTP server, but has higher chances of an email being rejected or being seen as spam. Use at your own risk. Only for Linux as Windows doesn't have Sendmail.
*/
func init() {
	plugins["sendmail"] = Plugin{"sendmail","Sendmail","Azareal","http://github.com/Azareal","",false,"Linux Only","",init_sendmail,activate_sendmail,deactivate_sendmail}
}

func init_sendmail() {
	add_hook("email_send_intercept", send_sendmail)
}

// Sendmail is only available on Linux
func activate_sendmail() error {
	if !enable_emails {
		return errors.New("You have emails disabled in your configuration file")
	}
	if runtime.GOOS != "linux" {
		return errors.New("This plugin only supports Linux")
	}
	return nil
}

func deactivate_sendmail() {
	remove_vhook("email_send_intercept")
}

func send_sendmail(data ...interface{}) interface{} {
	to := data[0].(string)
	subject := data[1].(string)
	body := data[2].(string)
	
	msg := "From: " + site_email + "\n"
	msg += "To: " + to + "\n"
	msg += "Subject: " + subject + "\n\n"
	msg += body + "\n"
	
	sendmail := exec.Command("/usr/sbin/sendmail","-t","-i")
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
	
	err = sendmail.Wait()
	if err != nil {
		return false
	}
	return true
}