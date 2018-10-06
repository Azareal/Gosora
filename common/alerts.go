/*
*
* Gosora Alerts System
* Copyright Azareal 2017 - 2018
*
 */
package common

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"../query_gen/lib"
)

type AlertStmts struct {
	addActivity      *sql.Stmt
	notifyWatchers   *sql.Stmt
	notifyOne        *sql.Stmt
	getWatchers      *sql.Stmt
	getActivityEntry *sql.Stmt
}

var alertStmts AlertStmts

// TODO: Move these statements into some sort of activity abstraction
// TODO: Rewrite the alerts logic
func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		alertStmts = AlertStmts{
			addActivity: acc.Insert("activity_stream").Columns("actor, targetUser, event, elementType, elementID").Fields("?,?,?,?,?").Prepare(),
			notifyWatchers: acc.SimpleInsertInnerJoin(
				qgen.DBInsert{"activity_stream_matches", "watcher, asid", ""},
				qgen.DBJoin{"activity_stream", "activity_subscriptions", "activity_subscriptions.user, activity_stream.asid", "activity_subscriptions.targetType = activity_stream.elementType AND activity_subscriptions.targetID = activity_stream.elementID AND activity_subscriptions.user != activity_stream.actor", "asid = ?", "", ""},
			),
			notifyOne:        acc.Insert("activity_stream_matches").Columns("watcher, asid").Fields("?,?").Prepare(),
			getWatchers:      acc.SimpleInnerJoin("activity_stream", "activity_subscriptions", "activity_subscriptions.user", "activity_subscriptions.targetType = activity_stream.elementType AND activity_subscriptions.targetID = activity_stream.elementID AND activity_subscriptions.user != activity_stream.actor", "asid = ?", "", ""),
			getActivityEntry: acc.Select("activity_stream").Columns("actor, targetUser, event, elementType, elementID").Where("asid = ?").Prepare(),
		}
		return acc.FirstError()
	})
}

// TODO: See if we can json.Marshal instead?
func escapeTextInJson(in string) string {
	in = strings.Replace(in, "\"", "\\\"", -1)
	return strings.Replace(in, "/", "\\/", -1)
}

func BuildAlert(asid int, event string, elementType string, actorID int, targetUserID int, elementID int, user User /* The current user */) (string, error) {
	var targetUser *User

	actor, err := Users.Get(actorID)
	if err != nil {
		return "", errors.New("Unable to find the actor")
	}

	/*if elementType != "forum" {
		targetUser, err = users.Get(targetUserID)
		if err != nil {
			LocalErrorJS("Unable to find the target user",w,r)
			return
		}
	}*/

	if event == "friend_invite" {
		return buildAlertString(GetTmplPhrase("alerts_new_friend_invite"), []string{actor.Name}, actor.Link, actor.Avatar, asid), nil
	}

	// Not that many events for us to handle in a forum
	if elementType == "forum" {
		if event == "reply" {
			topic, err := Topics.Get(elementID)
			if err != nil {
				DebugLogf("Unable to find linked topic %d", elementID)
				return "", errors.New(GetErrorPhrase("alerts_no_linked_topic"))
			}
			// Store the forum ID in the targetUser column instead of making a new one? o.O
			// Add an additional column for extra information later on when we add the ability to link directly to posts. We don't need the forum data for now...
			return buildAlertString(GetTmplPhrase("alerts_forum_new_topic"), []string{actor.Name, topic.Title}, topic.Link, actor.Avatar, asid), nil
		}
		return buildAlertString(GetTmplPhrase("alerts_forum_unknown_action"), []string{actor.Name}, "", actor.Avatar, asid), nil
	}

	var url, area string
	var phraseName = "alerts_" + elementType
	switch elementType {
	case "topic":
		topic, err := Topics.Get(elementID)
		if err != nil {
			DebugLogf("Unable to find linked topic %d", elementID)
			return "", errors.New(GetErrorPhrase("alerts_no_linked_topic"))
		}
		url = topic.Link
		area = topic.Title
		if targetUserID == user.ID {
			phraseName += "_own"
		}
	case "user":
		targetUser, err = Users.Get(elementID)
		if err != nil {
			DebugLogf("Unable to find target user %d", elementID)
			return "", errors.New("Unable to find the target user")
		}
		area = targetUser.Name
		url = targetUser.Link
		if targetUserID == user.ID {
			phraseName += "_own"
		}
	case "post":
		topic, err := TopicByReplyID(elementID)
		if err != nil {
			return "", errors.New("Unable to find the linked reply or parent topic")
		}
		url = topic.Link
		area = topic.Title
		if targetUserID == user.ID {
			phraseName += "_own"
		}
	default:
		return "", errors.New("Invalid elementType")
	}

	switch event {
	case "like":
		phraseName += "_like"
	case "mention":
		phraseName += "_mention"
	case "reply":
		phraseName += "_reply"
	}

	return buildAlertString(GetTmplPhrase(phraseName), []string{actor.Name, area}, url, actor.Avatar, asid), nil
}

func buildAlertString(msg string, sub []string, path string, avatar string, asid int) string {
	var substring string
	for _, item := range sub {
		substring += "\"" + escapeTextInJson(item) + "\","
	}
	if len(substring) > 0 {
		substring = substring[:len(substring)-1]
	}

	return `{"msg":"` + escapeTextInJson(msg) + `","sub":[` + substring + `],"path":"` + escapeTextInJson(path) + `","avatar":"` + escapeTextInJson(avatar) + `","asid":"` + strconv.Itoa(asid) + `"}`
}

func AddActivityAndNotifyAll(actor int, targetUser int, event string, elementType string, elementID int) error {
	res, err := alertStmts.addActivity.Exec(actor, targetUser, event, elementType, elementID)
	if err != nil {
		return err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	return NotifyWatchers(lastID)
}

func AddActivityAndNotifyTarget(actor int, targetUser int, event string, elementType string, elementID int) error {
	res, err := alertStmts.addActivity.Exec(actor, targetUser, event, elementType, elementID)
	if err != nil {
		return err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	err = NotifyOne(targetUser, lastID)
	if err != nil {
		return err
	}

	// Live alerts, if the target is online and WebSockets is enabled
	_ = WsHub.pushAlert(targetUser, int(lastID), event, elementType, actor, targetUser, elementID)
	return nil
}

func NotifyOne(watcher int, asid int64) error {
	_, err := alertStmts.notifyOne.Exec(watcher, asid)
	return err
}

func NotifyWatchers(asid int64) error {
	_, err := alertStmts.notifyWatchers.Exec(asid)
	if err != nil {
		return err
	}

	// Alert the subscribers about this without blocking us from doing something else
	if EnableWebsockets {
		go notifyWatchers(asid)
	}

	return nil
}

func notifyWatchers(asid int64) {
	rows, err := alertStmts.getWatchers.Query(asid)
	if err != nil && err != ErrNoRows {
		LogError(err)
		return
	}
	defer rows.Close()

	var uid int
	var uids []int
	for rows.Next() {
		err := rows.Scan(&uid)
		if err != nil {
			LogError(err)
			return
		}
		uids = append(uids, uid)
	}
	err = rows.Err()
	if err != nil {
		LogError(err)
		return
	}

	var actorID, targetUserID, elementID int
	var event, elementType string
	err = alertStmts.getActivityEntry.QueryRow(asid).Scan(&actorID, &targetUserID, &event, &elementType, &elementID)
	if err != nil && err != ErrNoRows {
		LogError(err)
		return
	}

	_ = WsHub.pushAlerts(uids, int(asid), event, elementType, actorID, targetUserID, elementID)
}
