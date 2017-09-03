package main

import "encoding/base64"
import "crypto/rand"
import "golang.org/x/crypto/bcrypt"

// Generate a cryptographically secure set of random bytes..
func GenerateSafeString(length int) (string, error) {
	rb := make([]byte, length)
	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(rb), nil
}

// Generate a bcrypt hash from a password and a salt
func BcryptGeneratePassword(password string) (hashedPassword string, salt string, err error) {
	salt, err = GenerateSafeString(saltLength)
	if err != nil {
		return "", "", err
	}

	password = password + salt
	hashedPassword, err = bcryptGeneratePasswordNoSalt(password)
	if err != nil {
		return "", "", err
	}
	return hashedPassword, salt, nil
}

func bcryptGeneratePasswordNoSalt(password string) (hash string, err error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
