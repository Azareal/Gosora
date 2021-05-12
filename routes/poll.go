package routes

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	c "github.com/Azareal/Gosora/common"
)

func PollVote(w http.ResponseWriter, r *http.Request, u *c.User, sPollID string) c.RouteError {
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
	_, ferr := c.SimpleForumUserCheck(w, r, u, topic.ParentID)
	if ferr != nil {
		return ferr
	}
	if !u.Perms.ViewTopic {
		return c.NoPermissions(w, r, u)
	}

	optIndex, err := strconv.Atoi(r.PostFormValue("poll_option_input"))
	if err != nil {
		return c.LocalError("Malformed input", w, r, u)
	}
	err = poll.CastVote(optIndex, u.ID, u.GetIP())
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/topic/"+strconv.Itoa(topic.ID), http.StatusSeeOther)
	return nil
}

func PollResults(w http.ResponseWriter, r *http.Request, u *c.User, sPollID string) c.RouteError {
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

	// TODO: Implement a version of this which doesn't rely so much on sequential order
	var ob bytes.Buffer
	ob.WriteRune('[')
	var i int
	e := poll.Resultsf(func(votes int) error {
		if i != 0 {
			ob.WriteRune(',')
		}
		ob.WriteString(strconv.Itoa(votes))
		i++
		return nil
	})
	if e != nil && e != sql.ErrNoRows {
		return c.InternalError(e, w, r)
	}
	ob.WriteRune(']')
	w.Write(ob.Bytes())

	return nil
}
