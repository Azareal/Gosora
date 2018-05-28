/*
*
*	Utility Functions And Stuff
*	Copyright Azareal 2017 - 2018
*
 */
package common

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
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
func RelativeTimeFromString(in string) (string, error) {
	if in == "" {
		return "", nil
	}

	t, err := time.Parse("2006-01-02 15:04:05", in)
	if err != nil {
		return "", err
	}

	return RelativeTime(t), nil
}

// TODO: Write a test for this
func RelativeTime(t time.Time) string {
	diff := time.Since(t)
	hours := diff.Hours()
	seconds := diff.Seconds()
	weeks := int(hours / 24 / 7)
	months := int(hours / 24 / 31)
	switch {
	case months > 3:
		if t.Year() != time.Now().Year() {
			//return t.Format("Mon Jan 2 2006")
			return t.Format("Jan 2 2006")
		}
		return t.Format("Jan 2")
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
func ConvertByteUnit(bytes float64) (float64, string) {
	switch {
	case bytes >= float64(Petabyte):
		return bytes / float64(Petabyte), "PB"
	case bytes >= float64(Terabyte):
		return bytes / float64(Terabyte), "TB"
	case bytes >= float64(Gigabyte):
		return bytes / float64(Gigabyte), "GB"
	case bytes >= float64(Megabyte):
		return bytes / float64(Megabyte), "MB"
	case bytes >= float64(Kilobyte):
		return bytes / float64(Kilobyte), "KB"
	default:
		return bytes, " bytes"
	}
}

// TODO: Write a test for this
func ConvertByteInUnit(bytes float64, unit string) (count float64) {
	switch unit {
	case "PB":
		count = bytes / float64(Petabyte)
	case "TB":
		count = bytes / float64(Terabyte)
	case "GB":
		count = bytes / float64(Gigabyte)
	case "MB":
		count = bytes / float64(Megabyte)
	case "KB":
		count = bytes / float64(Kilobyte)
	default:
		count = 0.1
	}

	if count < 0.1 {
		count = 0.1
	}
	return
}

// TODO: Write a test for this
// TODO: Re-add T as int64
func ConvertUnit(num int) (int, string) {
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
// TODO: Re-add quadrillion as int64
// TODO: Re-add trillion as int64
func ConvertFriendlyUnit(num int) (int, string) {
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

// TODO: Make slugs optional for certain languages across the entirety of Gosora?
// TODO: Let plugins replace NameToSlug and the URL building logic with their own
func NameToSlug(name string) (slug string) {
	// TODO: Do we want this reliant on config file flags? This might complicate tests and oddball uses
	if !Config.BuildSlugs {
		return ""
	}
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

// TODO: Write a test for this
func WeakPassword(password string, username string, email string) error {
	lowPassword := strings.ToLower(password)
	switch {
	case password == "":
		return errors.New("You didn't put in a password.")
	case strings.Contains(lowPassword, strings.ToLower(username)) && len(username) > 3:
		return errors.New("You can't use your username in your password.")
	case strings.Contains(lowPassword, strings.ToLower(email)):
		return errors.New("You can't use your email in your password.")
	case len(password) < 8:
		return errors.New("Your password needs to be at-least eight characters long")
	}

	if strings.Contains(lowPassword, "test") || /*strings.Contains(password,"123456") || */ strings.Contains(password, "123") || strings.Contains(lowPassword, "password") || strings.Contains(lowPassword, "qwerty") || strings.Contains(lowPassword, "fuck") || strings.Contains(lowPassword, "love") {
		return errors.New("You may not have 'test', '123', 'password', 'qwerty', 'love' or 'fuck' in your password")
	}

	var charMap = make(map[rune]int)
	var numbers, symbols, upper, lower int
	for _, char := range password {
		charItem, ok := charMap[char]
		if ok {
			charItem++
		} else {
			charItem = 1
		}
		charMap[char] = charItem

		if unicode.IsLetter(char) {
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

	if numbers == 0 {
		return errors.New("You don't have any numbers in your password")
	}
	if upper == 0 {
		return errors.New("You don't have any uppercase characters in your password")
	}
	if lower == 0 {
		return errors.New("You don't have any lowercase characters in your password")
	}
	if len(password) < 18 {
		if (len(password) / 2) > len(charMap) {
			return errors.New("You don't have enough unique characters in your password")
		}
	} else if (len(password) / 3) > len(charMap) {
		// Be a little lenient on the number of unique characters for long passwords
		return errors.New("You don't have enough unique characters in your password")
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
func WordCount(input string) (count int) {
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
func GetLevel(score int) (level int) {
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
func GetLevelScore(getLevel int) (score int) {
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
func GetLevels(maxLevel int) []float64 {
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

func BuildSlug(slug string, id int) string {
	if slug == "" || !Config.BuildSlugs {
		return strconv.Itoa(id)
	}
	return slug + "." + strconv.Itoa(id)
}
