/*
*
* Gosora Alerts System
* Copyright Azareal 2017 - 2020
*
 */
package common

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/Azareal/Gosora/common/phrases"
	"github.com/Azareal/Gosora/query_gen"
)

type Alert struct {
	ASID         int
	ActorID      int
	TargetUserID int
	Event        string
	ElementType  string
	ElementID    int
	CreatedAt time.Time

	Actor *User
}

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
			addActivity: acc.Insert("activity_stream").Columns("actor, targetUser, event, elementType, elementID, createdAt").Fields("?,?,?,?,?,UTC_TIMESTAMP()").Prepare(),
			notifyWatchers: acc.SimpleInsertInnerJoin(
				qgen.DBInsert{"activity_stream_matches", "watcher, asid", ""},
				qgen.DBJoin{"activity_stream", "activity_subscriptions", "activity_subscriptions.user, activity_stream.asid", "activity_subscriptions.targetType = activity_stream.elementType AND activity_subscriptions.targetID = activity_stream.elementID AND activity_subscriptions.user != activity_stream.actor", "asid = ?", "", ""},
			),
			notifyOne:        acc.Insert("activity_stream_matches").Columns("watcher, asid").Fields("?,?").Prepare(),
			getWatchers:      acc.SimpleInnerJoin("activity_stream", "activity_subscriptions", "activity_subscriptions.user", "activity_subscriptions.targetType = activity_stream.elementType AND activity_subscriptions.targetID = activity_stream.elementID AND activity_subscriptions.user != activity_stream.actor", "asid = ?", "", ""),
			getActivityEntry: acc.Select("activity_stream").Columns("actor, targetUser, event, elementType, elementID, createdAt").Where("asid = ?").Prepare(),
		}
		return acc.FirstError()
	})
}

// TODO: See if we can json.Marshal instead?
func escapeTextInJson(in string) string {
	in = strings.Replace(in, "\"", "\\\"", -1)
	return strings.Replace(in, "/", "\\/", -1)
}

func BuildAlert(alert Alert, user User /* The current user */) (out string, err error) {
	var targetUser *User
	if alert.Actor == nil {
		alert.Actor, err = Users.Get(alert.ActorID)
		if err != nil {
			return "", errors.New(phrases.GetErrorPhrase("alerts_no_actor"))
		}
	}

	/*if alert.ElementType != "forum" {
		targetUser, err = users.Get(alert.TargetUserID)
		if err != nil {
			LocalErrorJS("Unable to find the target user",w,r)
			return
		}
	}*/

	if alert.Event == "friend_invite" {
		return buildAlertString(".new_friend_invite", []string{alert.Actor.Name}, alert.Actor.Link, alert.Actor.Avatar, alert.ASID), nil
	}

	// Not that many events for us to handle in a forum
	if alert.ElementType == "forum" {
		if alert.Event == "reply" {
			topic, err := Topics.Get(alert.ElementID)
			if err != nil {
				DebugLogf("Unable to find linked topic %d", alert.ElementID)
				return "", errors.New(phrases.GetErrorPhrase("alerts_no_linked_topic"))
			}
			// Store the forum ID in the targetUser column instead of making a new one? o.O
			// Add an additional column for extra information later on when we add the ability to link directly to posts. We don't need the forum data for now...
			return buildAlertString(".forum_new_topic", []string{alert.Actor.Name, topic.Title}, topic.Link, alert.Actor.Avatar, alert.ASID), nil
		}
		return buildAlertString(".forum_unknown_action", []string{alert.Actor.Name}, "", alert.Actor.Avatar, alert.ASID), nil
	}

	var url, area string
	var phraseName = "." + alert.ElementType
	switch alert.ElementType {
	case "topic":
		topic, err := Topics.Get(alert.ElementID)
		if err != nil {
			DebugLogf("Unable to find linked topic %d", alert.ElementID)
			return "", errors.New(phrases.GetErrorPhrase("alerts_no_linked_topic"))
		}
		url = topic.Link
		area = topic.Title
		if alert.TargetUserID == user.ID {
			phraseName += "_own"
		}
	case "user":
		targetUser, err = Users.Get(alert.ElementID)
		if err != nil {
			DebugLogf("Unable to find target user %d", alert.ElementID)
			return "", errors.New(phrases.GetErrorPhrase("alerts_no_target_user"))
		}
		area = targetUser.Name
		url = targetUser.Link
		if alert.TargetUserID == user.ID {
			phraseName += "_own"
		}
	case "post":
		topic, err := TopicByReplyID(alert.ElementID)
		if err != nil {
			return "", errors.New(phrases.GetErrorPhrase("alerts_no_linked_topic_by_reply"))
		}
		url = topic.Link
		area = topic.Title
		if alert.TargetUserID == user.ID {
			phraseName += "_own"
		}
	default:
		return "", errors.New(phrases.GetErrorPhrase("alerts_invalid_elementtype"))
	}

	switch alert.Event {
	case "like":
		phraseName += "_like"
	case "mention":
		phraseName += "_mention"
	case "reply":
		phraseName += "_reply"
	}

	return buildAlertString(phraseName, []string{alert.Actor.Name, area}, url, alert.Actor.Avatar, alert.ASID), nil
}

func buildAlertString(msg string, sub []string, path string, avatar string, asid int) string {
	var substring string
	for _, item := range sub {
		substring += "\"" + escapeTextInJson(item) + "\","
	}
	if len(substring) > 0 {
		substring = substring[:len(substring)-1]
	}

	return `{"msg":"` + escapeTextInJson(msg) + `","sub":[` + substring + `],"path":"` + escapeTextInJson(path) + `","avatar":"` + escapeTextInJson(avatar) + `","id":` + strconv.Itoa(asid) + `}`
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

func AddActivityAndNotifyTarget(alert Alert) error {
	res, err := alertStmts.addActivity.Exec(alert.ActorID, alert.TargetUserID, alert.Event, alert.ElementType, alert.ElementID)
	if err != nil {
		return err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	err = NotifyOne(alert.TargetUserID, lastID)
	if err != nil {
		return err
	}
	alert.ASID = int(lastID)

	// Live alerts, if the target is online and WebSockets is enabled
	_ = WsHub.pushAlert(alert.TargetUserID, alert)
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

	var alert = Alert{ASID: int(asid)}
	err = alertStmts.getActivityEntry.QueryRow(asid).Scan(&alert.ActorID, &alert.TargetUserID, &alert.Event, &alert.ElementType, &alert.ElementID, &alert.CreatedAt)
	if err != nil && err != ErrNoRows {
		LogError(err)
		return
	}

	_ = WsHub.pushAlerts(uids, alert)
}
