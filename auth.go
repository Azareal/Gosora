/* Work in progress */
package main

import "log"
import "errors"
import "strconv"
import "net/http"
import "database/sql"

import "./query_gen/lib"
import "golang.org/x/crypto/bcrypt"

var auth Auth
var ErrMismatchedHashAndPassword = bcrypt.ErrMismatchedHashAndPassword
var ErrPasswordTooLong = errors.New("The password you selected is too long") // Silly, but we don't want bcrypt to bork on us

type Auth interface
{
	Authenticate(username string, password string) (uid int, err error)
	Logout(w http.ResponseWriter, uid int)
	ForceLogout(uid int) error
	SetCookies(w http.ResponseWriter, uid int, session string)
	GetCookies(r *http.Request) (uid int, session string, err error)
	SessionCheck(w http.ResponseWriter, r *http.Request) (user *User, halt bool)
	CreateSession(uid int) (session string, err error)
}

type DefaultAuth struct
{
	login *sql.Stmt
	logout *sql.Stmt
}

func NewDefaultAuth() *DefaultAuth {
	login_stmt, err := qgen.Builder.SimpleSelect("users","uid, password, salt","name = ?","","")
	if err != nil {
		log.Fatal(err)
	}
	logout_stmt, err := qgen.Builder.SimpleUpdate("users","session = ''","uid = ?")
	if err != nil {
		log.Fatal(err)
	}
	return &DefaultAuth{
		login: login_stmt,
		logout: logout_stmt,
	}
}

func (auth *DefaultAuth) Authenticate(username string, password string) (uid int, err error) {
	var real_password, salt string
	err = auth.login.QueryRow(username).Scan(&uid, &real_password, &salt)
	if err == ErrNoRows {
		return 0, errors.New("We couldn't find an account with that username.")
	} else if err != nil {
		LogError(err)
		return 0, errors.New("There was a glitch in the system. Please contact the system administrator.")
	}

	if salt == "" {
		// Send an email to admin for this?
		LogError(errors.New("Missing salt for user #" + strconv.Itoa(uid) + ". Potential security breach."))
		return 0, errors.New("There was a glitch in the system. Please contact the system administrator.")
	}

	err = CheckPassword(real_password,password,salt)
	if err == ErrMismatchedHashAndPassword {
		return 0, errors.New("That's not the correct password.")
	} else if err != nil {
		LogError(err)
		return 0, errors.New("There was a glitch in the system. Please contact the system administrator.")
	}

	return uid, nil
}

func (auth *DefaultAuth) ForceLogout(uid int) error {
	_, err := auth.logout.Exec(uid)
	if err != nil {
		LogError(err)
		return errors.New("There was a glitch in the system. Please contact the system administrator.")
	}

	// Flush the user out of the cache and reload
	err = users.Load(uid)
	if err != nil {
		return errors.New("Your account no longer exists!")
	}

	return nil
}

func (auth *DefaultAuth) Logout(w http.ResponseWriter, _ int) {
	cookie := http.Cookie{Name:"uid",Value:"",Path:"/",MaxAge: year}
	http.SetCookie(w,&cookie)
	cookie = http.Cookie{Name:"session",Value:"",Path:"/",MaxAge: year}
	http.SetCookie(w,&cookie)
}

func (auth *DefaultAuth) SetCookies(w http.ResponseWriter, uid int, session string) {
	cookie := http.Cookie{Name: "uid",Value: strconv.Itoa(uid),Path: "/",MaxAge: year}
	http.SetCookie(w,&cookie)
	cookie = http.Cookie{Name: "session",Value: session,Path: "/",MaxAge: year}
	http.SetCookie(w,&cookie)
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
		return &guest_user, false
	}

	// Is this session valid..?
	user, err = users.CascadeGet(uid)
	if err == ErrNoRows {
		return &guest_user, false
	} else if err != nil {
		InternalError(err,w,r)
		return &guest_user, true
	}

	if user.Session == "" || session != user.Session {
		return &guest_user, false
	}

	return user, false
}

func(auth *DefaultAuth) CreateSession(uid int) (session string, err error) {
	session, err = GenerateSafeString(sessionLength)
	if err != nil {
		return "", err
	}

	_, err = update_session_stmt.Exec(session, uid)
	if err != nil {
		return "", err
	}

	// Reload the user data
	_ = users.Load(uid)
	return session, nil
}
