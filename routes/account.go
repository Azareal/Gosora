package routes

import (
	"database/sql"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
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

func AccountLogout(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	if !user.Loggedin {
		return common.LocalError("You can't logout without logging in first.", w, r, user)
	}
	common.Auth.Logout(w, user.ID)
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

func AccountEditCriticalSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	var realPassword, salt string
	currentPassword := r.PostFormValue("account-current-password")
	newPassword := r.PostFormValue("account-new-password")
	confirmPassword := r.PostFormValue("account-confirm-password")

	// TODO: Use a reusable statement
	acc := qgen.Builder.Accumulator()
	err := acc.Select("users").Columns("password, salt").Where("uid = ?").QueryRow(user.ID).Scan(&realPassword, &salt)
	if err == sql.ErrNoRows {
		return common.LocalError("Your account no longer exists.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	err = common.CheckPassword(realPassword, currentPassword, salt)
	if err == common.ErrMismatchedHashAndPassword {
		return common.LocalError("That's not the correct password.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}
	if newPassword != confirmPassword {
		return common.LocalError("The two passwords don't match.", w, r, user)
	}
	common.SetPassword(user.ID, newPassword)

	// Log the user out as a safety precaution
	common.Auth.ForceLogout(user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func AccountEditAvatar(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	pi := common.Page{"Edit Avatar", user, headerVars, tList, nil}
	if common.RunPreRenderHook("pre_render_account_own_edit_avatar", w, r, &user, &pi) {
		return nil
	}
	err := common.Templates.ExecuteTemplate(w, "account_own_edit_avatar.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func AccountEditAvatarSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	var filename, ext string
	for _, fheaders := range r.MultipartForm.File {
		for _, hdr := range fheaders {
			if hdr.Filename == "" {
				continue
			}
			infile, err := hdr.Open()
			if err != nil {
				return common.LocalError("Upload failed", w, r, user)
			}
			defer infile.Close()

			// We don't want multiple files
			// TODO: Check the length of r.MultipartForm.File and error rather than doing this x.x
			if filename != "" {
				if filename != hdr.Filename {
					os.Remove("./uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext)
					return common.LocalError("You may only upload one avatar", w, r, user)
				}
			} else {
				filename = hdr.Filename
			}

			if ext == "" {
				extarr := strings.Split(hdr.Filename, ".")
				if len(extarr) < 2 {
					return common.LocalError("Bad file", w, r, user)
				}
				ext = extarr[len(extarr)-1]

				// TODO: Can we do this without a regex?
				reg, err := regexp.Compile("[^A-Za-z0-9]+")
				if err != nil {
					return common.LocalError("Bad file extension", w, r, user)
				}
				ext = reg.ReplaceAllString(ext, "")
				ext = strings.ToLower(ext)
			}

			outfile, err := os.Create("./uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext)
			if err != nil {
				return common.LocalError("Upload failed [File Creation Failed]", w, r, user)
			}
			defer outfile.Close()

			_, err = io.Copy(outfile, infile)
			if err != nil {
				return common.LocalError("Upload failed [Copy Failed]", w, r, user)
			}
		}
	}
	if ext == "" {
		return common.LocalError("No file", w, r, user)
	}

	err := user.ChangeAvatar("." + ext)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	user.Avatar = "/uploads/avatar_" + strconv.Itoa(user.ID) + "." + ext
	headerVars.NoticeList = append(headerVars.NoticeList, common.GetNoticePhrase("account_avatar_updated"))

	pi := common.Page{"Edit Avatar", user, headerVars, tList, nil}
	if common.RunPreRenderHook("pre_render_account_own_edit_avatar", w, r, &user, &pi) {
		return nil
	}
	err = common.Templates.ExecuteTemplate(w, "account_own_edit_avatar.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func AccountEditUsername(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	headerVars, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if r.FormValue("updated") == "1" {
		headerVars.NoticeList = append(headerVars.NoticeList, common.GetNoticePhrase("account_username_updated"))
	}

	pi := common.Page{"Edit Username", user, headerVars, tList, user.Name}
	if common.RunPreRenderHook("pre_render_account_own_edit_username", w, r, &user, &pi) {
		return nil
	}
	err := common.Templates.ExecuteTemplate(w, "account_own_edit_username.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func AccountEditUsernameSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	newUsername := html.EscapeString(strings.Replace(r.PostFormValue("account-new-username"), "\n", "", -1))
	err := user.ChangeName(newUsername)
	if err != nil {
		return common.LocalError("Unable to change the username. Does someone else already have this name?", w, r, user)
	}

	http.Redirect(w, r, "/user/edit/username/?updated=1", http.StatusSeeOther)
	return nil
}
