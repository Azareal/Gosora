package routes

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/Azareal/Gosora/common"
)

func BanUserSubmit(w http.ResponseWriter, r *http.Request, user common.User, suid string) common.RouteError {
	if !user.Perms.BanUsers {
		return common.NoPermissions(w, r, user)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return common.LocalError("The provided UserID is not a valid number.", w, r, user)
	}
	if uid == -2 {
		return common.LocalError("Why don't you like Merlin?", w, r, user)
	}

	targetUser, err := common.Users.Get(uid)
	if err == sql.ErrNoRows {
		return common.LocalError("The user you're trying to ban no longer exists.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	// TODO: Is there a difference between IsMod and IsSuperMod? Should we delete the redundant one?
	if targetUser.IsMod {
		return common.LocalError("You may not ban another staff member.", w, r, user)
	}
	if uid == user.ID {
		return common.LocalError("Why are you trying to ban yourself? Stop that.", w, r, user)
	}
	if targetUser.IsBanned {
		return common.LocalError("The user you're trying to unban is already banned.", w, r, user)
	}

	durationDays, err := strconv.Atoi(r.FormValue("ban-duration-days"))
	if err != nil {
		return common.LocalError("You can only use whole numbers for the number of days", w, r, user)
	}

	durationWeeks, err := strconv.Atoi(r.FormValue("ban-duration-weeks"))
	if err != nil {
		return common.LocalError("You can only use whole numbers for the number of weeks", w, r, user)
	}

	durationMonths, err := strconv.Atoi(r.FormValue("ban-duration-months"))
	if err != nil {
		return common.LocalError("You can only use whole numbers for the number of months", w, r, user)
	}

	var duration time.Duration
	if durationDays > 1 && durationWeeks > 1 && durationMonths > 1 {
		duration, _ = time.ParseDuration("0")
	} else {
		var seconds int
		seconds += durationDays * int(common.Day)
		seconds += durationWeeks * int(common.Week)
		seconds += durationMonths * int(common.Month)
		duration, _ = time.ParseDuration(strconv.Itoa(seconds) + "s")
	}

	err = targetUser.Ban(duration, user.ID)
	if err == sql.ErrNoRows {
		return common.LocalError("The user you're trying to ban no longer exists.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	err = common.ModLogs.Create("ban", uid, "user", user.LastIP, user.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/"+strconv.Itoa(uid), http.StatusSeeOther)
	return nil
}

func UnbanUser(w http.ResponseWriter, r *http.Request, user common.User, suid string) common.RouteError {
	if !user.Perms.BanUsers {
		return common.NoPermissions(w, r, user)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return common.LocalError("The provided UserID is not a valid number.", w, r, user)
	}

	targetUser, err := common.Users.Get(uid)
	if err == sql.ErrNoRows {
		return common.LocalError("The user you're trying to unban no longer exists.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if !targetUser.IsBanned {
		return common.LocalError("The user you're trying to unban isn't banned.", w, r, user)
	}

	err = targetUser.Unban()
	if err == common.ErrNoTempGroup {
		return common.LocalError("The user you're trying to unban is not banned", w, r, user)
	} else if err == sql.ErrNoRows {
		return common.LocalError("The user you're trying to unban no longer exists.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	err = common.ModLogs.Create("unban", uid, "user", user.LastIP, user.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/user/"+strconv.Itoa(uid), http.StatusSeeOther)
	return nil
}

func ActivateUser(w http.ResponseWriter, r *http.Request, user common.User, suid string) common.RouteError {
	if !user.Perms.ActivateUsers {
		return common.NoPermissions(w, r, user)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return common.LocalError("The provided UserID is not a valid number.", w, r, user)
	}

	targetUser, err := common.Users.Get(uid)
	if err == sql.ErrNoRows {
		return common.LocalError("The account you're trying to activate no longer exists.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	if targetUser.Active {
		return common.LocalError("The account you're trying to activate has already been activated.", w, r, user)
	}
	err = targetUser.Activate()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	err = common.ModLogs.Create("activate", targetUser.ID, "user", user.LastIP, user.ID)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	http.Redirect(w, r, "/user/"+strconv.Itoa(targetUser.ID), http.StatusSeeOther)
	return nil
}
