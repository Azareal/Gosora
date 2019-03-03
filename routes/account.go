package routes

import (
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
	"github.com/Azareal/Gosora/query_gen"
)

// A blank list to fill out that parameter in Page for routes which don't use it
var tList []interface{}

func AccountLogin(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header) common.RouteError {
	if user.Loggedin {
		return common.LocalError("You're already logged in.", w, r, user)
	}
	header.Title = phrases.GetTitlePhrase("login")
	pi := common.Page{header, tList, nil}
	return renderTemplate("login", w, r, header, pi)
}

// TODO: Log failed attempted logins?
// TODO: Lock IPS out if they have too many failed attempts?
// TODO: Log unusual countries in comparison to the country a user usually logs in from? Alert the user about this?
func AccountLoginSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	if user.Loggedin {
		return common.LocalError("You're already logged in.", w, r, user)
	}

	username := common.SanitiseSingleLine(r.PostFormValue("username"))
	uid, err, requiresExtraAuth := common.Auth.Authenticate(username, r.PostFormValue("password"))
	if err != nil {
		{
			// TODO: uid is currently set to 0 as authenticate fetches the user by username and password. Get the actual uid, so we can alert the user of attempted logins? What if someone takes advantage of the response times to deduce if an account exists?
			logItem := &common.LoginLogItem{UID: uid, Success: false, IPAddress: user.LastIP}
			_, err := logItem.Create()
			if err != nil {
				return common.InternalError(err, w, r)
			}
		}
		return common.LocalError(err.Error(), w, r, user)
	}

	// TODO: Take 2FA into account
	logItem := &common.LoginLogItem{UID: uid, Success: true, IPAddress: user.LastIP}
	_, err = logItem.Create()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Do we want to slacken this by only doing it when the IP changes?
	if requiresExtraAuth {
		provSession, signedSession, err := common.Auth.CreateProvisionalSession(uid)
		if err != nil {
			return common.InternalError(err, w, r)
		}
		// TODO: Use the login log ID in the provisional cookie?
		common.Auth.SetProvisionalCookies(w, uid, provSession, signedSession)
		http.Redirect(w, r, "/accounts/mfa_verify/", http.StatusSeeOther)
		return nil
	}

	return loginSuccess(uid, w, r, &user)
}

func loginSuccess(uid int, w http.ResponseWriter, r *http.Request, user *common.User) common.RouteError {
	userPtr, err := common.Users.Get(uid)
	if err != nil {
		return common.LocalError("Bad account", w, r, *user)
	}
	*user = *userPtr

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
	h.Write([]byte(common.SessionSigningKeyBox.Load().(string)))
	h.Write([]byte(provSession))
	h.Write([]byte(strconv.Itoa(uid)))
	expected := hex.EncodeToString(h.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(signedSession), []byte(expected)) == 1 {
		return true
	}

	h = sha256.New()
	h.Write([]byte(common.OldSessionSigningKeyBox.Load().(string)))
	h.Write([]byte(provSession))
	h.Write([]byte(strconv.Itoa(uid)))
	expected = hex.EncodeToString(h.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(signedSession), []byte(expected)) == 1
}

func AccountLoginMFAVerify(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header) common.RouteError {
	if user.Loggedin {
		return common.LocalError("You're already logged in.", w, r, user)
	}
	header.Title = phrases.GetTitlePhrase("login_mfa_verify")

	uid, provSession, signedSession, err := mfaGetCookies(r)
	if err != nil {
		return common.LocalError("Invalid cookie", w, r, user)
	}
	if !mfaVerifySession(provSession, signedSession, uid) {
		return common.LocalError("Invalid session", w, r, user)
	}

	pi := common.Page{header, tList, nil}
	if common.RunPreRenderHook("pre_render_login_mfa_verify", w, r, &user, &pi) {
		return nil
	}
	err = header.Theme.RunTmpl("login_mfa_verify", pi, w)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	return nil
}

func AccountLoginMFAVerifySubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	uid, provSession, signedSession, err := mfaGetCookies(r)
	if err != nil {
		return common.LocalError("Invalid cookie", w, r, user)
	}
	if !mfaVerifySession(provSession, signedSession, uid) {
		return common.LocalError("Invalid session", w, r, user)
	}
	var token = r.PostFormValue("mfa_token")

	err = common.Auth.ValidateMFAToken(token, uid)
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	return loginSuccess(uid, w, r, &user)
}

func AccountLogout(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	common.Auth.Logout(w, user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

func AccountRegister(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header) common.RouteError {
	if user.Loggedin {
		return common.LocalError("You're already logged in.", w, r, user)
	}
	header.Title = phrases.GetTitlePhrase("register")
	header.LooseCSP = true
	pi := common.Page{header, tList, nil}
	return renderTemplate("register", w, r, header, pi)
}

func isNumeric(data string) (numeric bool) {
	for _, char := range data {
		if char < 48 || char > 57 {
			return false
		}
	}
	return true
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
		regError(phrases.GetErrorPhrase("register_might_be_machine"), "trap-question")
	}
	if !common.Config.DisableJSAntispam {
		h := sha256.New()
		h.Write([]byte(common.JSTokenBox.Load().(string)))
		h.Write([]byte(user.LastIP))
		if r.PostFormValue("golden-watch") != hex.EncodeToString(h.Sum(nil)) {
			regError(phrases.GetErrorPhrase("register_might_be_machine"), "js-antispam")
		}
	}

	username := common.SanitiseSingleLine(r.PostFormValue("username"))
	// TODO: Add a dedicated function for validating emails
	email := common.SanitiseSingleLine(r.PostFormValue("email"))
	if username == "" {
		regError(phrases.GetErrorPhrase("register_need_username"), "no-username")
	}
	if email == "" {
		regError(phrases.GetErrorPhrase("register_need_email"), "no-email")
	}

	// This is so a numeric name won't interfere with mentioning a user by ID, there might be a better way of doing this like perhaps !@ to mean IDs and @ to mean usernames in the pre-parser
	usernameBits := strings.Split(username, " ")
	if isNumeric(usernameBits[0]) {
		regError(phrases.GetErrorPhrase("register_first_word_numeric"), "numeric-name")
	}

	ok := common.HasSuspiciousEmail(email)
	if ok {
		regError(phrases.GetErrorPhrase("register_suspicious_email"), "suspicious-email")
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
			regError(phrases.GetErrorPhrase("register_password_mismatch"), "password-mismatch")
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
			return common.LocalError(phrases.GetErrorPhrase("register_username_unavailable"), w, r, user)
		} else if err == common.ErrLongUsername {
			regLog.FailureReason += "username-too-long"
			err = regLog.Commit()
			if err != nil {
				return common.InternalError(err, w, r)
			}
			return common.LocalError(phrases.GetErrorPhrase("register_username_too_long_prefix")+strconv.Itoa(common.Config.MaxUsernameLength), w, r, user)
		}
		regLog.FailureReason += "internal-error"
		err2 := regLog.Commit()
		if err2 != nil {
			return common.InternalError(err2, w, r)
		}
		return common.InternalError(err, w, r)
	}

	session, err := common.Auth.CreateSession(uid)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	common.Auth.SetCookies(w, uid, session)

	// Check if this user actually owns this email, if email activation is on, automatically flip their account to active when the email is validated. Validation is also useful for determining whether this user should receive any alerts, etc. via email
	if common.Site.EnableEmails {
		token, err := common.GenerateSafeString(80)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		// TODO: Add an EmailStore and move this there
		_, err = qgen.NewAcc().Insert("emails").Columns("email, uid, validated, token").Fields("?,?,?,?").Exec(email, uid, 0, token)
		if err != nil {
			return common.InternalError(err, w, r)
		}

		err = common.SendValidationEmail(username, email, token)
		if err != nil {
			common.LogWarning(err)
			return common.LocalError(phrases.GetErrorPhrase("register_email_fail"), w, r, user)
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

// TODO: Figure a way of making this into middleware?
func accountEditHead(titlePhrase string, w http.ResponseWriter, r *http.Request, user *common.User, header *common.Header) {
	header.Title = phrases.GetTitlePhrase(titlePhrase)
	header.Path = "/user/edit/"
	header.AddSheet(header.Theme.Name + "/account.css")
	header.AddScript("account.js")
}

func AccountEdit(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header) common.RouteError {
	accountEditHead("account", w, r, &user, header)

	if r.FormValue("avatar_updated") == "1" {
		header.AddNotice("account_avatar_updated")
	} else if r.FormValue("username_updated") == "1" {
		header.AddNotice("account_username_updated")
	} else if r.FormValue("mfa_setup_success") == "1" {
		header.AddNotice("account_mfa_setup_success")
	}

	// TODO: Find a more efficient way of doing this
	var mfaSetup = false
	_, err := common.MFAstore.Get(user.ID)
	if err != sql.ErrNoRows && err != nil {
		return common.InternalError(err, w, r)
	} else if err != sql.ErrNoRows {
		mfaSetup = true
	}

	// Normalise the score so that the user sees their relative progress to the next level rather than showing them their total score
	prevScore := common.GetLevelScore(user.Level)
	currentScore := user.Score - prevScore
	nextScore := common.GetLevelScore(user.Level+1) - prevScore
	perc := int(math.Ceil((float64(nextScore) / float64(currentScore)) * 100))

	pi := common.Account{header, "dashboard", "account_own_edit", common.AccountDashPage{header, mfaSetup, currentScore, nextScore, user.Level + 1, perc * 2}}
	return renderTemplate("account", w, r, header, pi)
}

//edit_password
func AccountEditPassword(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header) common.RouteError {
	accountEditHead("account_password", w, r, &user, header)
	pi := common.Page{header, tList, nil}
	return renderTemplate("account_own_edit_password", w, r, header, pi)
}

// TODO: Require re-authentication if the user hasn't logged in in a while
func AccountEditPasswordSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	var realPassword, salt string
	currentPassword := r.PostFormValue("account-current-password")
	newPassword := r.PostFormValue("account-new-password")
	confirmPassword := r.PostFormValue("account-confirm-password")

	// TODO: Use a reusable statement
	err := qgen.NewAcc().Select("users").Columns("password, salt").Where("uid = ?").QueryRow(user.ID).Scan(&realPassword, &salt)
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

func AccountEditAvatarSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	// We don't want multiple files
	// TODO: Are we doing this correctly?
	filenameMap := make(map[string]bool)
	for _, fheaders := range r.MultipartForm.File {
		for _, hdr := range fheaders {
			if hdr.Filename == "" {
				continue
			}
			filenameMap[hdr.Filename] = true
		}
	}
	if len(filenameMap) > 1 {
		return common.LocalError("You may only upload one avatar", w, r, user)
	}

	var ext string
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

				if !common.ImageFileExts.Contains(ext) {
					return common.LocalError("You can only use an image for your avatar", w, r, user)
				}
			}

			// TODO: Centralise this string, so we don't have to change it in two different places when it changes
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
	// TODO: Only schedule a resize if the avatar isn't tiny
	err = user.ScheduleAvatarResize()
	if err != nil {
		return common.InternalError(err, w, r)
	}
	http.Redirect(w, r, "/user/edit/?avatar_updated=1", http.StatusSeeOther)
	return nil
}

func AccountEditUsernameSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	newUsername := common.SanitiseSingleLine(r.PostFormValue("account-new-username"))
	if newUsername == "" {
		return common.LocalError("You can't leave your username blank", w, r, user)
	}
	err := user.ChangeName(newUsername)
	if err != nil {
		return common.LocalError("Unable to change the username. Does someone else already have this name?", w, r, user)
	}

	http.Redirect(w, r, "/user/edit/?username_updated=1", http.StatusSeeOther)
	return nil
}

