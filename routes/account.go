package routes

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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
	header.Title = common.GetTitlePhrase("login")
	pi := common.Page{header, tList, nil}
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

	username := common.SanitiseSingleLine(r.PostFormValue("username"))
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
	header.Title = common.GetTitlePhrase("register")

	h := sha256.New()
	h.Write([]byte(common.JSTokenBox.Load().(string)))
	h.Write([]byte(user.LastIP))
	jsToken := hex.EncodeToString(h.Sum(nil))
	pi := common.Page{header, tList, jsToken}
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

	// TODO: Should we push multiple validation errors to the user instead of just one?
	var regSuccess = true
	var regErrMsg = ""
	var regErrReason = ""
	var regError = func(userMsg string, reason string) {
		regSuccess = false
		if regErrMsg == "" {
			regErrMsg = userMsg
		}
		regErrReason += reason + "|"
	}

	if r.PostFormValue("tos") != "0" {
		regError("You might be a machine", "trap-question")
	}
	h := sha256.New()
	h.Write([]byte(common.JSTokenBox.Load().(string)))
	h.Write([]byte(user.LastIP))
	if r.PostFormValue("antispam") != hex.EncodeToString(h.Sum(nil)) {
		regError("You might be a machine", "js-antispam")
	}

	username := common.SanitiseSingleLine(r.PostFormValue("username"))
	// TODO: Add a dedicated function for validating emails
	email := common.SanitiseSingleLine(r.PostFormValue("email"))
	if username == "" {
		regError("You didn't put in a username.", "no-username")
	}
	if email == "" {
		regError("You didn't put in an email.", "no-email")
	}

	password := r.PostFormValue("password")
	// ?  Move this into Create()? What if we want to programatically set weak passwords for tests?
	err := common.WeakPassword(password, username, email)
	if err != nil {
		regError(err.Error(), "weak-password")
	} else {
		// Do the two inputted passwords match..?
		confirmPassword := r.PostFormValue("confirm_password")
		if password != confirmPassword {
			regError("The two passwords don't match.", "password-mismatch")
		}
	}

	regLog := common.RegLogItem{Username: username, Email: email, FailureReason: regErrReason, Success: regSuccess, IPAddress: user.LastIP}
	_, err = regLog.Create()
	if err != nil {
		return common.InternalError(err, w, r)
	}
	if !regSuccess {
		return common.LocalError(regErrMsg, w, r, user)
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

	// TODO: Do the registration attempt logging a little less messily (without having to amend the result after the insert)
	uid, err := common.Users.Create(username, password, email, group, active)
	if err != nil {
		regLog.Success = false
		if err == common.ErrAccountExists {
			regLog.FailureReason += "username-exists"
			err = regLog.Commit()
			if err != nil {
				return common.InternalError(err, w, r)
			}
			return common.LocalError("This username isn't available. Try another.", w, r, user)
		} else if err == common.ErrLongUsername {
			regLog.FailureReason += "username-too-long"
			err = regLog.Commit()
			if err != nil {
				return common.InternalError(err, w, r)
			}
			return common.LocalError("The username is too long, max: "+strconv.Itoa(common.Config.MaxUsernameLength), w, r, user)
		}
		regLog.FailureReason += "internal-error"
		err2 := regLog.Commit()
		if err2 != nil {
			return common.InternalError(err2, w, r)
		}
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

// TODO: Rename this
func AccountEditCritical(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	// TODO: Add a phrase for this
	header.Title = "Edit Password"

	pi := common.Page{header, tList, nil}
	if common.RunPreRenderHook("pre_render_account_own_edit_password", w, r, &user, &pi) {
		return nil
	}
	err := common.Templates.ExecuteTemplate(w, "account_own_edit_password.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

// TODO: Rename this
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
	header, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	header.Title = common.GetTitlePhrase("account_avatar")
	if r.FormValue("updated") == "1" {
		header.AddNotice("account_avatar_updated")
	}

	pi := common.Page{header, tList, nil}
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
	_, ferr := common.SimpleUserCheck(w, r, &user)
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
	http.Redirect(w, r, "/user/edit/avatar/?updated=1", http.StatusSeeOther)
	return nil
}

func AccountEditUsername(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	header.Title = common.GetTitlePhrase("account_username")
	if r.FormValue("updated") == "1" {
		header.AddNotice("account_username_updated")
	}

	pi := common.Page{header, tList, user.Name}
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

	newUsername := common.SanitiseSingleLine(r.PostFormValue("account-new-username"))
	err := user.ChangeName(newUsername)
	if err != nil {
		return common.LocalError("Unable to change the username. Does someone else already have this name?", w, r, user)
	}

	http.Redirect(w, r, "/user/edit/username/?updated=1", http.StatusSeeOther)
	return nil
}

func AccountEditEmail(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	header.Title = common.GetTitlePhrase("account_email")

	emails, err := common.Emails.GetEmailsByUser(&user)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// Was this site migrated from another forum software? Most of them don't have multiple emails for a single user.
	// This also applies when the admin switches site.EnableEmails on after having it off for a while.
	if len(emails) == 0 {
		email := common.Email{UserID: user.ID}
		email.Email = user.Email
		email.Validated = false
		email.Primary = true
		emails = append(emails, email)
	}

	if !common.Site.EnableEmails {
		header.AddNotice("account_mail_disabled")
	}
	if r.FormValue("verified") == "1" {
		header.AddNotice("account_mail_verify_success")
	}

	pi := common.EmailListPage{header, emails, nil}
	if common.RunPreRenderHook("pre_render_account_own_edit_email", w, r, &user, &pi) {
		return nil
	}
	err = common.Templates.ExecuteTemplate(w, "account_own_edit_email.html", pi)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

// TODO: Should we make this an AnonAction so someone can do this without being logged in?
func AccountEditEmailTokenSubmit(w http.ResponseWriter, r *http.Request, user common.User, token string) common.RouteError {
	header, ferr := common.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !common.Site.EnableEmails {
		http.Redirect(w, r, "/user/edit/email/", http.StatusSeeOther)
		return nil
	}

	targetEmail := common.Email{UserID: user.ID}
	emails, err := common.Emails.GetEmailsByUser(&user)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	for _, email := range emails {
		if email.Token == token {
			targetEmail = email
		}
	}

	if len(emails) == 0 {
		return common.LocalError("A verification email was never sent for you!", w, r, user)
	}
	if targetEmail.Token == "" {
		return common.LocalError("That's not a valid token!", w, r, user)
	}

	err = common.Emails.VerifyEmail(user.Email)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// If Email Activation is on, then activate the account while we're here
	if header.Settings["activation_type"] == 2 {
		err = user.Activate()
		if err != nil {
			return common.InternalError(err, w, r)
		}
	}
	http.Redirect(w, r, "/user/edit/email/?verified=1", http.StatusSeeOther)

	return nil
}
