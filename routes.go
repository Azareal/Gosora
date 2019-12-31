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
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

// A blank list to fill out that parameter in Page for routes which don't use it
var tList []interface{}
var successJSONBytes = []byte(`{"success":1}`)

// TODO: Refactor this
// TODO: Use the phrase system
var phraseLoginAlerts = []byte(`{"msgs":[{"msg":"Login to see your alerts","path":"/accounts/login"}],"count":0}`)

// TODO: Refactor this endpoint
// TODO: Move this into the routes package
func routeAPI(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	// TODO: Don't make this too JSON dependent so that we can swap in newer more efficient formats
	w.Header().Set("Content-Type", "application/json")
	err := r.ParseForm()
	if err != nil {
		return c.PreErrorJS("Bad Form", w, r)
	}

	action := r.FormValue("action")
	if action == "" {
		action = "get"
	}
	if action != "get" && action != "set" {
		return c.PreErrorJS("Invalid Action", w, r)
	}

	switch r.FormValue("module") {
	// TODO: Split this into it's own function
	case "dismiss-alert":
		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			return c.PreErrorJS("Invalid id", w, r)
		}
		res, err := stmts.deleteActivityStreamMatch.Exec(user.ID, id)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		count, err := res.RowsAffected()
		if err != nil {
			return c.InternalError(err, w, r)
		}
		// Don't want to throw an internal error due to a socket closing
		if c.EnableWebsockets && count > 0 {
			_ = c.WsHub.PushMessage(user.ID, `{"event":"dismiss-alert","id":`+strconv.Itoa(id)+`}`)
		}
		w.Write(successJSONBytes)
	// TODO: Split this into it's own function
	case "alerts": // A feed of events tailored for a specific user
		if !user.Loggedin {
			var etag string
			_, ok := w.(c.GzipResponseWriter)
			if ok {
				etag = "\"" + strconv.FormatInt(c.StartTime.Unix(), 10) + "-ng\""
			} else {
				etag = "\"" + strconv.FormatInt(c.StartTime.Unix(), 10) + "-n\""
			}
			w.Header().Set("ETag", etag)
			if match := r.Header.Get("If-None-Match"); match != "" {
				if strings.Contains(match, etag) {
					w.WriteHeader(http.StatusNotModified)
					return nil
				}
			}
			w.Write(phraseLoginAlerts)
			return nil
		}

		var count int
		err = stmts.getActivityCountByWatcher.QueryRow(user.ID).Scan(&count)
		if err == ErrNoRows {
			return c.PreErrorJS("Unable to get the activity count", w, r)
		} else if err != nil {
			return c.InternalErrorJS(err, w, r)
		}

		rCreatedAt, _ := strconv.ParseInt(r.FormValue("t"), 10, 64)
		rCount, _ := strconv.Atoi(r.FormValue("c"))
		//log.Print("rCreatedAt:", rCreatedAt)
		//log.Print("rCount:", rCount)
		var actors []int
		var alerts []c.Alert
		var createdAt time.Time
		var topCreatedAt int64

		if count != 0 {
			rows, err := stmts.getActivityFeedByWatcher.Query(user.ID)
			if err != nil {
				return c.InternalErrorJS(err, w, r)
			}
			defer rows.Close()

			for rows.Next() {
				var al c.Alert
				err = rows.Scan(&al.ASID, &al.ActorID, &al.TargetUserID, &al.Event, &al.ElementType, &al.ElementID, &createdAt)
				if err != nil {
					return c.InternalErrorJS(err, w, r)
				}

				uCreatedAt := createdAt.Unix()
				//log.Print("uCreatedAt", uCreatedAt)
				//if rCreatedAt == 0 || rCreatedAt < uCreatedAt {
				alerts = append(alerts, al)
				actors = append(actors, al.ActorID)
				//}
				if uCreatedAt > topCreatedAt {
					topCreatedAt = uCreatedAt
				}
			}
			err = rows.Err()
			if err != nil {
				return c.InternalErrorJS(err, w, r)
			}
		}

		// Might not want to error here, if the account was deleted properly, we might want to figure out how we should handle deletions in general
		list, err := c.Users.BulkGetMap(actors)
		if err != nil {
			log.Print("actors:", actors)
			return c.InternalErrorJS(err, w, r)
		}

		var msglist string
		//var sb strings.Builder
		var ok bool
		for _, alert := range alerts {
			alert.Actor, ok = list[alert.ActorID]
			if !ok {
				return c.InternalErrorJS(errors.New("No such actor"), w, r)
			}
			res, err := c.BuildAlert(alert, user)
			if err != nil {
				return c.LocalErrorJS(err.Error(), w, r)
			}
			//sb.Write(res)
			msglist += res + ","
		}
		if len(msglist) != 0 {
			msglist = msglist[0 : len(msglist)-1]
		}

		if count == 0 || msglist == "" || (rCreatedAt != 0 && rCreatedAt >= topCreatedAt && count == rCount) {
			_, _ = io.WriteString(w, `{}`)
		} else {
			_, _ = io.WriteString(w, `{"msgs":[`+msglist+`],"count":`+strconv.Itoa(count)+`,"tc":`+strconv.Itoa(int(topCreatedAt))+`}`)
		}
	default:
		return c.PreErrorJS("Invalid Module", w, r)
	}
	return nil
}

