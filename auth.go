/*
*
* Gosora Authentication Interface
* Copyright Azareal 2017 - 2018
*
 */
package main

import "log"
import "errors"
import "strconv"
import "net/http"
import "database/sql"

import "./query_gen/lib"
import "golang.org/x/crypto/bcrypt"

var auth Auth

// ErrMismatchedHashAndPassword is thrown whenever a hash doesn't match it's unhashed password
var ErrMismatchedHashAndPassword = bcrypt.ErrMismatchedHashAndPassword

// ErrPasswordTooLong is silly, but we don't want bcrypt to bork on us
var ErrPasswordTooLong = errors.New("The password you selected is too long")

// Auth is the main authentication interface.
type Auth interface {
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
	login  *sql.Stmt
	logout *sql.Stmt
}

// NewDefaultAuth is a factory for spitting out DefaultAuths
func NewDefaultAuth() *DefaultAuth {
	loginStmt, err := qgen.Builder.SimpleSelect("users", "uid, password, salt", "name = ?", "", "")
	if err != nil {
		log.Fatal(err)
	}
	logoutStmt, err := qgen.Builder.SimpleUpdate("users", "session = ''", "uid = ?")
	if err != nil {
		log.Fatal(err)
	}
	return &DefaultAuth{
		login:  loginStmt,
		logout: logoutStmt,
	}
}

// Authenticate checks if a specific username and password is valid and returns the UID for the corresponding user, if so. Otherwise, a user safe error.
func (auth *DefaultAuth) Authenticate(username string, password string) (uid int, err error) {
	var realPassword, salt string
	err = auth.login.QueryRow(username).Scan(&uid, &realPassword, &salt)
	if err == ErrNoRows {
		return 0, errors.New("We couldn't find an account with that username.") // nolint
	} else if err != nil {
		LogError(err)
		return 0, errors.New("There was a glitch in the system. Please contact your local administrator.") // nolint
	}

	if salt == "" {
		// Send an email to admin for this?
		LogError(errors.New("Missing salt for user #" + strconv.Itoa(uid) + ". Potential security breach."))
		return 0, errors.New("There was a glitch in the system. Please contact your local administrator")
	}

	err = CheckPassword(realPassword, password, salt)
	if err == ErrMismatchedHashAndPassword {
		return 0, errors.New("That's not the correct password.")
	} else if err != nil {
		LogError(err)
		return 0, errors.New("There was a glitch in the system. Please contact your local administrator.")
	}

	return uid, nil
}

// ForceLogout logs the user out of every computer, not just the one they logged out of
func (auth *DefaultAuth) ForceLogout(uid int) error {
	_, err := auth.logout.Exec(uid)
	if err != nil {
		LogError(err)
		return errors.New("There was a glitch in the system. Please contact your local administrator.")
	}

	// Flush the user out of the cache and reload
	err = users.Reload(uid)
	if err != nil {
		return errors.New("Your account no longer exists.")
	}

	return nil
}

// Logout logs you out of the computer you requested the logout for, but not the other computers you're logged in with
func (auth *DefaultAuth) Logout(w http.ResponseWriter, _ int) {
	cookie := http.Cookie{Name: "uid", Value: "", Path: "/", MaxAge: year}
	http.SetCookie(w, &cookie)
	cookie = http.Cookie{Name: "session", Value: "", Path: "/", MaxAge: year}
	http.SetCookie(w, &cookie)
}

// TODO: Set the cookie domain
func (auth *DefaultAuth) SetCookies(w http.ResponseWriter, uid int, session string) {
	cookie := http.Cookie{Name: "uid", Value: strconv.Itoa(uid), Path: "/", MaxAge: year}
	http.SetCookie(w, &cookie)
	cookie = http.Cookie{Name: "session", Value: session, Path: "/", MaxAge: year}
	http.SetCookie(w, &cookie)
}

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

func (auth *DefaultAuth) SessionCheck(w http.ResponseWriter, r *http.Request) (user *User, halt bool) {
	uid, session, err := auth.GetCookies(r)
	if err != nil {
		return &guestUser, false
	}

	// Is this session valid..?
	user, err = users.Get(uid)
	if err == ErrNoRows {
		return &guestUser, false
	} else if err != nil {
		InternalError(err, w)
		return &guestUser, true
	}

	if user.Session == "" || session != user.Session {
		return &guestUser, false
	}

	return user, false
}

func (auth *DefaultAuth) CreateSession(uid int) (session string, err error) {
	session, err = GenerateSafeString(sessionLength)
	if err != nil {
		return "", err
	}

	_, err = updateSessionStmt.Exec(session, uid)
	if err != nil {
		return "", err
	}

	// Reload the user data
	_ = users.Reload(uid)
	return session, nil
}
