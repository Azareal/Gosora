/*
*
* Gosora Authentication Interface
* Copyright Azareal 2017 - 2019
*
 */
package common

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"../query_gen/lib"
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

var HashPrefixes = map[string]string{
	"$2a$": "bcrypt",
	//"argon2$": "argon2",
}

// AuthInt is the main authentication interface.
type AuthInt interface {
	Authenticate(username string, password string) (uid int, err error)
	Logout(w http.ResponseWriter, uid int)
	ForceLogout(uid int) error
	SetCookies(w http.ResponseWriter, uid int, session string)
	GetCookies(r *http.Request) (uid int, session string, err error)
	SessionCheck(w http.ResponseWriter, r *http.Request) (user *User, halt bool)
	CreateSession(uid int) (session string, err error)
}

// DefaultAuth is the default authenticator used by Gosora, may be swapped with an alternate authenticator in some situations. E.g. To support LDAP.
type DefaultAuth struct {
	login         *sql.Stmt
	logout        *sql.Stmt
	updateSession *sql.Stmt
}

// NewDefaultAuth is a factory for spitting out DefaultAuths
func NewDefaultAuth() (*DefaultAuth, error) {
	acc := qgen.Builder.Accumulator()
	return &DefaultAuth{
		login:         acc.Select("users").Columns("uid, password, salt").Where("name = ?").Prepare(),
		logout:        acc.Update("users").Set("session = ''").Where("uid = ?").Prepare(),
		updateSession: acc.Update("users").Set("session = ?").Where("uid = ?").Prepare(),
	}, acc.FirstError()
}

// Authenticate checks if a specific username and password is valid and returns the UID for the corresponding user, if so. Otherwise, a user safe error.
// TODO: Find a better way of handling errors we don't want to reach the user
func (auth *DefaultAuth) Authenticate(username string, password string) (uid int, err error) {
	var realPassword, salt string
	err = auth.login.QueryRow(username).Scan(&uid, &realPassword, &salt)
	if err == ErrNoRows {
		return 0, ErrNoUserByName
	} else if err != nil {
		LogError(err)
		return 0, ErrSecretError
	}

	err = CheckPassword(realPassword, password, salt)
	if err == ErrMismatchedHashAndPassword {
		return 0, ErrWrongPassword
	} else if err != nil {
		LogError(err)
		return 0, ErrSecretError
	}

	return uid, nil
}

// ForceLogout logs the user out of every computer, not just the one they logged out of
func (auth *DefaultAuth) ForceLogout(uid int) error {
	_, err := auth.logout.Exec(uid)
	if err != nil {
		LogError(err)
		return ErrSecretError
	}

	// Flush the user out of the cache
	ucache := Users.GetCache()
	if ucache != nil {
		ucache.Remove(uid)
	}

	return nil
}

// Logout logs you out of the computer you requested the logout for, but not the other computers you're logged in with
func (auth *DefaultAuth) Logout(w http.ResponseWriter, _ int) {
	cookie := http.Cookie{Name: "uid", Value: "", Path: "/", MaxAge: int(Year)}
	http.SetCookie(w, &cookie)
	cookie = http.Cookie{Name: "session", Value: "", Path: "/", MaxAge: int(Year)}
	http.SetCookie(w, &cookie)
}

// TODO: Set the cookie domain
// SetCookies sets the two cookies required for the current user to be recognised as a specific user in future requests
func (auth *DefaultAuth) SetCookies(w http.ResponseWriter, uid int, session string) {
	cookie := http.Cookie{Name: "uid", Value: strconv.Itoa(uid), Path: "/", MaxAge: int(Year)}
	http.SetCookie(w, &cookie)
	cookie = http.Cookie{Name: "session", Value: session, Path: "/", MaxAge: int(Year)}
	http.SetCookie(w, &cookie)
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

	if user.Session == "" || session != user.Session {
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

func CheckPassword(realPassword string, password string, salt string) (err error) {
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

func GeneratePassword(password string) (hash string, salt string, err error) {
	gen, ok := GeneratePasswordFuncs[DefaultHashAlgo]
	if !ok {
		return "", "", ErrHashNotExist
	}
	return gen(password)
}

func BcryptCheckPassword(realPassword string, password string, salt string) (err error) {
	return bcrypt.CompareHashAndPassword([]byte(realPassword), []byte(password+salt))
}

// Note: The salt is in the hash, therefore the salt parameter is blank
func BcryptGeneratePassword(password string) (hash string, salt string, err error) {
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

func Argon2CheckPassword(realPassword string, password string, salt string) (err error) {
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

func Argon2GeneratePassword(password string) (hash string, salt string, err error) {
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
