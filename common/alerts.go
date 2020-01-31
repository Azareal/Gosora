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

	//"fmt"

	"github.com/Azareal/Gosora/common/phrases"
	qgen "github.com/Azareal/Gosora/query_gen"
)

type Alert struct {
	ASID         int
	ActorID      int
	TargetUserID int
	Event        string
	ElementType  string
	ElementID    int
	CreatedAt    time.Time
	Extra        string

	Actor *User
}

type AlertStmts struct {
	notifyWatchers *sql.Stmt
	notifyOne      *sql.Stmt
	getWatchers    *sql.Stmt
}

var alertStmts AlertStmts

// TODO: Move these statements into some sort of activity abstraction
// TODO: Rewrite the alerts logic
func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		alertStmts = AlertStmts{
			notifyWatchers: acc.SimpleInsertInnerJoin(
				qgen.DBInsert{"activity_stream_matches", "watcher,asid", ""},
				qgen.DBJoin{"activity_stream", "activity_subscriptions", "activity_subscriptions.user, activity_stream.asid", "activity_subscriptions.targetType = activity_stream.elementType AND activity_subscriptions.targetID = activity_stream.elementID AND activity_subscriptions.user != activity_stream.actor", "asid=?", "", ""},
			),
			notifyOne:   acc.Insert("activity_stream_matches").Columns("watcher,asid").Fields("?,?").Prepare(),
			getWatchers: acc.SimpleInnerJoin("activity_stream", "activity_subscriptions", "activity_subscriptions.user", "activity_subscriptions.targetType = activity_stream.elementType AND activity_subscriptions.targetID = activity_stream.elementID AND activity_subscriptions.user != activity_stream.actor", "asid=?", "", ""),
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
	phraseName := "." + alert.ElementType
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
			DebugLogf("Unable to find linked topic by reply ID %d", alert.ElementID)
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

func buildAlertString(msg string, sub []string, path, avatar string, asid int) string {
	var subString string
	for _, item := range sub {
		subString += "\"" + escapeTextInJson(item) + "\","
	}
	if len(subString) > 0 {
		subString = subString[:len(subString)-1]
	}

	return `{"msg":"` + escapeTextInJson(msg) + `","sub":[` + subString + `],"path":"` + escapeTextInJson(path) + `","avatar":"` + escapeTextInJson(avatar) + `","id":` + strconv.Itoa(asid) + `}`
}

func AddActivityAndNotifyAll(a Alert) error {
	id, err := Activity.Add(a)
	if err != nil {
		return err
	}
	return NotifyWatchers(id)
}

func AddActivityAndNotifyTarget(a Alert) error {
	id, err := Activity.Add(a)
	if err != nil {
		return err
	}

	err = NotifyOne(a.TargetUserID, id)
	if err != nil {
		return err
	}
	a.ASID = id

	// Live alerts, if the target is online and WebSockets is enabled
	if EnableWebsockets {
		go func() {
			_ = WsHub.pushAlert(a.TargetUserID, a)
			//fmt.Println("err:",err)
		}()
	}
	return nil
}

func NotifyOne(watcher, asid int) error {
	_, err := alertStmts.notifyOne.Exec(watcher, asid)
	return err
}

func NotifyWatchers(asid int) error {
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

func notifyWatchers(asid int) {
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
	if err = rows.Err(); err != nil {
		LogError(err)
		return
	}

	alert, err := Activity.Get(asid)
	if err != nil && err != ErrNoRows {
		LogError(err)
		return
	}
	_ = WsHub.pushAlerts(uids, alert)
}

func DismissAlert(uid, aid int) {
	_ = WsHub.PushMessage(uid, `{"event":"dismiss-alert","id":`+strconv.Itoa(aid)+`}`)
}