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

// These notes are for me, don't worry about it too much ^_^
/*
"You received a friend invite from {user}"
"{x}{mentioned you on}{user}{'s profile}"
"{x}{mentioned you in}{topic}"
"{x}{likes}{you}"
"{x}{liked}{your topic}{topic}"
"{x}{liked}{your post on}{user}{'s profile}" todo
"{x}{liked}{your post in}{topic}"
"{x}{replied to}{your post in}{topic}" todo
"{x}{replied to}{topic}"
"{x}{replied to}{your topic}{topic}"
"{x}{created a new topic}{topic}"
*/

func BuildAlert(asid int, event string, elementType string, actorID int, targetUserID int, elementID int, user User /* The current user */) (string, error) {
	var targetUser *User

	actor, err := Users.Get(actorID)
	if err != nil {
		return "", errors.New("Unable to find the actor")
	}

	/*if elementType != "forum" {
		targetUser, err = users.Get(targetUser_id)
		if err != nil {
			LocalErrorJS("Unable to find the target user",w,r)
			return
		}
	}*/

	if event == "friend_invite" {
		return `{"msg":"You received a friend invite from {0}","sub":["` + actor.Name + `"],"path":"` + actor.Link + `","avatar":"` + strings.Replace(actor.Avatar, "/", "\\/", -1) + `","asid":"` + strconv.Itoa(asid) + `"}`, nil
	}

	var act, postAct, url, area string
	var startFrag, endFrag string
	switch elementType {
	case "forum":
		if event == "reply" {
			act = "created a new topic"
			topic, err := Topics.Get(elementID)
			if err != nil {
				DebugLogf("Unable to find linked topic %d", elementID)
				return "", errors.New("Unable to find the linked topic")
			}
			url = topic.Link
			area = topic.Title
			// Store the forum ID in the targetUser column instead of making a new one? o.O
			// Add an additional column for extra information later on when we add the ability to link directly to posts. We don't need the forum data for now...
		} else {
			act = "did something in a forum"
		}
	case "topic":
		topic, err := Topics.Get(elementID)
		if err != nil {
			DebugLogf("Unable to find linked topic %d", elementID)
			return "", errors.New("Unable to find the linked topic")
		}
		url = topic.Link
		area = topic.Title

		if targetUserID == user.ID {
			postAct = " your topic"
		}
	case "user":
		targetUser, err = Users.Get(elementID)
		if err != nil {
			DebugLogf("Unable to find target user %d", elementID)
			return "", errors.New("Unable to find the target user")
		}
		area = targetUser.Name
		endFrag = "'s profile"
		url = targetUser.Link
	case "post":
		topic, err := TopicByReplyID(elementID)
		if err != nil {
			return "", errors.New("Unable to find the linked reply or parent topic")
		}
		url = topic.Link
		area = topic.Title
		if targetUserID == user.ID {
			postAct = " your post in"
		}
	default:
		return "", errors.New("Invalid elementType")
	}

	switch event {
	case "like":
		if elementType == "user" {
			act = "likes"
			endFrag = ""
			if targetUser.ID == user.ID {
				area = "you"
			}
		} else {
			act = "liked"
		}
	case "mention":
		if elementType == "user" {
			act = "mentioned you on"
		} else {
			act = "mentioned you in"
			postAct = ""
		}
	case "reply":
		act = "replied to"
	}

	return `{"msg":"{0} ` + startFrag + act + postAct + ` {1}` + endFrag + `","sub":["` + actor.Name + `","` + area + `"],"path":"` + url + `","avatar":"` + actor.Avatar + `","asid":"` + strconv.Itoa(asid) + `"}`, nil
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
