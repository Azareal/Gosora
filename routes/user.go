package routes

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	c "github.com/Azareal/Gosora/common"
)

func BanUserSubmit(w http.ResponseWriter, r *http.Request, user c.User, suid string) c.RouteError {
	if !user.Perms.BanUsers {
		return c.NoPermissions(w, r, user)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return c.LocalError("The provided UserID is not a valid number.", w, r, user)
	}
	if uid == -2 {
		return c.LocalError("Why don't you like Merlin?", w, r, user)
	}

	targetUser, err := c.Users.Get(uid)
	if err == sql.ErrNoRows {
		return c.LocalError("The user you're trying to ban no longer exists.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	// TODO: Is there a difference between IsMod and IsSuperMod? Should we delete the redundant one?
	if targetUser.IsMod {
		return c.LocalError("You may not ban another staff member.", w, r, user)
	}
	if uid == user.ID {
		return c.LocalError("Why are you trying to ban yourself? Stop that.", w, r, user)
	}
	if targetUser.IsBanned {
		return c.LocalError("The user you're trying to unban is already banned.", w, r, user)
	}

	durationDays, err := strconv.Atoi(r.FormValue("ban-duration-days"))
	if err != nil {
		return c.LocalError("You can only use whole numbers for the number of days", w, r, user)
	}

	durationWeeks, err := strconv.Atoi(r.FormValue("ban-duration-weeks"))
	if err != nil {
		return c.LocalError("You can only use whole numbers for the number of weeks", w, r, user)
	}

	durationMonths, err := strconv.Atoi(r.FormValue("ban-duration-months"))
	if err != nil {
		return c.LocalError("You can only use whole numbers for the number of months", w, r, user)
	}

	var duration time.Duration
	if durationDays > 1 && durationWeeks > 1 && durationMonths > 1 {
		duration, _ = time.ParseDuration("0")
	} else {
		var seconds int
		seconds += durationDays * int(c.Day)
		seconds += durationWeeks * int(c.Week)
		seconds += durationMonths * int(c.Month)
		duration, _ = time.ParseDuration(strconv.Itoa(seconds) + "s")
	}

	err = targetUser.Ban(duration, user.ID)
	if err == sql.ErrNoRows {
		return c.LocalError("The user you're trying to ban no longer exists.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	err = c.ModLogs.Create("ban", uid, "user", user.LastIP, user.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// TODO: Trickle the hookTable down from the router
	hTbl := c.GetHookTable()
	skip, rerr := hTbl.VhookSkippable("action_end_ban_user", targetUser.ID, &user)
	if skip || rerr != nil {
		return rerr
	}

	http.Redirect(w, r, "/user/"+strconv.Itoa(uid), http.StatusSeeOther)
	return nil
}

func UnbanUser(w http.ResponseWriter, r *http.Request, user c.User, suid string) c.RouteError {
	if !user.Perms.BanUsers {
		return c.NoPermissions(w, r, user)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return c.LocalError("The provided UserID is not a valid number.", w, r, user)
	}

	targetUser, err := c.Users.Get(uid)
	if err == sql.ErrNoRows {
		return c.LocalError("The user you're trying to unban no longer exists.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	if !targetUser.IsBanned {
		return c.LocalError("The user you're trying to unban isn't banned.", w, r, user)
	}

	err = targetUser.Unban()
	if err == c.ErrNoTempGroup {
		return c.LocalError("The user you're trying to unban is not banned", w, r, user)
	} else if err == sql.ErrNoRows {
		return c.LocalError("The user you're trying to unban no longer exists.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	err = c.ModLogs.Create("unban", uid, "user", user.LastIP, user.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// TODO: Trickle the hookTable down from the router
	hTbl := c.GetHookTable()
	skip, rerr := hTbl.VhookSkippable("action_end_unban_user", targetUser.ID, &user)
	if skip || rerr != nil {
		return rerr
	}

	http.Redirect(w, r, "/user/"+strconv.Itoa(uid), http.StatusSeeOther)
	return nil
}

func ActivateUser(w http.ResponseWriter, r *http.Request, user c.User, suid string) c.RouteError {
	if !user.Perms.ActivateUsers {
		return c.NoPermissions(w, r, user)
	}

	uid, err := strconv.Atoi(suid)
	if err != nil {
		return c.LocalError("The provided UserID is not a valid number.", w, r, user)
	}

	targetUser, err := c.Users.Get(uid)
	if err == sql.ErrNoRows {
		return c.LocalError("The account you're trying to activate no longer exists.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	if targetUser.Active {
		return c.LocalError("The account you're trying to activate has already been activated.", w, r, user)
	}
	err = targetUser.Activate()
	if err != nil {
		return c.InternalError(err, w, r)
	}

	err = c.ModLogs.Create("activate", targetUser.ID, "user", user.LastIP, user.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// TODO: Trickle the hookTable down from the router
	hTbl := c.GetHookTable()
	skip, rerr := hTbl.VhookSkippable("action_end_activate_user", targetUser.ID, &user)
	if skip || rerr != nil {
		return rerr
	}

	http.Redirect(w, r, "/user/"+strconv.Itoa(targetUser.ID), http.StatusSeeOther)
	return nil
}
