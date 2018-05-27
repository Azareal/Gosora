package install

import "encoding/base64"
import "crypto/rand"
import "golang.org/x/crypto/bcrypt"

const saltLength int = 32

// Generate a cryptographically secure set of random bytes..
func GenerateSafeString(length int) (string, error) {
	rb := make([]byte, length)
	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(rb), nil
}

// Generate a bcrypt hash
// Note: The salt is in the hash, therefore the salt value is blank
func BcryptGeneratePassword(password string) (hash string, salt string, err error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	return string(hashedPassword), salt, nil
}
