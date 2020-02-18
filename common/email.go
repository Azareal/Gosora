package common

import (
	"crypto/tls"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"

	p "github.com/Azareal/Gosora/common/phrases"
)

func SendActivationEmail(username, email, token string) error {
	schema := "http"
	if Config.SslSchema {
		schema += "s"
	}
	// TODO: Move these to the phrase system
	subject := "Account Activation - " + Site.Name
	msg := "Dear " + username + ", to complete your registration on our forums, we need you to validate your email, so that we can confirm that this email actually belongs to you.\n\nClick on the following link to do so. " + schema + "://" + Site.URL + "/user/edit/token/" + token + "\n\nIf you haven't created an account here, then please feel free to ignore this email.\nWe're sorry for the inconvenience this may have caused."
	return SendEmail(email, subject, msg)
}

func SendValidationEmail(username, email, token string) error {
	schema := "http"
	if Config.SslSchema {
		schema += "s"
	}
	r := func(body *string) func(name, val string) {
		return func(name, val string) {
			*body = strings.Replace(*body, "{{"+name+"}}", val, -1)
		}
	}
	subject := p.GetAccountPhrase("ValidateEmailSubject")
	r1 := r(&subject)
	r1("name", Site.Name)
	body := p.GetAccountPhrase("ValidateEmailBody")
	r2 := r(&body)
	r2("username", username)
	r2("schema", schema)
	r2("url", Site.URL)
	r2("token", token)
	return SendEmail(email, subject, body)
}

// TODO: Refactor this
func SendEmail(email, subject, msg string) (err error) {
	// This hook is useful for plugin_sendmail or for testing tools. Possibly to hook it into some sort of mail server?
	ret, hasHook := GetHookTable().VhookNeedHook("email_send_intercept", email, subject, msg)
	if hasHook {
		return ret.(error)
	}

	from := mail.Address{"", Site.Email}
	to := mail.Address{"", email}
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subject

	body := ""
	for k, v := range headers {
		body += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	body += "\r\n" + msg

	var c *smtp.Client
	var conn *tls.Conn
	if Config.SMTPEnableTLS {
		tlsconfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         Config.SMTPServer,
		}
		conn, err = tls.Dial("tcp", Config.SMTPServer+":"+Config.SMTPPort, tlsconfig)
		if err != nil {
			LogWarning(err)
			return err
		}
		c, err = smtp.NewClient(conn, Config.SMTPServer)
	} else {
		c, err = smtp.Dial(Config.SMTPServer + ":" + Config.SMTPPort)
	}
	if err != nil {
		LogWarning(err)
		return err
	}

	if Config.SMTPUsername != "" {
		auth := smtp.PlainAuth("", Config.SMTPUsername, Config.SMTPPassword, Config.SMTPServer)
		err = c.Auth(auth)
		if err != nil {
			LogWarning(err)
			return err
		}
	}
	if err = c.Mail(from.Address); err != nil {
		LogWarning(err)
		return err
	}
	if err = c.Rcpt(to.Address); err != nil {
		LogWarning(err)
		return err
	}

	w, err := c.Data()
	if err != nil {
		LogWarning(err)
		return err
	}
	_, err = w.Write([]byte(body))
	if err != nil {
		LogWarning(err)
		return err
	}
	if err = w.Close(); err != nil {
		LogWarning(err)
		return err
	}
	if err = c.Quit(); err != nil {
		LogWarning(err)
		return err
	}

	return nil
}
