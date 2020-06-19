/*
*
*	Utility Functions And Stuff
*	Copyright Azareal 2017 - 2020
*
 */
package common

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
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
func (ver *Version) String() (out string) {
	out = strconv.Itoa(ver.Major) + "." + strconv.Itoa(ver.Minor) + "." + strconv.Itoa(ver.Patch)
	if ver.Tag != "" {
		out += "-" + ver.Tag
		if ver.TagID != 0 {
			out += strconv.Itoa(ver.TagID)
		}
	}
	return
}

// GenerateSafeString is for generating a cryptographically secure set of random bytes which is base64 encoded and safe for URLs
// TODO: Write a test for this
func GenerateSafeString(len int) (string, error) {
	rb := make([]byte, len)
	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(rb), nil
}

// GenerateStd32SafeString is for generating a cryptographically secure set of random bytes which is base32 encoded
// ? - Safe for URLs? Mostly likely due to the small range of characters
func GenerateStd32SafeString(len int) (string, error) {
	rb := make([]byte, len)
	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(rb), nil
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
	secs := diff.Seconds()
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
	case secs <= 1:
		return "a moment ago"
	case secs < 60:
		return fmt.Sprintf("%d seconds ago", int(secs))
	case secs < 120:
		return "a minute ago"
	case secs < 3600:
		return fmt.Sprintf("%d minutes ago", int(secs/60))
	case secs < 7200:
		return "an hour ago"
	}
	return fmt.Sprintf("%d hours ago", int(secs/60/60))
}

// TODO: Finish this faster and more localised version of RelativeTime
/*
// TODO: Write a test for this
// ! Experimental
func RelativeTimeBytes(t time.Time, lang int) []byte {
	diff := time.Since(t)
	hours := diff.Hours()
	secs := diff.Seconds()
	weeks := int(hours / 24 / 7)
	months := int(hours / 24 / 31)
	switch {
	case months > 3:
		if t.Year() != time.Now().Year() {
			return []byte(t.Format(phrases.RTime.MultiYear(lang)))
		}
		return []byte(t.Format(phrases.RTime.SingleYear(lang)))
	case months > 1:
		return phrases.RTime.Months(lang, months)
	case months == 1:
		return phrases.RTime.Month(lang)
	case weeks > 1:
		return phrases.RTime.Weeks(lang, weeks)
	case int(hours/24) == 7:
		return phrases.RTime.Week(lang)
	case int(hours/24) == 1:
		return phrases.RTime.Day(lang)
	case int(hours/24) > 1:
		return phrases.RTime.Days(lang, int(hours/24))
	case secs <= 1:
		return phrases.RTime.Moment(lang)
	case secs < 60:
		return phrases.RTime.Seconds(lang, int(secs))
	case secs < 120:
		return phrases.RTime.Minute(lang)
	case secs < 3600:
		return phrases.RTime.Minutes(lang, int(secs/60))
	case secs < 7200:
		return phrases.RTime.Hour(lang)
	}
	return phrases.RTime.Hours(lang, int(secs/60/60))
}
*/

var pMs = 1000
var pSec = pMs * 1000
var pMin = pSec * 60
var pHour = pMin * 60
var pDay = pHour * 24

func ConvertPerfUnit(quan float64) (out float64, unit string) {
	f := func() (float64, string) {
		switch {
		case quan >= float64(pDay):
			return quan / float64(pDay), "d"
		case quan >= float64(pHour):
			return quan / float64(pHour), "h"
		case quan >= float64(pMin):
			return quan / float64(pMin), "m"
		case quan >= float64(pSec):
			return quan / float64(pSec), "s"
		case quan >= float64(pMs):
			return quan / float64(pMs), "ms"
		}
		return quan, "μs"
	}
	out, unit = f()
	return math.Ceil(out), unit
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
	}
	return bytes, " bytes"
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
// TODO: Localise this?
func FriendlyUnitToBytes(quantity int, unit string) (bytes int, err error) {
	switch unit {
	case "PB":
		bytes = quantity * Petabyte
	case "TB":
		bytes = quantity * Terabyte
	case "GB":
		bytes = quantity * Gigabyte
	case "MB":
		bytes = quantity * Megabyte
	case "KB":
		bytes = quantity * Kilobyte
	case "":
		// Do nothing
	default:
		return bytes, errors.New("Unknown unit")
	}
	return bytes, nil
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
	}
	return num, ""
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
	}
	return num, ""
}