func AccountEditMFA(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header) common.RouteError {
	accountEditHead("account_mfa", w, r, &user, header)

	mfaItem, err := common.MFAstore.Get(user.ID)
	if err != sql.ErrNoRows && err != nil {
		return common.InternalError(err, w, r)
	} else if err == sql.ErrNoRows {
		return common.LocalError("Two-factor authentication hasn't been setup on your account", w, r, user)
	}

	pi := common.Page{header, tList, mfaItem.Scratch}
	return renderTemplate("account_own_edit_mfa", w, r, header, pi)
}

// If not setup, generate a string, otherwise give an option to disable mfa given the right code
func AccountEditMFASetup(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header) common.RouteError {
	accountEditHead("account_mfa_setup", w, r, &user, header)

	// Flash an error if mfa is already setup
	_, err := common.MFAstore.Get(user.ID)
	if err != sql.ErrNoRows && err != nil {
		return common.InternalError(err, w, r)
	} else if err != sql.ErrNoRows {
		return common.LocalError("You have already setup two-factor authentication", w, r, user)
	}

	// TODO: Entitise this?
	code, err := common.GenerateGAuthSecret()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	pi := common.Page{header, tList, common.FriendlyGAuthSecret(code)}
	return renderTemplate("account_own_edit_mfa_setup", w, r, header, pi)
}

