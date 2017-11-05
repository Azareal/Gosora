/*
*
*	Utility Functions And Stuff
*	Copyright Azareal 2017 - 2018
*
 */
package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Version stores a Gosora version
type Version struct {
	Major int
	Minor int
	Patch int
	Tag   string
	TagID int
}

// TODO: Write a test for this
func (version *Version) String() (out string) {
	out = strconv.Itoa(version.Major) + "." + strconv.Itoa(version.Minor) + "." + strconv.Itoa(version.Patch)
	if version.Tag != "" {
		out += "-" + version.Tag
		if version.TagID != 0 {
			out += strconv.Itoa(version.TagID)
		}
	}
	return
}

// GenerateSafeString is for generating a cryptographically secure set of random bytes...
// TODO: Write a test for this
func GenerateSafeString(length int) (string, error) {
	rb := make([]byte, length)
	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(rb), nil
}

// TODO: Write a test for this
func relativeTimeFromString(in string) (string, error) {
	if in == "" {
		return "", nil
	}

	t, err := time.Parse("2006-01-02 15:04:05", in)
	if err != nil {
		return "", err
	}

	return relativeTime(t), nil
}

// TODO: Write a test for this
func relativeTime(t time.Time) string {
	diff := time.Since(t)
	hours := diff.Hours()
	seconds := diff.Seconds()
	weeks := int(hours / 24 / 7)
	months := int(hours / 24 / 31)
	switch {
	case months > 11:
		//return t.Format("Mon Jan 2 2006")
		return t.Format("Jan 2 2006")
	case months > 1:
		return fmt.Sprintf("%d months ago", months)
	case months == 1:
		return "a month ago"
	case weeks > 1:
		return fmt.Sprintf("%d weeks ago", weeks)
	case int(hours/24) == 7:
		return "a week ago"
	case int(hours/24) == 1:
		return "1 day ago"
	case int(hours/24) > 1:
		return fmt.Sprintf("%d days ago", int(hours/24))
	case seconds <= 1:
		return "a moment ago"
	case seconds < 60:
		return fmt.Sprintf("%d seconds ago", int(seconds))
	case seconds < 120:
		return "a minute ago"
	case seconds < 3600:
		return fmt.Sprintf("%d minutes ago", int(seconds/60))
	case seconds < 7200:
		return "an hour ago"
	default:
		return fmt.Sprintf("%d hours ago", int(seconds/60/60))
	}
}

// TODO: Write a test for this
func convertByteUnit(bytes float64) (float64, string) {
	switch {
	case bytes >= float64(petabyte):
		return bytes / float64(petabyte), "PB"
	case bytes >= float64(terabyte):
		return bytes / float64(terabyte), "TB"
	case bytes >= float64(gigabyte):
		return bytes / float64(gigabyte), "GB"
	case bytes >= float64(megabyte):
		return bytes / float64(megabyte), "MB"
	case bytes >= float64(kilobyte):
		return bytes / float64(kilobyte), "KB"
	default:
		return bytes, " bytes"
	}
}

// TODO: Write a test for this
func convertByteInUnit(bytes float64, unit string) (count float64) {
	switch unit {
	case "PB":
		count = bytes / float64(petabyte)
	case "TB":
		count = bytes / float64(terabyte)
	case "GB":
		count = bytes / float64(gigabyte)
	case "MB":
		count = bytes / float64(megabyte)
	case "KB":
		count = bytes / float64(kilobyte)
	default:
		count = 0.1
	}

	if count < 0.1 {
		count = 0.1
	}
	return
}

// TODO: Write a test for this
func convertUnit(num int) (int, string) {
	switch {
	case num >= 1000000000000:
		return num / 1000000000000, "T"
	case num >= 1000000000:
		return num / 1000000000, "B"
	case num >= 1000000:
		return num / 1000000, "M"
	case num >= 1000:
		return num / 1000, "K"
	default:
		return num, ""
	}
}

// TODO: Write a test for this
func convertFriendlyUnit(num int) (int, string) {
	switch {
	case num >= 1000000000000000:
		return 0, " quadrillion"
	case num >= 1000000000000:
		return 0, " trillion"
	case num >= 1000000000:
		return num / 1000000000, " billion"
	case num >= 1000000:
		return num / 1000000, " million"
	case num >= 1000:
		return num / 1000, " thousand"
	default:
		return num, ""
	}
}

func nameToSlug(name string) (slug string) {
	name = strings.TrimSpace(name)
	name = strings.Replace(name, "  ", " ", -1)

	for _, char := range name {
		if unicode.IsLower(char) || unicode.IsNumber(char) {
			slug += string(char)
		} else if unicode.IsUpper(char) {
			slug += string(unicode.ToLower(char))
		} else if unicode.IsSpace(char) {
			slug += "-"
		}
	}

	if slug == "" {
		slug = "untitled"
	}
	return slug
}

