package routes

import (
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"html"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
	p "github.com/Azareal/Gosora/common/phrases"
	qgen "github.com/Azareal/Gosora/query_gen"
)

// A blank list to fill out that parameter in Page for routes which don't use it
var tList []interface{}

func AccountLogin(w http.ResponseWriter, r *http.Request, user c.User, h *c.Header) c.RouteError {
	if user.Loggedin {
		return c.LocalError("You're already logged in.", w, r, user)
	}
	h.Title = p.GetTitlePhrase("login")
	return renderTemplate("login", w, r, h, c.Page{h, tList, nil})
}

// TODO: Log failed attempted logins?
// TODO: Lock IPS out if they have too many failed attempts?
// TODO: Log unusual countries in comparison to the country a user usually logs in from? Alert the user about this?
func AccountLoginSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	if user.Loggedin {
		return c.LocalError("You're already logged in.", w, r, user)
	}

	name := c.SanitiseSingleLine(r.PostFormValue("username"))
	uid, err, requiresExtraAuth := c.Auth.Authenticate(name, r.PostFormValue("password"))
	if err != nil {
		// TODO: uid is currently set to 0 as authenticate fetches the user by username and password. Get the actual uid, so we can alert the user of attempted logins? What if someone takes advantage of the response times to deduce if an account exists?
		logItem := &c.LoginLogItem{UID: uid, Success: false, IP: user.GetIP()}
		_, err := logItem.Create()
		if err != nil {
			return c.InternalError(err, w, r)
		}
		return c.LocalError(err.Error(), w, r, user)
	}

	// TODO: Take 2FA into account
	logItem := &c.LoginLogItem{UID: uid, Success: true, IP: user.GetIP()}
	_, err = logItem.Create()
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// TODO: Do we want to slacken this by only doing it when the IP changes?
	if requiresExtraAuth {
		provSession, signedSession, err := c.Auth.CreateProvisionalSession(uid)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		// TODO: Use the login log ID in the provisional cookie?
		c.Auth.SetProvisionalCookies(w, uid, provSession, signedSession)
		http.Redirect(w, r, "/accounts/mfa_verify/", http.StatusSeeOther)
		return nil
	}

	return loginSuccess(uid, w, r, &user)
}

func loginSuccess(uid int, w http.ResponseWriter, r *http.Request, user *c.User) c.RouteError {
	userPtr, err := c.Users.Get(uid)
	if err != nil {
		return c.LocalError("Bad account", w, r, *user)
	}
	*user = *userPtr

	var session string
	if user.Session == "" {
		session, err = c.Auth.CreateSession(uid)
		if err != nil {
			return c.InternalError(err, w, r)
		}
	} else {
		session = user.Session
	}

	c.Auth.SetCookies(w, uid, session)
	if user.IsAdmin {
		// Is this error check redundant? We already check for the error in PreRoute for the same IP
		// TODO: Should we be logging this?
		log.Printf("#%d has logged in with IP %s", uid, user.GetIP())
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func extractCookie(name string, r *http.Request) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func mfaGetCookies(r *http.Request) (uid int, provSession string, signedSession string, err error) {
	suid, err := extractCookie("uid", r)
	if err != nil {
		return 0, "", "", err
	}
	uid, err = strconv.Atoi(suid)
	if err != nil {
		return 0, "", "", err
	}
	provSession, err = extractCookie("provSession", r)
	if err != nil {
		return 0, "", "", err
	}
	signedSession, err = extractCookie("signedSession", r)
	return uid, provSession, signedSession, err
}

func mfaVerifySession(provSession string, signedSession string, uid int) bool {
	h := sha256.New()
	h.Write([]byte(c.SessionSigningKeyBox.Load().(string)))
	h.Write([]byte(provSession))
	h.Write([]byte(strconv.Itoa(uid)))
	expected := hex.EncodeToString(h.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(signedSession), []byte(expected)) == 1 {
		return true
	}

	h = sha256.New()
	h.Write([]byte(c.OldSessionSigningKeyBox.Load().(string)))
	h.Write([]byte(provSession))
	h.Write([]byte(strconv.Itoa(uid)))
	expected = hex.EncodeToString(h.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(signedSession), []byte(expected)) == 1
}

func AccountLoginMFAVerify(w http.ResponseWriter, r *http.Request, user c.User, h *c.Header) c.RouteError {
	if user.Loggedin {
		return c.LocalError("You're already logged in.", w, r, user)
	}
	h.Title = p.GetTitlePhrase("login_mfa_verify")

	uid, provSession, signedSession, err := mfaGetCookies(r)
	if err != nil {
		return c.LocalError("Invalid cookie", w, r, user)
	}
	if !mfaVerifySession(provSession, signedSession, uid) {
		return c.LocalError("Invalid session", w, r, user)
	}

	return renderTemplate("login_mfa_verify", w, r, h, c.Page{h, tList, nil})
}

func AccountLoginMFAVerifySubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	uid, provSession, signedSession, err := mfaGetCookies(r)
	if err != nil {
		return c.LocalError("Invalid cookie", w, r, user)
	}
	if !mfaVerifySession(provSession, signedSession, uid) {
		return c.LocalError("Invalid session", w, r, user)
	}
	token := r.PostFormValue("mfa_token")

	err = c.Auth.ValidateMFAToken(token, uid)
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}

	return loginSuccess(uid, w, r, &user)
}

