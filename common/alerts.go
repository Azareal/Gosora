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

const AlertsGrowHint = len(`{"msgs":[],"count":,"tc":}`) + 1 + 10

// TODO: See if we can json.Marshal instead?
func escapeTextInJson(in string) string {
	in = strings.Replace(in, "\"", "\\\"", -1)
	return strings.Replace(in, "/", "\\/", -1)
}

func BuildAlert(a Alert, user User /* The current user */) (out string, err error) {
	var targetUser *User
	if a.Actor == nil {
		a.Actor, err = Users.Get(a.ActorID)
		if err != nil {
			return "", errors.New(phrases.GetErrorPhrase("alerts_no_actor"))
		}
	}

	/*if a.ElementType != "forum" {
		targetUser, err = users.Get(a.TargetUserID)
		if err != nil {
			LocalErrorJS("Unable to find the target user",w,r)
			return
		}
	}*/
	if a.Event == "friend_invite" {
		return buildAlertString(".new_friend_invite", []string{a.Actor.Name}, a.Actor.Link, a.Actor.Avatar, a.ASID), nil
	}

	// Not that many events for us to handle in a forum
	if a.ElementType == "forum" {
		if a.Event == "reply" {
			topic, err := Topics.Get(a.ElementID)
			if err != nil {
				DebugLogf("Unable to find linked topic %d", a.ElementID)
				return "", errors.New(phrases.GetErrorPhrase("alerts_no_linked_topic"))
			}
			// Store the forum ID in the targetUser column instead of making a new one? o.O
			// Add an additional column for extra information later on when we add the ability to link directly to posts. We don't need the forum data for now...
			return buildAlertString(".forum_new_topic", []string{a.Actor.Name, topic.Title}, topic.Link, a.Actor.Avatar, a.ASID), nil
		}
		return buildAlertString(".forum_unknown_action", []string{a.Actor.Name}, "", a.Actor.Avatar, a.ASID), nil
	}

	var url, area, phraseName string
	own := false
	switch a.ElementType {
	case "convo":
		convo, err := Convos.Get(a.ElementID)
		if err != nil {
			DebugLogf("Unable to find linked convo %d", a.ElementID)
			return "", errors.New(phrases.GetErrorPhrase("alerts_no_linked_convo"))
		}
		url = convo.Link
	case "topic":
		topic, err := Topics.Get(a.ElementID)
		if err != nil {
			DebugLogf("Unable to find linked topic %d", a.ElementID)
			return "", errors.New(phrases.GetErrorPhrase("alerts_no_linked_topic"))
		}
		url = topic.Link
		area = topic.Title
		own = a.TargetUserID == user.ID
	case "user":
		targetUser, err = Users.Get(a.ElementID)
		if err != nil {
			DebugLogf("Unable to find target user %d", a.ElementID)
			return "", errors.New(phrases.GetErrorPhrase("alerts_no_target_user"))
		}
		area = targetUser.Name
		url = targetUser.Link
		own = a.TargetUserID == user.ID
	case "post":
		topic, err := TopicByReplyID(a.ElementID)
		if err != nil {
			DebugLogf("Unable to find linked topic by reply ID %d", a.ElementID)
			return "", errors.New(phrases.GetErrorPhrase("alerts_no_linked_topic_by_reply"))
		}
		url = topic.Link
		area = topic.Title
		own = a.TargetUserID == user.ID
	default:
		return "", errors.New(phrases.GetErrorPhrase("alerts_invalid_elementtype"))
	}

	badEv := false
	switch a.Event {
	case "create", "like", "mention", "reply":
		// skip
	default:
		badEv = true
	}

	if own && !badEv {
		phraseName = "." + a.ElementType + "_own_" + a.Event
	} else if !badEv {
		phraseName = "." + a.ElementType + "_" + a.Event
	} else if own {
		phraseName = "." + a.ElementType + "_own"
	} else {
		phraseName = "." + a.ElementType
	}

	return buildAlertString(phraseName, []string{a.Actor.Name, area}, url, a.Actor.Avatar, a.ASID), nil
}

func buildAlertString(msg string, sub []string, path, avatar string, asid int) string {
	var sb strings.Builder
	buildAlertSb(&sb, msg, sub, path, avatar, asid)
	return sb.String()
}

const AlertsGrowHint2 = len(`{"msg":"","sub":[],"path":"","avatar":"","id":}`) + 5 + 3 + 1 + 1 + 1