// Form should bounce the random mfa secret back and the otp to be verified server-side to reduce the chances of a bug arising on the JS side which makes every code mismatch
func AccountEditMFASetupSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	// Flash an error if mfa is already setup
	_, err := common.MFAstore.Get(user.ID)
	if err != sql.ErrNoRows && err != nil {
		return common.InternalError(err, w, r)
	} else if err != sql.ErrNoRows {
		return common.LocalError("You have already setup two-factor authentication", w, r, user)
	}

	var code = r.PostFormValue("code")
	var otp = r.PostFormValue("otp")
	ok, err := common.VerifyGAuthToken(code, otp)
	if err != nil {
		//fmt.Println("err: ", err)
		return common.LocalError("Something weird happened", w, r, user) // TODO: Log this error?
	}
	// TODO: Use AJAX for this
	if !ok {
		return common.LocalError("The token isn't right", w, r, user)
	}

	// TODO: How should we handle races where a mfa key is already setup? Right now, it's a fairly generic error, maybe try parsing the error message?
	err = common.MFAstore.Create(code, user.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/edit/?mfa_setup_success=1", http.StatusSeeOther)
	return nil
}

// TODO: Implement this
func AccountEditMFADisableSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimpleUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	// Flash an error if mfa is already setup
	mfaItem, err := common.MFAstore.Get(user.ID)
	if err != sql.ErrNoRows && err != nil {
		return common.InternalError(err, w, r)
	} else if err == sql.ErrNoRows {
		return common.LocalError("You don't have two-factor enabled on your account", w, r, user)
	}

	err = mfaItem.Delete()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/edit/?mfa_disabled=1", http.StatusSeeOther)
	return nil
}

func AccountEditEmail(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header) common.RouteError {
	accountEditHead("account_email", w, r, &user, header)
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

	pi := common.Account{header, "edit_emails", "account_own_edit_email", common.EmailListPage{header, emails}}
	return renderTemplate("account", w, r, header, pi)
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

func AccountLogins(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header) common.RouteError {
	accountEditHead("account_logins", w, r, &user, header)

	logCount := common.LoginLogs.Count(user.ID)
	page, _ := strconv.Atoi(r.FormValue("page"))
	perPage := 12
	offset, page, lastPage := common.PageOffset(logCount, page, perPage)

	logs, err := common.LoginLogs.GetOffset(user.ID, offset, perPage)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	pageList := common.Paginate(logCount, perPage, 5)
	pi := common.Account{header, "logins", "account_logins", common.AccountLoginsPage{header, logs, common.Paginator{pageList, page, lastPage}}}
	return renderTemplate("account", w, r, header, pi)
}

func LevelList(w http.ResponseWriter, r *http.Request, user common.User, header *common.Header) common.RouteError {
	header.Title = phrases.GetTitlePhrase("account_level_list")

	var fScores = common.GetLevels(20)
	var levels = make([]common.LevelListItem, len(fScores))
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
		levels[i] = common.LevelListItem{i, iScore, status, perc * 2}
	}

	pi := common.LevelListPage{header, levels[1:]}
	return renderTemplate("level_list", w, r, header, pi)
}
