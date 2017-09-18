/*
*
* Gosora Alerts System
* Copyright Azareal 2017 - 2018
*
 */
package main

import "log"
import "strings"
import "strconv"
import "errors"

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

func buildAlert(asid int, event string, elementType string, actorID int, targetUserID int, elementID int, user User /* The current user */) (string, error) {
	var targetUser *User

	actor, err := users.Get(actorID)
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
			topic, err := topics.Get(elementID)
			if err != nil {
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
		topic, err := topics.Get(elementID)
		if err != nil {
			return "", errors.New("Unable to find the linked topic")
		}
		url = topic.Link
		area = topic.Title

		if targetUserID == user.ID {
			postAct = " your topic"
		}
	case "user":
		targetUser, err = users.Get(elementID)
		if err != nil {
			return "", errors.New("Unable to find the target user")
		}
		area = targetUser.Name
		endFrag = "'s profile"
		url = targetUser.Link
	case "post":
		topic, err := getTopicByReply(elementID)
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

func notifyWatchers(asid int64) {
	rows, err := getWatchersStmt.Query(asid)
	if err != nil && err != ErrNoRows {
		log.Fatal(err.Error())
		return
	}
	defer rows.Close()

	var uid int
	var uids []int
	for rows.Next() {
		err := rows.Scan(&uid)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		uids = append(uids, uid)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	var actorID, targetUserID, elementID int
	var event, elementType string
	err = getActivityEntryStmt.QueryRow(asid).Scan(&actorID, &targetUserID, &event, &elementType, &elementID)
	if err != nil && err != ErrNoRows {
		log.Fatal(err.Error())
		return
	}

	_ = wsHub.pushAlerts(uids, int(asid), event, elementType, actorID, targetUserID, elementID)
}
