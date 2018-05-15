/*
*
*	Gosora Route Handlers
*	Copyright Azareal 2016 - 2018
*
 */
package main

import (
	"net/http"
	"strconv"

	"./common"
)

// A blank list to fill out that parameter in Page for routes which don't use it
var tList []interface{}

//var nList []string
var successJSONBytes = []byte(`{"success":"1"}`)

// TODO: Refactor this
// TODO: Use the phrase system
var phraseLoginAlerts = []byte(`{"msgs":[{"msg":"Login to see your alerts","path":"/accounts/login"}]}`)

// TODO: Refactor this endpoint
func routeAPI(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	// TODO: Don't make this too JSON dependent so that we can swap in newer more efficient formats
	w.Header().Set("Content-Type", "application/json")
	err := r.ParseForm()
	if err != nil {
		return common.PreErrorJS("Bad Form", w, r)
	}

	action := r.FormValue("action")
	if action != "get" && action != "set" {
		return common.PreErrorJS("Invalid Action", w, r)
	}

	switch r.FormValue("module") {
	case "dismiss-alert":
		asid, err := strconv.Atoi(r.FormValue("asid"))
		if err != nil {
			return common.PreErrorJS("Invalid asid", w, r)
		}
		_, err = stmts.deleteActivityStreamMatch.Exec(user.ID, asid)
		if err != nil {
			return common.InternalError(err, w, r)
		}
	case "alerts": // A feed of events tailored for a specific user
		if !user.Loggedin {
			w.Write(phraseLoginAlerts)
			return nil
		}

		var msglist, event, elementType string
		var asid, actorID, targetUserID, elementID int
		var msgCount int

		err = stmts.getActivityCountByWatcher.QueryRow(user.ID).Scan(&msgCount)
		if err == ErrNoRows {
			return common.PreErrorJS("Couldn't find the parent topic", w, r)
		} else if err != nil {
			return common.InternalErrorJS(err, w, r)
		}

		rows, err := stmts.getActivityFeedByWatcher.Query(user.ID)
		if err != nil {
			return common.InternalErrorJS(err, w, r)
		}
		defer rows.Close()

		for rows.Next() {
			err = rows.Scan(&asid, &actorID, &targetUserID, &event, &elementType, &elementID)
			if err != nil {
				return common.InternalErrorJS(err, w, r)
			}
			res, err := common.BuildAlert(asid, event, elementType, actorID, targetUserID, elementID, user)
			if err != nil {
				return common.LocalErrorJS(err.Error(), w, r)
			}
			msglist += res + ","
		}
		err = rows.Err()
		if err != nil {
			return common.InternalErrorJS(err, w, r)
		}

		if len(msglist) != 0 {
			msglist = msglist[0 : len(msglist)-1]
		}
		_, _ = w.Write([]byte(`{"msgs":[` + msglist + `],"msgCount":` + strconv.Itoa(msgCount) + `}`))
	default:
		return common.PreErrorJS("Invalid Module", w, r)
	}
	return nil
}
