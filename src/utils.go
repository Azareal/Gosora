package main
import "encoding/base64"
import "crypto/rand"

// Generate a cryptographically secure set of random bytes..
func GenerateSafeString(length int) (string, error) {
	rb := make([]byte,length)
	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(rb), nil
}