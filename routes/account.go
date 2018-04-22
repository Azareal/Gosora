package routes

import (
	"html"
	"log"
	"net/http"
	"strconv"
	"strings"

	"../common"
	"../query_gen/lib"
)

// A blank list to fill out that parameter in Page for routes which don't use it
var tList []interface{}

func AccountLogin(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if user.Loggedin {
		return common.LocalError("You're already logged in.", w, r, user)
	}
	pi := common.Page{common.GetTitlePhrase("login"), user, header, tList, nil}
	if common.RunPreRenderHook("pre_render_login", w, r, &user, &pi) {
		return nil
	}
	err := common.RunThemeTemplate(header.Theme.Name, "login", pi, w)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

// TODO: Log failed attempted logins?
// TODO: Lock IPS out if they have too many failed attempts?
// TODO: Log unusual countries in comparison to the country a user usually logs in from? Alert the user about this?
func AccountLoginSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	if user.Loggedin {
		return common.LocalError("You're already logged in.", w, r, user)
	}

	username := html.EscapeString(strings.Replace(r.PostFormValue("username"), "\n", "", -1))
	uid, err := common.Auth.Authenticate(username, r.PostFormValue("password"))
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	userPtr, err := common.Users.Get(uid)
	if err != nil {
		return common.LocalError("Bad account", w, r, user)
	}
	user = *userPtr

	var session string
	if user.Session == "" {
		session, err = common.Auth.CreateSession(uid)
		if err != nil {
			return common.InternalError(err, w, r)
		}
	} else {
		session = user.Session
	}

	common.Auth.SetCookies(w, uid, session)
	if user.IsAdmin {
		// Is this error check redundant? We already check for the error in PreRoute for the same IP
		// TODO: Should we be logging this?
		log.Printf("#%d has logged in with IP %s", uid, user.LastIP)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func AccountRegister(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if user.Loggedin {
		return common.LocalError("You're already logged in.", w, r, user)
	}
	pi := common.Page{common.GetTitlePhrase("register"), user, header, tList, nil}
	if common.RunPreRenderHook("pre_render_register", w, r, &user, &pi) {
		return nil
	}
	err := common.RunThemeTemplate(header.Theme.Name, "register", pi, w)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func AccountRegisterSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerLite, _ := common.SimpleUserCheck(w, r, &user)

	username := html.EscapeString(strings.Replace(r.PostFormValue("username"), "\n", "", -1))
	if username == "" {
		return common.LocalError("You didn't put in a username.", w, r, user)
	}
	email := html.EscapeString(strings.Replace(r.PostFormValue("email"), "\n", "", -1))
	if email == "" {
		return common.LocalError("You didn't put in an email.", w, r, user)
	}

	password := r.PostFormValue("password")
	// ?  Move this into Create()? What if we want to programatically set weak passwords for tests?
	err := common.WeakPassword(password, username, email)
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	confirmPassword := r.PostFormValue("confirm_password")
	common.DebugLog("Registration Attempt! Username: " + username) // TODO: Add more controls over what is logged when?

	// Do the two inputted passwords match..?
	if password != confirmPassword {
		return common.LocalError("The two passwords don't match.", w, r, user)
	}

	var active bool
	var group int
	switch headerLite.Settings["activation_type"] {
	case 1: // Activate All
		active = true
		group = common.Config.DefaultGroup
	default: // Anything else. E.g. Admin Activation or Email Activation.
		group = common.Config.ActivationGroup
	}

	uid, err := common.Users.Create(username, password, email, group, active)
	if err == common.ErrAccountExists {
		return common.LocalError("This username isn't available. Try another.", w, r, user)
	} else if err == common.ErrLongUsername {
		return common.LocalError("The username is too long, max: "+strconv.Itoa(common.Config.MaxUsernameLength), w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	// Check if this user actually owns this email, if email activation is on, automatically flip their account to active when the email is validated. Validation is also useful for determining whether this user should receive any alerts, etc. via email
	if common.Site.EnableEmails {
		token, err := common.GenerateSafeString(80)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		// TODO: Add an EmailStore and move this there
		acc := qgen.Builder.Accumulator()
		_, err = acc.Insert("emails").Columns("email, uid, validated, token").Fields("?,?,?,?").Exec(email, uid, 0, token)
		//_, err = stmts.addEmail.Exec(email, uid, 0, token)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		if !common.SendValidationEmail(username, email, token) {
			return common.LocalError("We were unable to send the email for you to confirm that this email address belongs to you. You may not have access to some functionality until you do so. Please ask an administrator for assistance.", w, r, user)
		}
	}

	session, err := common.Auth.CreateSession(uid)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	common.Auth.SetCookies(w, uid, session)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func AccountEditCritical(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	pi := common.Page{"Edit Password", user, header, tList, nil}
	if common.RunPreRenderHook("pre_render_account_own_edit_critical", w, r, &user, &pi) {
		return nil
	}
	err := common.Templates.ExecuteTemplate(w, "account_own_edit.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}
