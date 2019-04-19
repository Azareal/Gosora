package routes

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/query_gen"
)

func PollVote(w http.ResponseWriter, r *http.Request, user c.User, sPollID string) c.RouteError {
	pollID, err := strconv.Atoi(sPollID)
	if err != nil {
		return c.PreError("The provided PollID is not a valid number.", w, r)
	}

	poll, err := c.Polls.Get(pollID)
	if err == sql.ErrNoRows {
		return c.PreError("The poll you tried to vote for doesn't exist.", w, r)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	var topic *c.Topic
	if poll.ParentTable == "replies" {
		reply, err := c.Rstore.Get(poll.ParentID)
		if err == sql.ErrNoRows {
			return c.PreError("The parent post doesn't exist.", w, r)
		} else if err != nil {
			return c.InternalError(err, w, r)
		}
		topic, err = c.Topics.Get(reply.ParentID)
	} else if poll.ParentTable == "topics" {
		topic, err = c.Topics.Get(poll.ParentID)
	} else {
		return c.InternalError(errors.New("Unknown parentTable for poll"), w, r)
	}

	if err == sql.ErrNoRows {
		return c.PreError("The parent topic doesn't exist.", w, r)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	// TODO: Add hooks to make use of headerLite
	_, ferr := c.SimpleForumUserCheck(w, r, &user, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ViewTopic {
		return c.NoPermissions(w, r, user)
	}

	optionIndex, err := strconv.Atoi(r.PostFormValue("poll_option_input"))
	if err != nil {
		return c.LocalError("Malformed input", w, r, user)
	}

	err = poll.CastVote(optionIndex, user.ID, user.LastIP)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/topic/"+strconv.Itoa(topic.ID), http.StatusSeeOther)
	return nil
}

func PollResults(w http.ResponseWriter, r *http.Request, user c.User, sPollID string) c.RouteError {
	//log.Print("in PollResults")
	pollID, err := strconv.Atoi(sPollID)
	if err != nil {
		return c.PreError("The provided PollID is not a valid number.", w, r)
	}

	poll, err := c.Polls.Get(pollID)
	if err == sql.ErrNoRows {
		return c.PreError("The poll you tried to vote for doesn't exist.", w, r)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	// TODO: Abstract this
	rows, err := qgen.NewAcc().Select("polls_options").Columns("votes").Where("pollID = ?").Orderby("option ASC").Query(poll.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	defer rows.Close()

	var optionList = ""
	for rows.Next() {
		var votes int
		err := rows.Scan(&votes)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		optionList += strconv.Itoa(votes) + ","
	}
	err = rows.Err()
	if err != nil {
		return c.InternalError(err, w, r)
	}

	// TODO: Implement a version of this which doesn't rely so much on sequential order
	if len(optionList) > 0 {
		optionList = optionList[:len(optionList)-1]
	}
	w.Write([]byte("[" + optionList + "]"))
	return nil
}