// TODO: Make slugs optional for certain languages across the entirety of Gosora?
// TODO: Let plugins replace NameToSlug and the URL building logic with their own
/*func NameToSlug(name string) (slug string) {
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
}*/

// TODO: Make slugs optional for certain languages across the entirety of Gosora?
// TODO: Let plugins replace NameToSlug and the URL building logic with their own
func NameToSlug(name string) (slug string) {
	// TODO: Do we want this reliant on config file flags? This might complicate tests and oddball uses
	if !Config.BuildSlugs {
		return ""
	}
	name = strings.TrimSpace(name)
	name = strings.Replace(name, "  ", " ", -1)

	var sb strings.Builder
	for _, char := range name {
		if unicode.IsLower(char) || unicode.IsNumber(char) {
			sb.WriteRune(char)
		} else if unicode.IsUpper(char) {
			sb.WriteRune(unicode.ToLower(char))
		} else if unicode.IsSpace(char) {
			sb.WriteByte('-')
		}
	}

	if sb.Len() == 0 {
		return "untitled"
	}
	return sb.String()
}

// TODO: Write a test for this
func HasSuspiciousEmail(email string) bool {
	if email == "" {
		return false
	}
	lowEmail := strings.ToLower(email)
	// TODO: Use a more flexible blacklist, perhaps with a similar mechanism to the HTML tag registration system in PreparseMessage()
	if !strings.Contains(lowEmail, "@") || strings.Contains(lowEmail, "casino") || strings.Contains(lowEmail, "viagra") || strings.Contains(lowEmail, "pharma") || strings.Contains(lowEmail, "pill") {
		return true
	}

	var dotCount, shortBits, currentSegmentLength int
	for _, char := range lowEmail {
		if char == '.' {
			dotCount++
			if currentSegmentLength < 3 {
				shortBits++
			}
			currentSegmentLength = 0
		} else {
			currentSegmentLength++
		}
	}

	return dotCount > 7 || shortBits > 2
}

func unmarshalJsonFile(name string, in interface{}) error {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, in)
}

func unmarshalJsonFileIgnore404(name string, in interface{}) error {
	data, err := ioutil.ReadFile(name)
	if err == os.ErrPermission || err == os.ErrClosed {
		return err
	} else if err != nil {
		return nil
	}
	return json.Unmarshal(data, in)
}

func CanonEmail(email string) string {
	email = strings.ToLower(email)

	// Gmail emails are equivalent without the dots
	espl := strings.Split(email, "@")
	if len(espl) >= 2 && espl[1] == "gmail.com" {
		return strings.Replace(espl[0], ".", "", -1) + "@" + espl[1]
	}

	return email
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
func writeFile(name, content string) (err error) {
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

// The word counter might run into problems with some languages where words aren't as obviously demarcated, I would advise turning it off in those cases, or if it becomes annoying in general, really.
func WordCount(input string) (count int) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0
	}

	var inSpace bool
	for _, value := range input {
		if unicode.IsSpace(value) || unicode.IsPunct(value) {
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
	var current float64
	var expFactor = 2.8

	for i := 1; i <= getLevel; i++ {
		_, bit := math.Modf(float64(i) / 10)
		if bit == 0 {
			expFactor += 0.1
		}
		current = base + math.Pow(float64(i), expFactor) + (current / 3)
		//fmt.Println("level: ", i)
		//fmt.Println("current: ", current)
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

// TODO: Write a test for this
// SanitiseSingleLine is a generic function for escaping html entities and removing silly characters from usernames and topic titles. It also strips newline characters
func SanitiseSingleLine(in string) string {
	in = strings.Replace(in, "\n", "", -1)
	in = strings.Replace(in, "\r", "", -1)
	return SanitiseBody(in)
}

// TODO: Write a test for this
// TODO: Add more strange characters
// TODO: Strip all sub-32s minus \r and \n?
// SanitiseBody is the same as SanitiseSingleLine, but it doesn't strip newline characters
func SanitiseBody(in string) string {
	in = strings.Replace(in, "​", "", -1) // Strip Zero length space
	in = html.EscapeString(in)
	return strings.TrimSpace(in)
}

func BuildSlug(slug string, id int) string {
	if slug == "" || !Config.BuildSlugs {
		return strconv.Itoa(id)
	}
	return slug + "." + strconv.Itoa(id)
}
