package common

import (
	"crypto/tls"
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
	subject := "Validate Your Email @ " + Site.Name
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
	body := "Subject: " + subject + "\n\n" + msg + "\n"

	var c *smtp.Client
	if Config.SMTPEnableTLS {
		tlsconfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         Config.SMTPServer,
		}
		conn, err := tls.Dial("tcp", Config.SMTPServer+":"+Config.SMTPPort, tlsconfig)
		if err != nil {
			return err
		}
		c, err = smtp.NewClient(conn, Config.SMTPServer)
		if err != nil {
			return err
		}
	} else {
		c, err = smtp.Dial(Config.SMTPServer + ":" + Config.SMTPPort)
		if err != nil {
			return err
		}
	}

	if Config.SMTPUsername != "" {
		auth := smtp.PlainAuth("", Config.SMTPUsername, Config.SMTPPassword, Config.SMTPServer)
		err = c.Auth(auth)
		if err != nil {
			return err
		}
	}

	err = c.Mail(Site.Email)
	if err != nil {
		return err
	}
	err = c.Rcpt(email)
	if err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(body))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
