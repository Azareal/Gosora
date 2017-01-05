package main
import "log"
import "fmt"
import "time"
import "os"
import "encoding/base64"
import "crypto/rand"
import "net/smtp"

// Generate a cryptographically secure set of random bytes..
func GenerateSafeString(length int) (string, error) {
	rb := make([]byte,length)
	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(rb), nil
}

func relative_time(in string) (string, error) {
	layout := "2006-01-02 15:04:05"
	t, err := time.Parse(layout, in)
	if err != nil {
		return "", err
	}
	
	diff := time.Since(t)
	hours := diff.Hours()
	seconds := diff.Seconds()
	switch {
		case (hours / 24) > 7:
			return t.Format("Mon Jan 2 2006"), err
		case int(hours / 24) == 1:
			return "1 day ago", err
		case int(hours / 24) > 1:
			return fmt.Sprintf("%d days ago", int(hours / 24)), err
		case seconds <= 1:
			return "a moment ago", err
		case seconds < 60:
			return fmt.Sprintf("%d seconds ago", int(seconds)), err
		case seconds < 120:
			return "a minute ago", err
		case seconds < 3600:
			return fmt.Sprintf("%d minutes ago", int(seconds / 60)), err
		case seconds < 7200:
			return "an hour ago", err
		default:
			return fmt.Sprintf("%d hours ago", int(seconds / 60 / 60)), err
	}
}

func SendEmail(email string, subject string, msg string) bool {
	// This hook is useful for plugin_sendmail or for testing tools. Possibly to hook it into some sort of mail server?
	if vhooks["email_send_intercept"] != nil {
		return vhooks["email_send_intercept"](email, subject, msg).(bool)
	}
	body := "Subject: " + subject + "\n\n" + msg + "\n"
	
	con, err := smtp.Dial(smtp_server)
	if err != nil {
		return false
	}
	
	err = con.Mail(site_email)
	if err != nil {
		return false
	}
	err = con.Rcpt(email)
	if err != nil {
		return false
	}
	
	email_data, err := con.Data()
	if err != nil {
		return false
	}
	_, err = fmt.Fprintf(email_data, body)
	if err != nil {
		return false
	}
	
	err = email_data.Close()
	if err != nil {
		return false
	}
	err = con.Quit()
	if err != nil {
		return false
	}
	return true
}

func write_file(name string, content string) {
	f, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	
	_, err = f.WriteString(content)
	if err != nil {
		log.Fatal(err)
	}
	f.Sync()
	f.Close()
}
