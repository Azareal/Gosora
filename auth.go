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

type Auth interface
{
	Authenticate(username string, password string) (int,error)
	Logout(w http.ResponseWriter, uid int)
	ForceLogout(uid int) error
	SetCookies(w http.ResponseWriter, uid int, session string)
	CreateSession(uid int) (string, error)
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

func(auth *DefaultAuth) CreateSession(uid int) (string, error) {
	session, err := GenerateSafeString(sessionLength)
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
