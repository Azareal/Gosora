package common

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var weakPassStrings []string
var weakPassLit = make(map[string]struct{})
var ErrWeakPasswordNone = errors.New("You didn't put in a password.")
var ErrWeakPasswordShort = errors.New("Your password needs to be at-least eight characters long")
var ErrWeakPasswordNameInPass = errors.New("You can't use your name in your password.")
var ErrWeakPasswordEmailInPass = errors.New("You can't use your email in your password.")
var ErrWeakPasswordCommon = errors.New("You may not use a password that is in common use")
var ErrWeakPasswordNoNumbers = errors.New("You don't have any numbers in your password")
var ErrWeakPasswordNoUpper = errors.New("You don't have any uppercase characters in your password")
var ErrWeakPasswordNoLower = errors.New("You don't have any lowercase characters in your password")
var ErrWeakPasswordUniqueChars = errors.New("You don't have enough unique characters in your password")
var ErrWeakPasswordContains error

type weakpassHolder struct {
	Contains []string `json:"contains"`
	Literal  []string `json:"literal"`
}

func InitWeakPasswords() error {
	var weakpass weakpassHolder
	err := unmarshalJsonFile("./config/weakpass_default.json", &weakpass)
	if err != nil {
		return err
	}

	wcon := make(map[string]struct{})
	for _, item := range weakpass.Contains {
		wcon[item] = struct{}{}
	}
	for _, item := range weakpass.Literal {
		weakPassLit[item] = struct{}{}
	}

	weakpass = weakpassHolder{}
	err = unmarshalJsonFileIgnore404("./config/weakpass.json", &weakpass)
	if err != nil {
		return err
	}

	for _, item := range weakpass.Contains {
		wcon[item] = struct{}{}
	}
	for _, item := range weakpass.Literal {
		weakPassLit[item] = struct{}{}
	}
	weakPassStrings = make([]string, len(wcon))
	var i int
	for pattern, _ := range wcon {
		weakPassStrings[i] = pattern
		i++
	}

	s := "You may not have "
	for i, passBit := range weakPassStrings {
		if i > 0 {
			if i == len(weakPassStrings)-1 {
				s += " or "
			} else {
				s += ", "
			}
		}
		s += "'" + passBit + "'"
	}
	ErrWeakPasswordContains = errors.New(s + " in your password")

	return nil
}

func WeakPassword(password, username, email string) error {
	lowPassword := strings.ToLower(password)
	switch {
	case password == "":
		return ErrWeakPasswordNone
	case len(password) < 8:
		return ErrWeakPasswordShort
	case len(username) > 3 && strings.Contains(lowPassword, strings.ToLower(username)):
		return ErrWeakPasswordNameInPass
	case len(email) > 2 && strings.Contains(lowPassword, strings.ToLower(email)):
		return ErrWeakPasswordEmailInPass
	}
	if len(lowPassword) > 30 {
		return nil
	}

	litPass := lowPassword
	for i := 0; i < 10; i++ {
		litPass = strings.TrimSuffix(litPass, strconv.Itoa(i))
	}
	_, ok := weakPassLit[litPass]
	if ok {
		return ErrWeakPasswordCommon
	}
	for _, passBit := range weakPassStrings {
		if strings.Contains(lowPassword, passBit) {
			return ErrWeakPasswordContains
		}
	}

	charMap := make(map[rune]int)
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

	if upper == 0 {
		return ErrWeakPasswordNoUpper
	}
	if lower == 0 {
		return ErrWeakPasswordNoLower
	}
	if len(password) < 18 {
		if numbers == 0 {
			return ErrWeakPasswordNoNumbers
		}
		if (len(password) / 2) > len(charMap) {
			return ErrWeakPasswordUniqueChars
		}
	} else if (len(password) / 3) > len(charMap) {
		// Be a little lenient on the number of unique characters for long passwords
		return ErrWeakPasswordUniqueChars
	}
	return nil
}
