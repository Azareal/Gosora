/*
*
* Gosora Authentication Interface
* Copyright Azareal 2017 - 2020
*
 */
package common

import (
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Azareal/Gosora/common/gauth"
	qgen "github.com/Azareal/Gosora/query_gen"

	//"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// TODO: Write more authentication tests
var Auth AuthInt

const SaltLength int = 32
const SessionLength int = 80

// ErrMismatchedHashAndPassword is thrown whenever a hash doesn't match it's unhashed password
var ErrMismatchedHashAndPassword = bcrypt.ErrMismatchedHashAndPassword

// nolint
var ErrHashNotExist = errors.New("We don't recognise that hashing algorithm")
var ErrTooFewHashParams = errors.New("You haven't provided enough hash parameters")

// ErrPasswordTooLong is silly, but we don't want bcrypt to bork on us
var ErrPasswordTooLong = errors.New("The password you selected is too long")
var ErrWrongPassword = errors.New("That's not the correct password.")
var ErrBadMFAToken = errors.New("I'm not sure where you got that from, but that's not a valid 2FA token")
var ErrWrongMFAToken = errors.New("That 2FA token isn't correct")
var ErrNoMFAToken = errors.New("This user doesn't have 2FA setup")
var ErrSecretError = errors.New("There was a glitch in the system. Please contact your local administrator.")
var ErrNoUserByName = errors.New("We couldn't find an account with that username.")
var DefaultHashAlgo = "bcrypt" // Override this in the configuration file, not here

//func(realPassword string, password string, salt string) (err error)
var CheckPasswordFuncs = map[string]func(string, string, string) error{
	"bcrypt": BcryptCheckPassword,
	//"argon2": Argon2CheckPassword,
}

//func(password string) (hashedPassword string, salt string, err error)
var GeneratePasswordFuncs = map[string]func(string) (string, string, error){
	"bcrypt": BcryptGeneratePassword,
	//"argon2": Argon2GeneratePassword,
}

// TODO: Redirect 2b to bcrypt too?
var HashPrefixes = map[string]string{
	"$2a$": "bcrypt",
	//"argon2$": "argon2",
}

// AuthInt is the main authentication interface.
type AuthInt interface {
	Authenticate(name, password string) (uid int, err error, requiresExtraAuth bool)
	ValidateMFAToken(mfaToken string, uid int) error
	Logout(w http.ResponseWriter, uid int)
	ForceLogout(uid int) error
	SetCookies(w http.ResponseWriter, uid int, session string)
	SetProvisionalCookies(w http.ResponseWriter, uid int, session, signedSession string) // To avoid logging someone in until they've passed the MFA check
	GetCookies(r *http.Request) (uid int, session string, err error)
	SessionCheck(w http.ResponseWriter, r *http.Request) (u *User, halt bool)
	CreateSession(uid int) (session string, err error)
	CreateProvisionalSession(uid int) (provSession, signedSession string, err error) // To avoid logging someone in until they've passed the MFA check
}

// DefaultAuth is the default authenticator used by Gosora, may be swapped with an alternate authenticator in some situations. E.g. To support LDAP.
type DefaultAuth struct {
	login         *sql.Stmt
	logout        *sql.Stmt
	updateSession *sql.Stmt
}

// NewDefaultAuth is a factory for spitting out DefaultAuths
func NewDefaultAuth() (*DefaultAuth, error) {
	acc := qgen.NewAcc()
	return &DefaultAuth{
		login:         acc.Select("users").Columns("uid, password, salt").Where("name = ?").Prepare(),
		logout:        acc.Update("users").Set("session = ''").Where("uid = ?").Prepare(),
		updateSession: acc.Update("users").Set("session = ?").Where("uid = ?").Prepare(),
	}, acc.FirstError()
}

// Authenticate checks if a specific username and password is valid and returns the UID for the corresponding user, if so. Otherwise, a user safe error.
// IF MFA is enabled, then pass it back a flag telling the caller that authentication isn't complete yet
// TODO: Find a better way of handling errors we don't want to reach the user
func (auth *DefaultAuth) Authenticate(name string, password string) (uid int, err error, requiresExtraAuth bool) {
	var realPassword, salt string
	err = auth.login.QueryRow(name).Scan(&uid, &realPassword, &salt)
	if err == ErrNoRows {
		return 0, ErrNoUserByName, false
	} else if err != nil {
		LogError(err)
		return 0, ErrSecretError, false
	}

	err = CheckPassword(realPassword, password, salt)
	if err == ErrMismatchedHashAndPassword {
		return 0, ErrWrongPassword, false
	} else if err != nil {
		LogError(err)
		return 0, ErrSecretError, false
	}

	_, err = MFAstore.Get(uid)
	if err != sql.ErrNoRows && err != nil {
		LogError(err)
		return 0, ErrSecretError, false
	}
	if err != ErrNoRows {
		return uid, nil, true
	}

	return uid, nil, false
}

func (auth *DefaultAuth) ValidateMFAToken(mfaToken string, uid int) error {
	mfaItem, err := MFAstore.Get(uid)
	if err != sql.ErrNoRows && err != nil {
		LogError(err)
		return ErrSecretError
	}
	if err == ErrNoRows {
		return ErrNoMFAToken
	}

	ok, err := VerifyGAuthToken(mfaItem.Secret, mfaToken)
	if err != nil {
		return ErrBadMFAToken
	}
	if ok {
		return nil
	}

	for i, scratch := range mfaItem.Scratch {
		if subtle.ConstantTimeCompare([]byte(scratch), []byte(mfaToken)) == 1 {
			err = mfaItem.BurnScratch(i)
			if err != nil {
				LogError(err)
				return ErrSecretError
			}
			return nil
		}
	}

	return ErrWrongMFAToken
}

// ForceLogout logs the user out of every computer, not just the one they logged out of
func (auth *DefaultAuth) ForceLogout(uid int) error {
	_, err := auth.logout.Exec(uid)
	if err != nil {
		LogError(err)
		return ErrSecretError
	}

	// Flush the user out of the cache
	if uc := Users.GetCache(); uc != nil {
		uc.Remove(uid)
	}
	return nil
}

func setCookie(w http.ResponseWriter, cookie *http.Cookie, sameSite string) {
	if v := cookie.String(); v != "" {
		switch sameSite {
		case "lax":
			v = v + "; SameSite=lax"
		case "strict":
			v = v + "; SameSite"
		}
		w.Header().Add("Set-Cookie", v)
	}
}

func deleteCookie(w http.ResponseWriter, cookie *http.Cookie) {
	cookie.MaxAge = -1
	http.SetCookie(w, cookie)
}

// Logout logs you out of the computer you requested the logout for, but not the other computers you're logged in with
func (auth *DefaultAuth) Logout(w http.ResponseWriter, _ int) {
	cookie := http.Cookie{Name: "uid", Value: "", Path: "/"}
	deleteCookie(w, &cookie)
	cookie = http.Cookie{Name: "session", Value: "", Path: "/"}
	deleteCookie(w, &cookie)
}

// TODO: Set the cookie domain
// SetCookies sets the two cookies required for the current user to be recognised as a specific user in future requests
func (auth *DefaultAuth) SetCookies(w http.ResponseWriter, uid int, session string) {
	cookie := http.Cookie{Name: "uid", Value: strconv.Itoa(uid), Path: "/", MaxAge: int(Year)}
	setCookie(w, &cookie, "lax")
	cookie = http.Cookie{Name: "session", Value: session, Path: "/", MaxAge: int(Year)}
	setCookie(w, &cookie, "lax")
}

// TODO: Set the cookie domain
// SetProvisionalCookies sets the two cookies required for guests to be recognised as having passed the initial login but not having passed the additional checks (e.g. multi-factor authentication)
func (auth *DefaultAuth) SetProvisionalCookies(w http.ResponseWriter, uid int, provSession string, signedSession string) {
	cookie := http.Cookie{Name: "uid", Value: strconv.Itoa(uid), Path: "/", MaxAge: int(Year)}
	setCookie(w, &cookie, "lax")
	cookie = http.Cookie{Name: "provSession", Value: provSession, Path: "/", MaxAge: int(Year)}
	setCookie(w, &cookie, "lax")
	cookie = http.Cookie{Name: "signedSession", Value: signedSession, Path: "/", MaxAge: int(Year)}
	setCookie(w, &cookie, "lax")
}

// GetCookies fetches the current user's session cookies
func (auth *DefaultAuth) GetCookies(r *http.Request) (uid int, session string, err error) {
	// Are there any session cookies..?
	cookie, err := r.Cookie("uid")
	if err != nil {
		return 0, "", err
	}
	uid, err = strconv.Atoi(cookie.Value)
	if err != nil {
		return 0, "", err
	}
	cookie, err = r.Cookie("session")
	if err != nil {
		return 0, "", err
	}
	return uid, cookie.Value, err
}

// SessionCheck checks if a user has session cookies and whether they're valid
func (auth *DefaultAuth) SessionCheck(w http.ResponseWriter, r *http.Request) (user *User, halt bool) {
	uid, session, err := auth.GetCookies(r)
	if err != nil {
		return &GuestUser, false
	}

	// Is this session valid..?
	user, err = Users.Get(uid)
	if err == ErrNoRows {
		return &GuestUser, false
	} else if err != nil {
		InternalError(err, w, r)
		return &GuestUser, true
	}

	// We need to do a constant time compare, otherwise someone might be able to deduce the session character by character based on how long it takes to do the comparison. Change this at your own peril.
	if user.Session == "" || subtle.ConstantTimeCompare([]byte(session), []byte(user.Session)) != 1 {
		return &GuestUser, false
	}

	return user, false
}

// CreateSession generates a new session to allow a remote client to stay logged in as a specific user
func (auth *DefaultAuth) CreateSession(uid int) (session string, err error) {
	session, err = GenerateSafeString(SessionLength)
	if err != nil {
		return "", err
	}

	_, err = auth.updateSession.Exec(session, uid)
	if err != nil {
		return "", err
	}

	// Flush the user data from the cache
	ucache := Users.GetCache()
	if ucache != nil {
		ucache.Remove(uid)
	}
	return session, nil
}

func (auth *DefaultAuth) CreateProvisionalSession(uid int) (provSession string, signedSession string, err error) {
	provSession, err = GenerateSafeString(SessionLength)
	if err != nil {
		return "", "", err
	}

	h := sha256.New()
	h.Write([]byte(SessionSigningKeyBox.Load().(string)))
	h.Write([]byte(provSession))
	h.Write([]byte(strconv.Itoa(uid)))
	return provSession, hex.EncodeToString(h.Sum(nil)), nil
}

func CheckPassword(realPassword, password, salt string) (err error) {
	blasted := strings.Split(realPassword, "$")
	prefix := blasted[0]
	if len(blasted) > 1 {
		prefix += "$" + blasted[1] + "$"
	}
	algo, ok := HashPrefixes[prefix]
	if !ok {
		return ErrHashNotExist
	}
	checker := CheckPasswordFuncs[algo]
	return checker(realPassword, password, salt)
}

func GeneratePassword(password string) (hash, salt string, err error) {
	gen, ok := GeneratePasswordFuncs[DefaultHashAlgo]
	if !ok {
		return "", "", ErrHashNotExist
	}
	return gen(password)
}

func BcryptCheckPassword(realPassword, password, salt string) (err error) {
	return bcrypt.CompareHashAndPassword([]byte(realPassword), []byte(password+salt))
}

// Note: The salt is in the hash, therefore the salt parameter is blank
func BcryptGeneratePassword(password string) (hash, salt string, err error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	return string(hashedPassword), salt, nil
}

/*const (
	argon2Time    uint32 = 3
	argon2Memory  uint32 = 32 * 1024
	argon2Threads uint8  = 4
	argon2KeyLen  uint32 = 32
)

func Argon2CheckPassword(realPassword, password, salt string) (err error) {
	split := strings.Split(realPassword, "$")
	// TODO: Better validation
	if len(split) < 5 {
		return ErrTooFewHashParams
	}
	realKey, _ := base64.StdEncoding.DecodeString(split[len(split)-1])
	time, _ := strconv.Atoi(split[1])
	memory, _ := strconv.Atoi(split[2])
	threads, _ := strconv.Atoi(split[3])
	keyLen, _ := strconv.Atoi(split[4])
	key := argon2.Key([]byte(password), []byte(salt), uint32(time), uint32(memory), uint8(threads), uint32(keyLen))
	if subtle.ConstantTimeCompare(realKey, key) != 1 {
		return ErrMismatchedHashAndPassword
	}
	return nil
}

func Argon2GeneratePassword(password string) (hash, salt string, err error) {
	sbytes := make([]byte, SaltLength)
	_, err = rand.Read(sbytes)
	if err != nil {
		return "", "", err
	}
	key := argon2.Key([]byte(password), sbytes, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	hash = base64.StdEncoding.EncodeToString(key)
	return fmt.Sprintf("argon2$%d%d%d%d%s%s", argon2Time, argon2Memory, argon2Threads, argon2KeyLen, salt, hash), string(sbytes), nil
}
*/

// TODO: Test this with Google Authenticator proper
func FriendlyGAuthSecret(secret string) (out string) {
	for i, char := range secret {
		out += string(char)
		if (i+1)%4 == 0 {
			out += " "
		}
	}
	return strings.TrimSpace(out)
}
func GenerateGAuthSecret() (string, error) {
	return GenerateStd32SafeString(14)
}
func VerifyGAuthToken(secret, token string) (bool, error) {
	trueToken, err := gauth.GetTOTPToken(secret)
	return subtle.ConstantTimeCompare([]byte(trueToken), []byte(token)) == 1, err
}
