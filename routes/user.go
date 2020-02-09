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

	durDays, err := strconv.Atoi(r.FormValue("dur-days"))
	if err != nil {
		return c.LocalError("You can only use whole numbers for the number of days", w, r, user)
	}
	durWeeks, err := strconv.Atoi(r.FormValue("dur-weeks"))
	if err != nil {
		return c.LocalError("You can only use whole numbers for the number of weeks", w, r, user)
	}
	durMonths, err := strconv.Atoi(r.FormValue("dur-months"))
	if err != nil {
		return c.LocalError("You can only use whole numbers for the number of months", w, r, user)
	}
	deletePosts := false
	switch r.FormValue("delete-posts") {
	case "1":
		deletePosts = true
	}

	var dur time.Duration
	if durDays > 1 && durWeeks > 1 && durMonths > 1 {
		dur, _ = time.ParseDuration("0")
	} else {
		var secs int
		secs += durDays * int(c.Day)
		secs += durWeeks * int(c.Week)
		secs += durMonths * int(c.Month)
		dur, _ = time.ParseDuration(strconv.Itoa(secs) + "s")
	}

	err = targetUser.Ban(dur, user.ID)
	if err == sql.ErrNoRows {
		return c.LocalError("The user you're trying to ban no longer exists.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	err = c.ModLogs.Create("ban", uid, "user", user.GetIP(), user.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	if deletePosts {
		err = targetUser.DeletePosts()
		if err == sql.ErrNoRows {
			return c.LocalError("The user you're trying to ban no longer exists.", w, r, user)
		} else if err != nil {
			return c.InternalError(err, w, r)
		}
		err = c.ModLogs.Create("delete-posts", uid, "user", user.GetIP(), user.ID)
		if err != nil {
			return c.InternalError(err, w, r)
		}
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

	err = c.ModLogs.Create("unban", uid, "user", user.GetIP(), user.ID)
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

	targetUser, err = c.Users.Get(uid)
	if err == sql.ErrNoRows {
		return c.LocalError("The account you're trying to activate no longer exists.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	err = c.GroupPromotions.PromoteIfEligible(targetUser, targetUser.Level, targetUser.Posts, targetUser.CreatedAt)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	targetUser.CacheRemove()

	err = c.ModLogs.Create("activate", targetUser.ID, "user", user.GetIP(), user.ID)
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

func DeletePostsSubmit(w http.ResponseWriter, r *http.Request, user c.User, suid string) c.RouteError {
	if !user.Perms.BanUsers {
		return c.NoPermissions(w, r, user)
	}
	uid, err := strconv.Atoi(suid)
	if err != nil {
		return c.LocalError("The provided UserID is not a valid number.", w, r, user)
	}

	targetUser, err := c.Users.Get(uid)
	if err == sql.ErrNoRows {
		return c.LocalError("The user you're trying to purge posts of no longer exists.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	// TODO: Is there a difference between IsMod and IsSuperMod? Should we delete the redundant one?
	if targetUser.IsMod {
		return c.LocalError("You may not purge the posts of another staff member.", w, r, user)
	}

	err = targetUser.DeletePosts()
	if err == sql.ErrNoRows {
		return c.LocalError("The user you're trying to purge posts of no longer exists.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}
	err = c.ModLogs.Create("delete-posts", uid, "user", user.GetIP(), user.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// TODO: Trickle the hookTable down from the router
	hTbl := c.GetHookTable()
	skip, rerr := hTbl.VhookSkippable("action_end_delete_posts", targetUser.ID, &user)
	if skip || rerr != nil {
		return rerr
	}

	http.Redirect(w, r, "/user/"+strconv.Itoa(uid), http.StatusSeeOther)
	return nil
}
