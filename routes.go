/*
*
*	Gosora Route Handlers
*	Copyright Azareal 2016 - 2020
*
 */
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

// A blank list to fill out that parameter in Page for routes which don't use it
var tList []interface{}
var successJSONBytes = []byte(`{"success":"1"}`)

// TODO: Refactor this
// TODO: Use the phrase system
var phraseLoginAlerts = []byte(`{"msgs":[{"msg":"Login to see your alerts","path":"/accounts/login"}],"msgCount":0}`)

// TODO: Refactor this endpoint
// TODO: Move this into the routes package
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
	// TODO: Split this into it's own function
	case "dismiss-alert":
		asid, err := strconv.Atoi(r.FormValue("asid"))
		if err != nil {
			return common.PreErrorJS("Invalid asid", w, r)
		}
		res, err := stmts.deleteActivityStreamMatch.Exec(user.ID, asid)
		if err != nil {
			return common.InternalError(err, w, r)
		}
		count, err := res.RowsAffected()
		if err != nil {
			return common.InternalError(err, w, r)
		}
		// Don't want to throw an internal error due to a socket closing
		if common.EnableWebsockets && count > 0 {
			_ = common.WsHub.PushMessage(user.ID, `{"event":"dismiss-alert","asid":`+strconv.Itoa(asid)+`}`)
		}
	// TODO: Split this into it's own function
	case "alerts": // A feed of events tailored for a specific user
		if !user.Loggedin {
			w.Write(phraseLoginAlerts)
			return nil
		}

		var msglist string
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

		var actors []int
		var alerts []common.Alert
		for rows.Next() {
			var alert common.Alert
			err = rows.Scan(&alert.ASID, &alert.ActorID, &alert.TargetUserID, &alert.Event, &alert.ElementType, &alert.ElementID)
			if err != nil {
				return common.InternalErrorJS(err, w, r)
			}
			alerts = append(alerts, alert)
			actors = append(actors, alert.ActorID)
		}
		err = rows.Err()
		if err != nil {
			return common.InternalErrorJS(err, w, r)
		}

		// Might not want to error here, if the account was deleted properly, we might want to figure out how we should handle deletions in general
		list, err := common.Users.BulkGetMap(actors)
		if err != nil {
			log.Print("actors:", actors)
			return common.InternalErrorJS(err, w, r)
		}

		var ok bool
		for _, alert := range alerts {
			alert.Actor, ok = list[alert.ActorID]
			if !ok {
				return common.InternalErrorJS(errors.New("No such actor"), w, r)
			}

			res, err := common.BuildAlert(alert, user)
			if err != nil {
				return common.LocalErrorJS(err.Error(), w, r)
			}

			msglist += res + ","
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

// TODO: Remove this line after we move routeAPIPhrases to the routes package
var cacheControlMaxAge = "max-age=" + strconv.Itoa(int(common.Day))

// TODO: Be careful with exposing the panel phrases here, maybe move them into a different namespace? We also need to educate the admin that phrases aren't necessarily secret
// TODO: Move to the routes package
var phraseWhitelist = []string{
	"topic",
	"status",
	"alerts",
	"paginator",
}

func routeAPIPhrases(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	// TODO: Don't make this too JSON dependent so that we can swap in newer more efficient formats
	h := w.Header()
	h.Set("Content-Type", "application/json")
	h.Set("Cache-Control", cacheControlMaxAge) //Cache-Control: max-age=31536000

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

	var plist map[string]string
	// A little optimisation to avoid copying entries from one map to the other, if we don't have to mutate it
	// TODO: Reduce the amount of duplication here
	if len(positives) > 1 {
		plist = make(map[string]string)
		for _, positive := range positives {
			// ! Constrain it to a subset of phrases for now
			var ok = false
			for _, item := range phraseWhitelist {
				if strings.HasPrefix(positive, item) {
					ok = true
					break
				}
			}
			if !ok {
				return common.PreErrorJS("Outside of phrase prefix whitelist", w, r)
			}
			pPhrases, ok := phrases.GetTmplPhrasesByPrefix(positive)
			if !ok {
				return common.PreErrorJS("No such prefix", w, r)
			}
			for name, phrase := range pPhrases {
				plist[name] = phrase
			}
		}
	} else {
		// ! Constrain it to a subset of phrases for now
		var ok = false
		for _, item := range phraseWhitelist {
			if strings.HasPrefix(positives[0], item) {
				ok = true
				break
			}
		}
		if !ok {
			return common.PreErrorJS("Outside of phrase prefix whitelist", w, r)
		}
		pPhrases, ok := phrases.GetTmplPhrasesByPrefix(positives[0])
		if !ok {
			return common.PreErrorJS("No such prefix", w, r)
		}
		plist = pPhrases
	}

	for _, negation := range negations {
		for name, _ := range plist {
			if strings.HasPrefix(name, negation) {
				delete(plist, name)
			}
		}
	}

	// TODO: Cache the output of this, especially for things like topic, so we don't have to waste more time than we need on this
	jsonBytes, err := json.Marshal(plist)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	w.Write(jsonBytes)

	return nil
}

// A dedicated function so we can shake things up every now and then to make the token harder to parse
// TODO: Are we sure we want to do this by ID, just in case we reuse this and have multiple antispams on the page?
func routeJSAntispam(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	h := sha256.New()
	h.Write([]byte(common.JSTokenBox.Load().(string)))
	h.Write([]byte(user.LastIP))
	jsToken := hex.EncodeToString(h.Sum(nil))

	var innerCode = "`document.getElementByld('golden-watch').value = '" + jsToken + "';`"
	w.Write([]byte(`let hihi = ` + innerCode + `;
hihi = hihi.replace('ld','Id');
eval(hihi);`))

	return nil
}
