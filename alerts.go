package main

import "log"
import "strings"
import "errors"

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

func build_alert(event string, elementType string, actor_id int, targetUser_id int, elementID int, user User /* The current user */) (string, error) {
	var targetUser *User

	actor, err := users.CascadeGet(actor_id)
	if err != nil {
		return "", errors.New("Unable to find the actor")
	}

	/*if elementType != "forum" {
		targetUser, err = users.CascadeGet(targetUser_id)
		if err != nil {
			LocalErrorJS("Unable to find the target user",w,r)
			return
		}
	}*/

	if event == "friend_invite" {
		return `{"msg":"You received a friend invite from {0}","sub":["` + actor.Name + `"],"path":"`+actor.Link+`","avatar":"`+strings.Replace(actor.Avatar,"/","\\/",-1)+`"}`, nil
	}

	var act, post_act, url, area string
	var start_frag, end_frag string
	switch(elementType) {
		case "forum":
			if event == "reply" {
				act = "created a new topic"
				topic, err := topics.CascadeGet(elementID)
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
			topic, err := topics.CascadeGet(elementID)
			if err != nil {
				return "", errors.New("Unable to find the linked topic")
			}
			url = topic.Link
			area = topic.Title

			if targetUser_id == user.ID {
				post_act = " your topic"
			}
		case "user":
			targetUser, err = users.CascadeGet(elementID)
			if err != nil {
				return "", errors.New("Unable to find the target user")
			}
			area = targetUser.Name
			end_frag = "'s profile"
			url = targetUser.Link
		case "post":
			topic, err := get_topic_by_reply(elementID)
			if err != nil {
				return "", errors.New("Unable to find the linked reply or parent topic")
			}
			url = topic.Link
			area = topic.Title
			if targetUser_id == user.ID {
				post_act = " your post in"
			}
		default:
			return "", errors.New("Invalid elementType")
	}

	switch(event) {
		case "like":
			if elementType == "user" {
				act = "likes"
				end_frag = ""
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
				post_act = ""
			}
		case "reply": act = "replied to"
	}

	return `{"msg":"{0} ` + start_frag + act + post_act + ` {1}` + end_frag + `","sub":["` + actor.Name + `","` + area + `"],"path":"` + url + `","avatar":"` + actor.Avatar + `"}`, nil
}

func notify_watchers(asid int64) {
	rows, err := get_watchers_stmt.Query(asid)
	if err != nil && err != ErrNoRows {
		log.Fatal(err.Error())
		return
	}

	var uid int
	var uids []int
	for rows.Next() {
		err := rows.Scan(&uid)
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		uids = append(uids,uid)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	rows.Close()

	var actor_id, targetUser_id, elementID int
	var event, elementType string
	err = get_activity_entry_stmt.QueryRow(asid).Scan(&actor_id, &targetUser_id, &event, &elementType, &elementID)
	if err != nil && err != ErrNoRows {
		log.Fatal(err.Error())
		return
	}

	_ = ws_hub.push_alerts(uids, event, elementType, actor_id, targetUser_id, elementID)
}