func SendEmail(email string, subject string, msg string) (res bool) {
	// This hook is useful for plugin_sendmail or for testing tools. Possibly to hook it into some sort of mail server?
	if vhooks["email_send_intercept"] != nil {
		return vhooks["email_send_intercept"](email, subject, msg).(bool)
	}
	body := "Subject: " + subject + "\n\n" + msg + "\n"

	con, err := smtp.Dial(config.SMTPServer + ":" + config.SMTPPort)
	if err != nil {
		return
	}

	if config.SMTPUsername != "" {
		auth := smtp.PlainAuth("", config.SMTPUsername, config.SMTPPassword, config.SMTPServer)
		err = con.Auth(auth)
		if err != nil {
			return
		}
	}

	err = con.Mail(site.Email)
	if err != nil {
		return
	}
	err = con.Rcpt(email)
	if err != nil {
		return
	}

	emailData, err := con.Data()
	if err != nil {
		return
	}
	_, err = fmt.Fprintf(emailData, body)
	if err != nil {
		return
	}

	err = emailData.Close()
	if err != nil {
		return
	}
	err = con.Quit()
	if err != nil {
		return
	}
	return true
}

// TODO: Write a test for this
func weakPassword(password string) error {
	if len(password) < 8 {
		return errors.New("your password needs to be at-least eight characters long")
	}
	var charMap = make(map[rune]int)
	var numbers /*letters, */, symbols, upper, lower int
	for _, char := range password {
		charItem, ok := charMap[char]
		if ok {
			charItem++
		} else {
			charItem = 1
		}
		charMap[char] = charItem

		if unicode.IsLetter(char) {
			//letters++
			if unicode.IsUpper(char) {
				upper++
			} else {
				lower++
			}
		} else if unicode.IsNumber(char) {
			numbers++
		} else {
			symbols++
		}
	}

	// TODO: Disable the linter on these and fix up the grammar
	if numbers == 0 {
		return errors.New("you don't have any numbers in your password")
	}
	/*if letters == 0 {
		return errors.New("You don't have any letters in your password.")
	}*/
	if upper == 0 {
		return errors.New("you don't have any uppercase characters in your password")
	}
	if lower == 0 {
		return errors.New("you don't have any lowercase characters in your password")
	}
	if (len(password) / 2) > len(charMap) {
		return errors.New("you don't have enough unique characters in your password")
	}

	if strings.Contains(strings.ToLower(password), "test") || /*strings.Contains(strings.ToLower(password),"123456") || */ strings.Contains(strings.ToLower(password), "123") || strings.Contains(strings.ToLower(password), "password") || strings.Contains(strings.ToLower(password), "qwerty") {
		return errors.New("you may not have 'test', '123', 'password' or 'qwerty' in your password")
	}
	return nil
}

// TODO: Write a test for this
func createFile(name string) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	return f.Close()
}

// TODO: Write a test for this
func writeFile(name string, content string) (err error) {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	_, err = f.WriteString(content)
	if err != nil {
		return err
	}
	err = f.Sync()
	if err != nil {
		return err
	}
	return f.Close()
}

// TODO: Write a test for this
func Stripslashes(text string) string {
	text = strings.Replace(text, "/", "", -1)
	return strings.Replace(text, "\\", "", -1)
}

// TODO: Write a test for this
func wordCount(input string) (count int) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0
	}
	var inSpace bool
	for _, value := range input {
		if unicode.IsSpace(value) {
			if !inSpace {
				inSpace = true
			}
		} else if inSpace {
			count++
			inSpace = false
		}
	}
	return count + 1
}

// TODO: Write a test for this
func getLevel(score int) (level int) {
	var base float64 = 25
	var current, prev float64
	var expFactor = 2.8

	for i := 1; ; i++ {
		_, bit := math.Modf(float64(i) / 10)
		if bit == 0 {
			expFactor += 0.1
		}
		current = base + math.Pow(float64(i), expFactor) + (prev / 3)
		prev = current
		if float64(score) < current {
			break
		}
		level++
	}
	return level
}

// TODO: Write a test for this
func getLevelScore(getLevel int) (score int) {
	var base float64 = 25
	var current, prev float64
	var level int
	expFactor := 2.8

	for i := 1; ; i++ {
		_, bit := math.Modf(float64(i) / 10)
		if bit == 0 {
			expFactor += 0.1
		}
		current = base + math.Pow(float64(i), expFactor) + (prev / 3)
		prev = current
		level++
		if level <= getLevel {
			break
		}
	}
	return int(math.Ceil(current))
}

// TODO: Write a test for this
func getLevels(maxLevel int) []float64 {
	var base float64 = 25
	var current, prev float64 // = 0
	var expFactor = 2.8
	var out []float64
	out = append(out, 0)

	for i := 1; i <= maxLevel; i++ {
		_, bit := math.Modf(float64(i) / 10)
		if bit == 0 {
			expFactor += 0.1
		}
		current = base + math.Pow(float64(i), expFactor) + (prev / 3)
		prev = current
		out = append(out, current)
	}
	return out
}

func buildSlug(slug string, id int) string {
	if slug == "" {
		return strconv.Itoa(id)
	}
	return slug + "." + strconv.Itoa(id)
}

// TODO: Make a store for this?
func addModLog(action string, elementID int, elementType string, ipaddress string, actorID int) (err error) {
	_, err = stmts.addModlogEntry.Exec(action, elementID, elementType, ipaddress, actorID)
	return err
}

// TODO: Make a store for this?
func addAdminLog(action string, elementID string, elementType int, ipaddress string, actorID int) (err error) {
	_, err = stmts.addAdminlogEntry.Exec(action, elementID, elementType, ipaddress, actorID)
	return err
}