// TODO: Remove this line after we move routeAPIPhrases to the routes package
var cacheControlMaxAge = "max-age=" + strconv.Itoa(int(c.Day))

// TODO: Be careful with exposing the panel phrases here, maybe move them into a different namespace? We also need to educate the admin that phrases aren't necessarily secret
// TODO: Move to the routes package
var phraseWhitelist = []string{
	"topic",
	"status",
	"alerts",
	"paginator",
	"analytics",

	"panel", // We're going to handle this specially below as this is a security boundary
}

func routeAPIPhrases(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	// TODO: Don't make this too JSON dependent so that we can swap in newer more efficient formats
	h := w.Header()
	h.Set("Content-Type", "application/json")

	err := r.ParseForm()
	if err != nil {
		return c.PreErrorJS("Bad Form", w, r)
	}
	query := r.FormValue("q")
	if query == "" {
		return c.PreErrorJS("No query provided", w, r)
	}

	var negations, positives []string
	for _, queryBit := range strings.Split(query, ",") {
		queryBit = strings.TrimSpace(queryBit)
		if queryBit[0] == '!' && len(queryBit) > 1 {
			queryBit = strings.TrimPrefix(queryBit, "!")
			for _, char := range queryBit {
				if !unicode.IsLetter(char) && char != '-' && char != '_' {
					return c.PreErrorJS("No symbols allowed, only - and _", w, r)
				}
			}
			negations = append(negations, queryBit)
		} else {
			for _, char := range queryBit {
				if !unicode.IsLetter(char) && char != '-' && char != '_' {
					return c.PreErrorJS("No symbols allowed, only - and _", w, r)
				}
			}
			positives = append(positives, queryBit)
		}
	}
	if len(positives) == 0 {
		return c.PreErrorJS("You haven't requested any phrases", w, r)
	}
	h.Set("Cache-Control", cacheControlMaxAge) //Cache-Control: max-age=31536000

	var etag string
	_, ok := w.(c.GzipResponseWriter)
	if ok {
		etag = "\"" + strconv.FormatInt(c.StartTime.Unix(), 10) + "-g\""
	} else {
		etag = "\"" + strconv.FormatInt(c.StartTime.Unix(), 10) + "\""
	}

	var plist map[string]string
	var notModified, private bool
	var posLoop = func(positive string) c.RouteError {
		// ! Constrain it to a subset of phrases for now
		for _, item := range phraseWhitelist {
			if strings.HasPrefix(positive, item) {
				// TODO: Break this down into smaller security boundaries based on control panel sections?
				// TODO: Do we have to be so strict with panel phrases?
				if strings.HasPrefix(positive, "panel") {
					private = true
					ok = user.IsSuperMod
				} else {
					ok = true
					if notModified {
						return nil
					}
					w.Header().Set("ETag", etag)
					match := r.Header.Get("If-None-Match")
					if match != "" && strings.Contains(match, etag) {
						notModified = true
						return nil
					}
				}
				break
			}
		}
		if !ok {
			return c.PreErrorJS("Outside of phrase prefix whitelist", w, r)
		}
		return nil
	}

	// A little optimisation to avoid copying entries from one map to the other, if we don't have to mutate it
	if len(positives) > 1 {
		plist = make(map[string]string)
		for _, positive := range positives {
			rerr := posLoop(positive)
			if rerr != nil {
				return rerr
			}
			pPhrases, ok := phrases.GetTmplPhrasesByPrefix(positive)
			if !ok {
				return c.PreErrorJS("No such prefix", w, r)
			}
			for name, phrase := range pPhrases {
				plist[name] = phrase
			}
		}
	} else {
		rerr := posLoop(positives[0])
		if rerr != nil {
			return rerr
		}
		pPhrases, ok := phrases.GetTmplPhrasesByPrefix(positives[0])
		if !ok {
			return c.PreErrorJS("No such prefix", w, r)
		}
		plist = pPhrases
	}

	if private {
		w.Header().Set("Cache-Control", "private")
	} else if notModified {
		w.WriteHeader(http.StatusNotModified)
		return nil
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
		return c.InternalError(err, w, r)
	}
	w.Write(jsonBytes)
	return nil
}

// A dedicated function so we can shake things up every now and then to make the token harder to parse
// TODO: Are we sure we want to do this by ID, just in case we reuse this and have multiple antispams on the page?
func routeJSAntispam(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	h := sha256.New()
	h.Write([]byte(c.JSTokenBox.Load().(string)))
	h.Write([]byte(user.GetIP()))
	jsToken := hex.EncodeToString(h.Sum(nil))

	innerCode := "`document.getElementByld('golden-watch').value = '" + jsToken + "';`"
	io.WriteString(w, `let hihi = `+innerCode+`;
hihi = hihi.replace('ld','Id');
eval(hihi);`)

	return nil
}