// TODO: Use a string builder?
func buildAlertSb(sb *strings.Builder, msg string, sub []string, path, avatar string, asid int) {
	sb.WriteString(`{"msg":"`)
	sb.WriteString(escapeTextInJson(msg))
	sb.WriteString(`","sub":[`)
	for i, it := range sub {
		if i != 0 {
			sb.WriteString(",\"")
		} else {
			sb.WriteString("\"")
		}
		sb.WriteString(escapeTextInJson(it))
		sb.WriteString("\"")
	}
	sb.WriteString(`],"path":"`)
	sb.WriteString(escapeTextInJson(path))
	sb.WriteString(`","avatar":"`)
	sb.WriteString(escapeTextInJson(avatar))
	sb.WriteString(`","id":`)
	sb.WriteString(strconv.Itoa(asid))
	sb.WriteRune('}')
}

func BuildAlertSb(sb *strings.Builder, a *Alert, user User /* The current user */) (err error) {
	var targetUser *User
	if a.Actor == nil {
		a.Actor, err = Users.Get(a.ActorID)
		if err != nil {
			return errors.New(phrases.GetErrorPhrase("alerts_no_actor"))
		}
	}

	/*if a.ElementType != "forum" {
		targetUser, err = users.Get(a.TargetUserID)
		if err != nil {
			LocalErrorJS("Unable to find the target user",w,r)
			return
		}
	}*/
	if a.Event == "friend_invite" {
		buildAlertSb(sb, ".new_friend_invite", []string{a.Actor.Name}, a.Actor.Link, a.Actor.Avatar, a.ASID)
		return nil
	}

	// Not that many events for us to handle in a forum
	if a.ElementType == "forum" {
		if a.Event == "reply" {
			topic, err := Topics.Get(a.ElementID)
			if err != nil {
				DebugLogf("Unable to find linked topic %d", a.ElementID)
				return errors.New(phrases.GetErrorPhrase("alerts_no_linked_topic"))
			}
			// Store the forum ID in the targetUser column instead of making a new one? o.O
			// Add an additional column for extra information later on when we add the ability to link directly to posts. We don't need the forum data for now...
			buildAlertSb(sb, ".forum_new_topic", []string{a.Actor.Name, topic.Title}, topic.Link, a.Actor.Avatar, a.ASID)
			return nil
		}
		buildAlertSb(sb, ".forum_unknown_action", []string{a.Actor.Name}, "", a.Actor.Avatar, a.ASID)
		return nil
	}

	var url, area string
	own := false
	switch a.ElementType {
	case "convo":
		convo, err := Convos.Get(a.ElementID)
		if err != nil {
			DebugLogf("Unable to find linked convo %d", a.ElementID)
			return errors.New(phrases.GetErrorPhrase("alerts_no_linked_convo"))
		}
		url = convo.Link
	case "topic":
		topic, err := Topics.Get(a.ElementID)
		if err != nil {
			DebugLogf("Unable to find linked topic %d", a.ElementID)
			return errors.New(phrases.GetErrorPhrase("alerts_no_linked_topic"))
		}
		url = topic.Link
		area = topic.Title
		own = a.TargetUserID == user.ID
	case "user":
		targetUser, err = Users.Get(a.ElementID)
		if err != nil {
			DebugLogf("Unable to find target user %d", a.ElementID)
			return errors.New(phrases.GetErrorPhrase("alerts_no_target_user"))
		}
		area = targetUser.Name
		url = targetUser.Link
		own = a.TargetUserID == user.ID
	case "post":
		topic, err := TopicByReplyID(a.ElementID)
		if err != nil {
			DebugLogf("Unable to find linked topic by reply ID %d", a.ElementID)
			return errors.New(phrases.GetErrorPhrase("alerts_no_linked_topic_by_reply"))
		}
		url = topic.Link
		area = topic.Title
		own = a.TargetUserID == user.ID
	default:
		return errors.New(phrases.GetErrorPhrase("alerts_invalid_elementtype"))
	}

	sb.WriteString(`{"msg":".`)
	sb.WriteString(a.ElementType)
	if own {
		sb.WriteString("_own_")
	} else {
		sb.WriteRune('_')
	}
	switch a.Event {
	case "create", "like", "mention", "reply":
		sb.WriteString(a.Event)
	}

	sb.WriteString(`","sub":["`)
	sb.WriteString(escapeTextInJson(a.Actor.Name))
	sb.WriteString("\",\"")
	sb.WriteString(escapeTextInJson(area))
	sb.WriteString(`"],"path":"`)
	sb.WriteString(escapeTextInJson(url))
	sb.WriteString(`","avatar":"`)
	sb.WriteString(escapeTextInJson(a.Actor.Avatar))
	sb.WriteString(`","id":`)
	sb.WriteString(strconv.Itoa(a.ASID))
	sb.WriteRune('}')

	return nil
}

//var AlertsGrowHint3 = len(`{"msg":"._","sub":["",""],"path":"","avatar":"","id":}`) + 3 + 2 + 2 + 2 + 2 + 1

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
