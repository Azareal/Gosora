/*
*
*	Gosora Route Handlers
*	Copyright Azareal 2016 - 2018
*
 */
package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"unicode"

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

// TODO: Be careful with exposing the panel phrases here, maybe move them into a different namespace? We also need to educate the admin that phrases aren't necessarily secret
func routeAPIPhrases(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	// TODO: Don't make this too JSON dependent so that we can swap in newer more efficient formats
	w.Header().Set("Content-Type", "application/json")
	err := r.ParseForm()
	if err != nil {
		return common.PreErrorJS("Bad Form", w, r)
	}

	query := r.FormValue("query")
	if query == "" {
		return common.PreErrorJS("No query provided", w, r)
	}

	var negations []string
	var positives []string

	queryBits := strings.Split(query, ",")
	for _, queryBit := range queryBits {
		queryBit = strings.TrimSpace(queryBit)
		if queryBit[0] == '!' && len(queryBit) > 1 {
			queryBit = strings.TrimPrefix(queryBit, "!")
			for _, char := range queryBit {
				if !unicode.IsLetter(char) && char != '-' && char != '_' {
					return common.PreErrorJS("No symbols allowed, only - and _", w, r)
				}
			}
			negations = append(negations, queryBit)
		} else {
			for _, char := range queryBit {
				if !unicode.IsLetter(char) && char != '-' && char != '_' {
					return common.PreErrorJS("No symbols allowed, only - and _", w, r)
				}
			}
			positives = append(positives, queryBit)
		}
	}
	if len(positives) == 0 {
		return common.PreErrorJS("You haven't requested any phrases", w, r)
	}

	var phrases map[string]string
	// A little optimisation to avoid copying entries from one map to the other, if we don't have to mutate it
	if len(positives) > 1 {
		phrases = make(map[string]string)
		for _, positive := range positives {
			// ! Constrain it to topic and status phrases for now
			if !strings.HasPrefix(positive, "topic") && !strings.HasPrefix(positive, "status") {
				return common.PreErrorJS("Not implemented!", w, r)
			}
			pPhrases, ok := common.GetTmplPhrasesByPrefix(positive)
			if !ok {
				return common.PreErrorJS("No such prefix", w, r)
			}
			for name, phrase := range pPhrases {
				phrases[name] = phrase
			}
		}
	} else {
		// ! Constrain it to topic and status phrases for now
		if !strings.HasPrefix(positives[0], "topic") && !strings.HasPrefix(positives[0], "status") {
			return common.PreErrorJS("Not implemented!", w, r)
		}
		pPhrases, ok := common.GetTmplPhrasesByPrefix(positives[0])
		if !ok {
			return common.PreErrorJS("No such prefix", w, r)
		}
		phrases = pPhrases
	}

	for _, negation := range negations {
		for name, _ := range phrases {
			if strings.HasPrefix(name, negation) {
				delete(phrases, name)
			}
		}
	}

	// TODO: Cache the output of this, especially for things like topic, so we don't have to waste more time than we need on this
	jsonBytes, err := json.Marshal(phrases)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	w.Write(jsonBytes)

	return nil
}
