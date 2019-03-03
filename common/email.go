package common

import (
	"crypto/tls"
	"fmt"
	"net/mail"
	"net/smtp"
)

type Email struct {
	UserID    int
	Email     string
	Validated bool
	Primary   bool
	Token     string
}

func SendValidationEmail(username string, email string, token string) error {
	var schema = "http"
	if Site.EnableSsl {
		schema += "s"
	}

	// TODO: Move these to the phrase system
	subject := "Validate Your Email - " + Site.Name
	msg := "Dear " + username + ", following your registration on our forums, we ask you to validate your email, so that we can confirm that this email actually belongs to you.\n\nClick on the following link to do so. " + schema + "://" + Site.URL + "/user/edit/token/" + token + "\n\nIf you haven't created an account here, then please feel free to ignore this email.\nWe're sorry for the inconvenience this may have caused."
	return SendEmail(email, subject, msg)
}

// TODO: Refactor this
func SendEmail(email string, subject string, msg string) (err error) {
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
	if Config.SMTPEnableTLS {
		tlsconfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         Config.SMTPServer,
		}
		conn, err := tls.Dial("tcp", Config.SMTPServer+":"+Config.SMTPPort, tlsconfig)
		if err != nil {
			LogWarning(err)
			return err
		}
		c, err = smtp.NewClient(conn, Config.SMTPServer)
		if err != nil {
			LogWarning(err)
			return err
		}
	} else {
		c, err = smtp.Dial(Config.SMTPServer + ":" + Config.SMTPPort)
		if err != nil {
			LogWarning(err)
			return err
		}
	}

	if Config.SMTPUsername != "" {
		auth := smtp.PlainAuth("", Config.SMTPUsername, Config.SMTPPassword, Config.SMTPServer)
		err = c.Auth(auth)
		if err != nil {
			LogWarning(err)
			return err
		}
	}

	err = c.Mail(from.Address)
	if err != nil {
		LogWarning(err)
		return err
	}
	err = c.Rcpt(to.Address)
	if err != nil {
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

	err = w.Close()
	if err != nil {
		LogWarning(err)
		return err
	}
	err = c.Quit()
	if err != nil {
		LogWarning(err)
		return err
	}

	return nil
}
