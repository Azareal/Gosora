/*
*
* Gosora Authentication Interface
* Copyright Azareal 2017 - 2018
*
 */
package common

import "errors"
import "strconv"
import "net/http"
import "database/sql"

import "golang.org/x/crypto/bcrypt"
import "../query_gen/lib"

var Auth AuthInt

// ErrMismatchedHashAndPassword is thrown whenever a hash doesn't match it's unhashed password
var ErrMismatchedHashAndPassword = bcrypt.ErrMismatchedHashAndPassword

// nolint
// ErrPasswordTooLong is silly, but we don't want bcrypt to bork on us
var ErrPasswordTooLong = errors.New("The password you selected is too long")
var ErrWrongPassword = errors.New("That's not the correct password.")
var ErrSecretError = errors.New("There was a glitch in the system. Please contact your local administrator.")
var ErrNoUserByName = errors.New("We couldn't find an account with that username.")

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
func (auth *DefaultAuth) Authenticate(username string, password string) (uid int, err error) {
	var realPassword, salt string
	err = auth.login.QueryRow(username).Scan(&uid, &realPassword, &salt)
	if err == ErrNoRows {
		return 0, ErrNoUserByName
	} else if err != nil {
		LogError(err)
		return 0, ErrSecretError
	}

	if salt == "" {
		// Send an email to admin for this?
		LogError(errors.New("Missing salt for user #" + strconv.Itoa(uid) + ". Potential security breach."))
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
	cookie := http.Cookie{Name: "uid", Value: "", Path: "/", MaxAge: Year}
	http.SetCookie(w, &cookie)
	cookie = http.Cookie{Name: "session", Value: "", Path: "/", MaxAge: Year}
	http.SetCookie(w, &cookie)
}

// TODO: Set the cookie domain
// SetCookies sets the two cookies required for the current user to be recognised as a specific user in future requests
func (auth *DefaultAuth) SetCookies(w http.ResponseWriter, uid int, session string) {
	cookie := http.Cookie{Name: "uid", Value: strconv.Itoa(uid), Path: "/", MaxAge: Year}
	http.SetCookie(w, &cookie)
	cookie = http.Cookie{Name: "session", Value: session, Path: "/", MaxAge: Year}
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
