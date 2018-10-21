package common

import (
	"fmt"
	"net/smtp"
)

type Email struct {
	UserID    int
	Email     string
	Validated bool
	Primary   bool
	Token     string
}

func SendValidationEmail(username string, email string, token string) bool {
	var schema = "http"
	if Site.EnableSsl {
		schema += "s"
	}

	// TODO: Move these to the phrase system
	subject := "Validate Your Email @ " + Site.Name
	msg := "Dear " + username + ", following your registration on our forums, we ask you to validate your email, so that we can confirm that this email actually belongs to you.\n\nClick on the following link to do so. " + schema + "://" + Site.URL + "/user/edit/token/" + token + "\n\nIf you haven't created an account here, then please feel free to ignore this email.\nWe're sorry for the inconvenience this may have caused."
	return SendEmail(email, subject, msg)
}

// TODO: Refactor this
// TODO: Add support for TLS
func SendEmail(email string, subject string, msg string) bool {
	// This hook is useful for plugin_sendmail or for testing tools. Possibly to hook it into some sort of mail server?
	ret, hasHook := GetHookTable().VhookNeedHook("email_send_intercept", email, subject, msg)
	if hasHook {
		return ret.(bool)
	}
	body := "Subject: " + subject + "\n\n" + msg + "\n"

	con, err := smtp.Dial(Config.SMTPServer + ":" + Config.SMTPPort)
	if err != nil {
		return false
	}

	if Config.SMTPUsername != "" {
		auth := smtp.PlainAuth("", Config.SMTPUsername, Config.SMTPPassword, Config.SMTPServer)
		err = con.Auth(auth)
		if err != nil {
			return false
		}
	}

	err = con.Mail(Site.Email)
	if err != nil {
		return false
	}
	err = con.Rcpt(email)
	if err != nil {
		return false
	}

	emailData, err := con.Data()
	if err != nil {
		return false
	}
	_, err = fmt.Fprintf(emailData, body)
	if err != nil {
		return false
	}

	err = emailData.Close()
	if err != nil {
		return false
	}

	return con.Quit() == nil
}