func AccountLogout(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	c.Auth.Logout(w, user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func AccountRegister(w http.ResponseWriter, r *http.Request, user c.User, h *c.Header) c.RouteError {
	if user.Loggedin {
		return c.LocalError("You're already logged in.", w, r, user)
	}
	h.Title = p.GetTitlePhrase("register")
	h.AddScriptAsync("register.js")
	return renderTemplate("register", w, r, h, c.Page{h, tList, h.Settings["activation_type"] != 2})
}

func isNumeric(data string) (numeric bool) {
	for _, char := range data {
		if char < 48 || char > 57 {
			return false
		}
	}
	return true
}

func AccountRegisterSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	headerLite, _ := c.SimpleUserCheck(w, r, &user)

	// TODO: Should we push multiple validation errors to the user instead of just one?
	regSuccess := true
	regErrMsg := ""
	regErrReason := ""
	regError := func(userMsg string, reason string) {
		regSuccess = false
		if regErrMsg == "" {
			regErrMsg = userMsg
		}
		regErrReason += reason + "|"
	}

	if r.PostFormValue("tos") != "0" {
		regError(p.GetErrorPhrase("register_might_be_machine"), "trap-question")
	}
	if !c.Config.DisableJSAntispam {
		h := sha256.New()
		h.Write([]byte(c.JSTokenBox.Load().(string)))
		h.Write([]byte(user.GetIP()))
		if r.PostFormValue("golden-watch") != hex.EncodeToString(h.Sum(nil)) {
			regError(p.GetErrorPhrase("register_might_be_machine"), "js-antispam")
		}
	}

	name := c.SanitiseSingleLine(r.PostFormValue("name"))
	if name == "" {
		regError(p.GetErrorPhrase("register_need_username"), "no-username")
	}
	// This is so a numeric name won't interfere with mentioning a user by ID, there might be a better way of doing this like perhaps !@ to mean IDs and @ to mean usernames in the pre-parser
	nameBits := strings.Split(name, " ")
	if isNumeric(nameBits[0]) {
		regError(p.GetErrorPhrase("register_first_word_numeric"), "numeric-name")
	}
	if strings.Contains(name, "http://") || strings.Contains(name, "https://") || strings.Contains(name, "ftp://") || strings.Contains(name, "ssh://") {
		regError(p.GetErrorPhrase("register_url_username"), "url-name")
	}

	// TODO: Add a dedicated function for validating emails
	email := c.SanitiseSingleLine(r.PostFormValue("email"))
	if headerLite.Settings["activation_type"] == 2 && email == "" {
		regError(p.GetErrorPhrase("register_need_email"), "no-email")
	}
	if c.HasSuspiciousEmail(email) {
		regError(p.GetErrorPhrase("register_suspicious_email"), "suspicious-email")
	}

	password := r.PostFormValue("password")
	// ?  Move this into Create()? What if we want to programatically set weak passwords for tests?
	err := c.WeakPassword(password, name, email)
	if err != nil {
		regError(err.Error(), "weak-password")
	} else {
		// Do the two inputted passwords match..?
		confirmPassword := r.PostFormValue("confirm_password")
		if password != confirmPassword {
			regError(p.GetErrorPhrase("register_password_mismatch"), "password-mismatch")
		}
	}

	regLog := c.RegLogItem{Username: name, Email: email, FailureReason: regErrReason, Success: regSuccess, IP: user.GetIP()}
	if !c.Config.DisableRegLog {
		_, err := regLog.Create()
		if err != nil {
			return c.InternalError(err, w, r)
		}
	}
	if !regSuccess {
		return c.LocalError(regErrMsg, w, r, user)
	}

	var active bool
	var group int
	switch headerLite.Settings["activation_type"] {
	case 1: // Activate All
		active = true
		group = c.Config.DefaultGroup
	default: // Anything else. E.g. Admin Activation or Email Activation.
		group = c.Config.ActivationGroup
	}

	addReason := func(reason string) error {
		if c.Config.DisableRegLog {
			return nil
		}
		regLog.FailureReason += reason
		return regLog.Commit()
	}

	// TODO: Do the registration attempt logging a little less messily (without having to amend the result after the insert)
	uid, err := c.Users.Create(name, password, email, group, active)
	if err != nil {
		regLog.Success = false
		if err == c.ErrAccountExists {
			err = addReason("username-exists")
			if err != nil {
				return c.InternalError(err, w, r)
			}
			return c.LocalError(p.GetErrorPhrase("register_username_unavailable"), w, r, user)
		} else if err == c.ErrLongUsername {
			err = addReason("username-too-long")
			if err != nil {
				return c.InternalError(err, w, r)
			}
			return c.LocalError(p.GetErrorPhrase("register_username_too_long_prefix")+strconv.Itoa(c.Config.MaxUsernameLength), w, r, user)
		}
		err2 := addReason("internal-error")
		if err2 != nil {
			return c.InternalError(err2, w, r)
		}
		return c.InternalError(err, w, r)
	}

	session, err := c.Auth.CreateSession(uid)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	c.Auth.SetCookies(w, uid, session)

	// Check if this user actually owns this email, if email activation is on, automatically flip their account to active when the email is validated. Validation is also useful for determining whether this user should receive any alerts, etc. via email
	if c.Site.EnableEmails && email != "" {
		token, err := c.GenerateSafeString(80)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		// TODO: Add an EmailStore and move this there
		_, err = qgen.NewAcc().Insert("emails").Columns("email,uid,validated,token").Fields("?,?,?,?").Exec(email, uid, 0, token)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		err = c.SendActivationEmail(name, email, token)
		if err != nil {
			return c.LocalError(p.GetErrorPhrase("register_email_fail"), w, r, user)
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

// TODO: Figure a way of making this into middleware?
func accountEditHead(titlePhrase string, w http.ResponseWriter, r *http.Request, user *c.User, h *c.Header) {
	h.Title = p.GetTitlePhrase(titlePhrase)
	h.Path = "/user/edit/"
	h.AddSheet(h.Theme.Name + "/account.css")
	h.AddScriptAsync("account.js")
}

func AccountEdit(w http.ResponseWriter, r *http.Request, user c.User, header *c.Header) c.RouteError {
	accountEditHead("account", w, r, &user, header)
	if r.FormValue("avatar_updated") == "1" {
		header.AddNotice("account_avatar_updated")
	} else if r.FormValue("username_updated") == "1" {
		header.AddNotice("account_username_updated")
	} else if r.FormValue("mfa_setup_success") == "1" {
		header.AddNotice("account_mfa_setup_success")
	}

	// TODO: Find a more efficient way of doing this
	mfaSetup := false
	_, err := c.MFAstore.Get(user.ID)
	if err != sql.ErrNoRows && err != nil {
		return c.InternalError(err, w, r)
	} else if err != sql.ErrNoRows {
		mfaSetup = true
	}

	// Normalise the score so that the user sees their relative progress to the next level rather than showing them their total score
	prevScore := c.GetLevelScore(user.Level)
	currentScore := user.Score - prevScore
	nextScore := c.GetLevelScore(user.Level+1) - prevScore
	perc := int(math.Ceil((float64(nextScore) / float64(currentScore)) * 100))

	pi := c.Account{header, "dashboard", "account_own_edit", c.AccountDashPage{header, mfaSetup, currentScore, nextScore, user.Level + 1, perc * 2}}
	return renderTemplate("account", w, r, header, pi)
}

//edit_password
func AccountEditPassword(w http.ResponseWriter, r *http.Request, user c.User, h *c.Header) c.RouteError {
	accountEditHead("account_password", w, r, &user, h)
	return renderTemplate("account_own_edit_password", w, r, h, c.Page{h, tList, nil})
}

// TODO: Require re-authentication if the user hasn't logged in in a while
func AccountEditPasswordSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	var realPassword, salt string
	currentPassword := r.PostFormValue("current-password")
	newPassword := r.PostFormValue("new-password")
	confirmPassword := r.PostFormValue("confirm-password")

	// TODO: Use a reusable statement
	err := qgen.NewAcc().Select("users").Columns("password,salt").Where("uid=?").QueryRow(user.ID).Scan(&realPassword, &salt)
	if err == sql.ErrNoRows {
		return c.LocalError("Your account no longer exists.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	err = c.CheckPassword(realPassword, currentPassword, salt)
	if err == c.ErrMismatchedHashAndPassword {
		return c.LocalError("That's not the correct password.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	if newPassword != confirmPassword {
		return c.LocalError("The two passwords don't match.", w, r, user)
	}
	c.SetPassword(user.ID, newPassword) // TODO: Limited version of WeakPassword()

	// Log the user out as a safety precaution
	c.Auth.ForceLogout(user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func AccountEditAvatarSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.UploadAvatars {
		return c.NoPermissions(w, r, user)
	}

	ext, ferr := c.UploadAvatar(w, r, user, user.ID)
	if ferr != nil {
		return ferr
	}
	ferr = c.ChangeAvatar("."+ext, w, r, user)
	if ferr != nil {
		return ferr
	}

	// TODO: Only schedule a resize if the avatar isn't tiny
	err := user.ScheduleAvatarResize()
	if err != nil {
		return c.InternalError(err, w, r)
	}
	http.Redirect(w, r, "/user/edit/?avatar_updated=1", http.StatusSeeOther)
	return nil
}

func AccountEditRevokeAvatarSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	ferr = c.ChangeAvatar("", w, r, user)
	if ferr != nil {
		return ferr
	}

	http.Redirect(w, r, "/user/edit/?avatar_updated=1", http.StatusSeeOther)
	return nil
}

func AccountEditUsernameSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	newUsername := c.SanitiseSingleLine(r.PostFormValue("account-new-username"))
	if newUsername == "" {
		return c.LocalError("You can't leave your username blank", w, r, user)
	}
	err := user.ChangeName(newUsername)
	if err != nil {
		return c.LocalError("Unable to change the username. Does someone else already have this name?", w, r, user)
	}

	http.Redirect(w, r, "/user/edit/?username_updated=1", http.StatusSeeOther)
	return nil
}

func AccountEditMFA(w http.ResponseWriter, r *http.Request, user c.User, h *c.Header) c.RouteError {
	accountEditHead("account_mfa", w, r, &user, h)

	mfaItem, err := c.MFAstore.Get(user.ID)
	if err != sql.ErrNoRows && err != nil {
		return c.InternalError(err, w, r)
	} else if err == sql.ErrNoRows {
		return c.LocalError("Two-factor authentication hasn't been setup on your account", w, r, user)
	}

	pi := c.Page{h, tList, mfaItem.Scratch}
	return renderTemplate("account_own_edit_mfa", w, r, h, pi)
}

// If not setup, generate a string, otherwise give an option to disable mfa given the right code
func AccountEditMFASetup(w http.ResponseWriter, r *http.Request, user c.User, h *c.Header) c.RouteError {
	accountEditHead("account_mfa_setup", w, r, &user, h)

	// Flash an error if mfa is already setup
	_, err := c.MFAstore.Get(user.ID)
	if err != sql.ErrNoRows && err != nil {
		return c.InternalError(err, w, r)
	} else if err != sql.ErrNoRows {
		return c.LocalError("You have already setup two-factor authentication", w, r, user)
	}

	// TODO: Entitise this?
	code, err := c.GenerateGAuthSecret()
	if err != nil {
		return c.InternalError(err, w, r)
	}

	pi := c.Page{h, tList, c.FriendlyGAuthSecret(code)}
	return renderTemplate("account_own_edit_mfa_setup", w, r, h, pi)
}

// Form should bounce the random mfa secret back and the otp to be verified server-side to reduce the chances of a bug arising on the JS side which makes every code mismatch
func AccountEditMFASetupSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	// Flash an error if mfa is already setup
	_, err := c.MFAstore.Get(user.ID)
	if err != sql.ErrNoRows && err != nil {
		return c.InternalError(err, w, r)
	} else if err != sql.ErrNoRows {
		return c.LocalError("You have already setup two-factor authentication", w, r, user)
	}

	code := r.PostFormValue("code")
	otp := r.PostFormValue("otp")
	ok, err := c.VerifyGAuthToken(code, otp)
	if err != nil {
		//fmt.Println("err: ", err)
		return c.LocalError("Something weird happened", w, r, user) // TODO: Log this error?
	}
	// TODO: Use AJAX for this
	if !ok {
		return c.LocalError("The token isn't right", w, r, user)
	}

	// TODO: How should we handle races where a mfa key is already setup? Right now, it's a fairly generic error, maybe try parsing the error message?
	err = c.MFAstore.Create(code, user.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/edit/?mfa_setup_success=1", http.StatusSeeOther)
	return nil
}

// TODO: Implement this
func AccountEditMFADisableSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	_, ferr := c.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	// Flash an error if mfa is already setup
	mfaItem, err := c.MFAstore.Get(user.ID)
	if err != sql.ErrNoRows && err != nil {
		return c.InternalError(err, w, r)
	} else if err == sql.ErrNoRows {
		return c.LocalError("You don't have two-factor enabled on your account", w, r, user)
	}

	err = mfaItem.Delete()
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/edit/?mfa_disabled=1", http.StatusSeeOther)
	return nil
}

func AccountEditPrivacy(w http.ResponseWriter, r *http.Request, user c.User, h *c.Header) c.RouteError {
	accountEditHead("account_privacy", w, r, &user, h)
	profileComments := false
	receiveConvos := false
	enableEmbeds := !c.DefaultParseSettings.NoEmbed
	if user.ParseSettings != nil {
		enableEmbeds = !user.ParseSettings.NoEmbed
	}
	pi := c.Account{h, "privacy", "account_own_edit_privacy", c.AccountPrivacyPage{h, profileComments, receiveConvos, enableEmbeds}}
	return renderTemplate("account", w, r, h, pi)
}

func AccountEditPrivacySubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	//headerLite, _ := c.SimpleUserCheck(w, r, &user)

	sEnableEmbeds := r.FormValue("enable_embeds")
	enableEmbeds, err := strconv.Atoi(sEnableEmbeds)
	if err != nil {
		return c.LocalError("enable_embeds must be 0 or 1", w, r, user)
	}
	if sEnableEmbeds != r.FormValue("o_enable_embeds") {
		err = (&user).UpdatePrivacy(enableEmbeds)
		if err != nil {
			return c.InternalError(err, w, r)
		}
	}

	http.Redirect(w, r, "/user/edit/privacy/?updated=1", http.StatusSeeOther)
	return nil
}

func AccountEditEmail(w http.ResponseWriter, r *http.Request, user c.User, h *c.Header) c.RouteError {
	accountEditHead("account_email", w, r, &user, h)
	emails, err := c.Emails.GetEmailsByUser(&user)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// Was this site migrated from another forum software? Most of them don't have multiple emails for a single user.
	// This also applies when the admin switches site.EnableEmails on after having it off for a while.
	if len(emails) == 0 && user.Email != "" {
		emails = append(emails, c.Email{UserID: user.ID, Email: user.Email, Validated: false, Primary: true})
	}

	if !c.Site.EnableEmails {
		h.AddNotice("account_mail_disabled")
	}
	if r.FormValue("verified") == "1" {
		h.AddNotice("account_mail_verify_success")
	}

	pi := c.Account{h, "edit_emails", "account_own_edit_email", c.EmailListPage{h, emails}}
	return renderTemplate("account", w, r, h, pi)
}

func AccountEditEmailAddSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	email := r.PostFormValue("email")
	_, err := c.Emails.Get(&user, email)
	if err == nil {
		return c.LocalError("You have already added this email.", w, r, user)
	} else if err != sql.ErrNoRows && err != nil {
		return c.InternalError(err, w, r)
	}

	var token string
	if c.Site.EnableEmails {
		token, err = c.GenerateSafeString(80)
		if err != nil {
			return c.InternalError(err, w, r)
		}
	}
	err = c.Emails.Add(user.ID, email, token)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	if c.Site.EnableEmails {
		err = c.SendValidationEmail(user.Name, email, token)
		if err != nil {
			return c.LocalError(p.GetErrorPhrase("register_email_fail"), w, r, user)
		}
	}

	http.Redirect(w, r, "/user/edit/email/?added=1", http.StatusSeeOther)
	return nil
}

func AccountEditEmailRemoveSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	headerLite, _ := c.SimpleUserCheck(w, r, &user)
	email := r.PostFormValue("email")

	// Quick and dirty check
	_, err := c.Emails.Get(&user, email)
	if err == sql.ErrNoRows {
		return c.LocalError("This email isn't set on this user.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	if headerLite.Settings["activation_type"] == 2 && user.Email == email {
		return c.LocalError("You can't remove your primary email when mandatory email activation is enabled.", w, r, user)
	}

	err = c.Emails.Delete(user.ID, email)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/edit/email/?removed=1", http.StatusSeeOther)
	return nil
}

// TODO: Should we make this an AnonAction so someone can do this without being logged in?
func AccountEditEmailTokenSubmit(w http.ResponseWriter, r *http.Request, user c.User, token string) c.RouteError {
	header, ferr := c.UserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !c.Site.EnableEmails {
		http.Redirect(w, r, "/user/edit/email/", http.StatusSeeOther)
		return nil
	}

	targetEmail := c.Email{UserID: user.ID}
	emails, err := c.Emails.GetEmailsByUser(&user)
	if err == sql.ErrNoRows {
		return c.LocalError("A verification email was never sent for you!", w, r, user)
	} else if err != nil {
		// TODO: Better error if we don't have an email or it's not in the emails table for some reason
		return c.LocalError("You are not logged in", w, r, user)
	}
	for _, email := range emails {
		if subtle.ConstantTimeCompare([]byte(email.Token), []byte(token)) == 1 {
			targetEmail = email
		}
	}

	if len(emails) == 0 {
		return c.LocalError("A verification email was never sent for you!", w, r, user)
	}
	if targetEmail.Token == "" {
		return c.LocalError("That's not a valid token!", w, r, user)
	}

	err = c.Emails.VerifyEmail(user.Email)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// If Email Activation is on, then activate the account while we're here
	if header.Settings["activation_type"] == 2 {
		if err = user.Activate(); err != nil {
			return c.InternalError(err, w, r)
		}
	}
	http.Redirect(w, r, "/user/edit/email/?verified=1", http.StatusSeeOther)
	return nil
}

func AccountLogins(w http.ResponseWriter, r *http.Request, user c.User, h *c.Header) c.RouteError {
	accountEditHead("account_logins", w, r, &user, h)
	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 12
	offset, page, lastPage := c.PageOffset(c.LoginLogs.CountUser(user.ID), page, perPage)

	logs, err := c.LoginLogs.GetOffset(user.ID, offset, perPage)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	pageList := c.Paginate(page, lastPage, 5)
	pi := c.Account{h, "logins", "account_logins", c.AccountLoginsPage{h, logs, c.Paginator{pageList, page, lastPage}}}
	return renderTemplate("account", w, r, h, pi)
}

func AccountBlocked(w http.ResponseWriter, r *http.Request, user c.User, h *c.Header) c.RouteError {
	accountEditHead("account_blocked", w, r, &user, h)
	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 12
	offset, page, lastPage := c.PageOffset(c.UserBlocks.BlockedByCount(user.ID), page, perPage)

	uids, err := c.UserBlocks.BlockedByOffset(user.ID, offset, perPage)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	var blocks []*c.User
	for _, uid := range uids {
		u, err := c.Users.Get(uid)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		blocks = append(blocks, u)
	}

	pageList := c.Paginate(page, lastPage, 5)
	pi := c.Account{h, "blocked", "account_blocked", c.AccountBlocksPage{h, blocks, c.Paginator{pageList, page, lastPage}}}
	return renderTemplate("account", w, r, h, pi)
}

func LevelList(w http.ResponseWriter, r *http.Request, user c.User, h *c.Header) c.RouteError {
	h.Title = p.GetTitlePhrase("account_level_list")

	fScores := c.GetLevels(20)
	levels := make([]c.LevelListItem, len(fScores))
	for i, fScore := range fScores {
		var status string
		if user.Level > i {
			status = "complete"
		} else if user.Level < i {
			status = "future"
		} else {
			status = "inprogress"
		}
		iScore := int(math.Ceil(fScore))
		perc := int(math.Ceil((fScore / float64(user.Score)) * 100))
		levels[i] = c.LevelListItem{i, iScore, status, perc * 2}
	}

	return renderTemplate("level_list", w, r, h, c.LevelListPage{h, levels[1:]})
}

func Alerts(w http.ResponseWriter, r *http.Request, user c.User, h *c.Header) c.RouteError {
	return nil
}

func AccountPasswordReset(w http.ResponseWriter, r *http.Request, user c.User, h *c.Header) c.RouteError {
	if user.Loggedin {
		return c.LocalError("You're already logged in.", w, r, user)
	}
	if !c.Site.EnableEmails {
		return c.LocalError(p.GetNoticePhrase("account_mail_disabled"), w, r, user)
	}
	if r.FormValue("email_sent") == "1" {
		h.AddNotice("password_reset_email_sent")
	}
	h.Title = p.GetTitlePhrase("password_reset")
	return renderTemplate("password_reset", w, r, h, c.Page{h, tList, nil})
}

// TODO: Ratelimit this
func AccountPasswordResetSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	if user.Loggedin {
		return c.LocalError("You're already logged in.", w, r, user)
	}
	if !c.Site.EnableEmails {
		return c.LocalError(p.GetNoticePhrase("account_mail_disabled"), w, r, user)
	}

	username := r.PostFormValue("username")
	tuser, err := c.Users.GetByName(username)
	if err == sql.ErrNoRows {
		// Someone trying to stir up trouble?
		http.Redirect(w, r, "/accounts/password-reset/?email_sent=1", http.StatusSeeOther)
		return nil
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	token, err := c.GenerateSafeString(80)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// TODO: Move these queries somewhere else
	var disc string
	err = qgen.NewAcc().Select("password_resets").Columns("createdAt").DateCutoff("createdAt", 1, "hour").QueryRow().Scan(&disc)
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	if err == nil {
		return c.LocalError("You can only send a password reset email for a user once an hour", w, r, user)
	}

	count, err := qgen.NewAcc().Count("password_resets").DateCutoff("createdAt", 6, "hour").Total()
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	if count >= 3 {
		return c.LocalError("You can only send a password reset email for a user three times every six hours", w, r, user)
	}

	count, err = qgen.NewAcc().Count("password_resets").DateCutoff("createdAt", 12, "hour").Total()
	if err != nil && err != sql.ErrNoRows {
		return c.InternalError(err, w, r)
	}
	if count >= 4 {
		return c.LocalError("You can only send a password reset email for a user four times every twelve hours", w, r, user)
	}

	err = c.PasswordResetter.Create(tuser.Email, tuser.ID, token)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	var s string
	if c.Config.SslSchema {
		s = "s"
	}

	err = c.SendEmail(tuser.Email, p.GetTmplPhrase("password_reset_subject"), p.GetTmplPhrasef("password_reset_body", tuser.Name, "http"+s+"://"+c.Site.URL+"/accounts/password-reset/token/?uid="+strconv.Itoa(tuser.ID)+"&token="+token))
	if err != nil {
		return c.LocalError(p.GetErrorPhrase("password_reset_email_fail"), w, r, user)
	}

	http.Redirect(w, r, "/accounts/password-reset/?email_sent=1", http.StatusSeeOther)
	return nil
}

func AccountPasswordResetToken(w http.ResponseWriter, r *http.Request, user c.User, header *c.Header) c.RouteError {
	if user.Loggedin {
		return c.LocalError("You're already logged in.", w, r, user)
	}
	// TODO: Find a way to flash this notice
	/*if r.FormValue("token_verified") == "1" {
		header.AddNotice("password_reset_token_token_verified")
	}*/

	uid, err := strconv.Atoi(r.FormValue("uid"))
	if err != nil {
		return c.LocalError("Invalid uid", w, r, user)
	}
	token := r.FormValue("token")
	err = c.PasswordResetter.ValidateToken(uid, token)
	if err == sql.ErrNoRows || err == c.ErrBadResetToken {
		return c.LocalError("This reset token has expired.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	_, err = c.MFAstore.Get(uid)
	if err != sql.ErrNoRows && err != nil {
		return c.InternalError(err, w, r)
	}
	mfa := err != sql.ErrNoRows

	header.Title = p.GetTitlePhrase("password_reset_token")
	return renderTemplate("password_reset_token", w, r, header, c.ResetPage{header, uid, html.EscapeString(token), mfa})
}

func AccountPasswordResetTokenSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	if user.Loggedin {
		return c.LocalError("You're already logged in.", w, r, user)
	}
	uid, err := strconv.Atoi(r.FormValue("uid"))
	if err != nil {
		return c.LocalError("Invalid uid", w, r, user)
	}
	if !c.Users.Exists(uid) {
		return c.LocalError("This reset token has expired.", w, r, user)
	}

	err = c.PasswordResetter.ValidateToken(uid, r.FormValue("token"))
	if err == sql.ErrNoRows || err == c.ErrBadResetToken {
		return c.LocalError("This reset token has expired.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	mfaToken := r.PostFormValue("mfa_token")
	err = c.Auth.ValidateMFAToken(mfaToken, uid)
	if err != nil && err != c.ErrNoMFAToken {
		return c.LocalError(err.Error(), w, r, user)
	}

	newPassword := r.PostFormValue("password")
	confirmPassword := r.PostFormValue("confirm_password")
	if newPassword != confirmPassword {
		return c.LocalError("The two passwords don't match.", w, r, user)
	}
	c.SetPassword(uid, newPassword) // TODO: Limited version of WeakPassword()

	err = c.PasswordResetter.FlushTokens(uid)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// Log the user out as a safety precaution
	c.Auth.ForceLogout(uid)

	//http.Redirect(w, r, "/accounts/password-reset/token/?token_verified=1", http.StatusSeeOther)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}
