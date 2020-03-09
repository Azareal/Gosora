package common

import (
	"database/sql"
	//"log"
	"strconv"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var Recalc RecalcInt

type RecalcInt interface {
	Replies() (count int, err error)
	Forums() (count int, err error)
	Subscriptions() (count int, err error)
	ActivityStream() (count int, err error)
	Users() error
	Attachments() (count int, err error)
}

type DefaultRecalc struct {
	getActivitySubscriptions *sql.Stmt
	getActivityStream        *sql.Stmt
	getAttachments           *sql.Stmt
	getTopicCount            *sql.Stmt
	resetTopicCount          *sql.Stmt
}

func NewDefaultRecalc(acc *qgen.Accumulator) (*DefaultRecalc, error) {
	return &DefaultRecalc{
		getActivitySubscriptions: acc.Select("activity_subscriptions").Columns("targetID,targetType").Prepare(),
		getActivityStream:        acc.Select("activity_stream").Columns("asid,event,elementID,elementType,extra").Prepare(),
		getAttachments:           acc.Select("attachments").Columns("attachID,originID,originTable").Prepare(),
		getTopicCount:            acc.Count("topics").Where("parentID=?").Prepare(),
		//resetTopicCount:          acc.SimpleUpdateSelect("forums", "topicCount = tc", "topics", "count(*) as tc", "parentID=?", "", ""),
		// TODO: Avoid using RawPrepare
		resetTopicCount: acc.RawPrepare("UPDATE forums, (SELECT COUNT(*) as tc FROM topics WHERE parentID=?) AS src SET forums.topicCount=src.tc WHERE forums.fid=?"),
	}, acc.FirstError()
}

func (s *DefaultRecalc) Replies() (count int, err error) {
	var ltid int
	err = Rstore.Each(func(r *Reply) error {
		if ltid == r.ParentID && r.ParentID > 0 {
			//return nil
		}
		if !Topics.Exists(r.ParentID) {
			// TODO: Delete in chunks not one at a time?
			if err := r.Delete(); err != nil {
				return err
			}
			count++
		}
		return nil
	})
	return count, err
}

func (s *DefaultRecalc) Forums() (count int, err error) {
	err = Forums.Each(func(f *Forum) error {
		_, err := s.resetTopicCount.Exec(f.ID, f.ID)
		if err != nil {
			return err
		}
		count++
		return nil
	})
	return count, err
}

func (s *DefaultRecalc) Subscriptions() (count int, err error) {
	err = eachall(s.getActivitySubscriptions, func(r *sql.Rows) error {
		var targetID int
		var targetType string
		err := r.Scan(&targetID, &targetType)
		if err != nil {
			return err
		}
		if targetType == "topic" {
			if !Topics.Exists(targetID) {
				// TODO: Delete in chunks not one at a time?
				err := Subscriptions.DeleteResource(targetID, targetType)
				if err != nil {
					return err
				}
				count++
			}
		}
		return nil
	})
	return count, err
}

type Existable interface {
	Exists(id int) bool
}

func (s *DefaultRecalc) ActivityStream() (count int, err error) {
	err = eachall(s.getActivityStream, func(r *sql.Rows) error {
		var asid, elementID int
		var event, elementType, extra string
		err := r.Scan(&asid, &event, &elementID, &elementType, &extra)
		if err != nil {
			return err
		}
		//log.Print("asid:",asid)
		var s Existable
		switch elementType {
		case "user":
			if event == "reply" {
				extraI, _ := strconv.Atoi(extra)
				if extraI > 0 {
					s = Prstore
					elementID = extraI
				} else {
					return nil
				}
			} else {
				return nil
			}
		case "topic":
			s = Topics
			// TODO: Delete reply events with an empty extra field
			if event == "reply" {
				extraI, _ := strconv.Atoi(extra)
				if extraI > 0 {
					s = Rstore
					elementID = extraI
				}
			}
		case "post":
			s = Rstore
			// TODO: Add a TopicExistsByReplyID for efficiency
			/*_, err = TopicByReplyID(elementID)
			if err == sql.ErrNoRows {
				// TODO: Delete in chunks not one at a time?
				err := Activity.Delete(asid)
				if err != nil {
					return err
				}
				count++
			} else if err != nil {
				return err
			}*/
		default:
			return nil
		}
		if !s.Exists(elementID) {
			// TODO: Delete in chunks not one at a time?
			err := Activity.Delete(asid)
			if err != nil {
				return err
			}
			count++
		}
		return nil
	})
	return count, err
}

func (s *DefaultRecalc) Users() error {
	return Users.Each(func(u *User) error {
		return u.RecalcPostStats()
	})
}

func (s *DefaultRecalc) Attachments() (count int, err error) {
	err = eachall(s.getAttachments, func(r *sql.Rows) error {
		var aid, originID int
		var originType string
		err := r.Scan(&aid, &originID, &originType)
		if err != nil {
			return err
		}
		var s Existable
		switch originType {
		case "topics":
			s = Topics
		case "replies":
			s = Rstore
		default:
			return nil
		}
		if !s.Exists(originID) {
			// TODO: Delete in chunks not one at a time?
			err := Attachments.Delete(aid)
			if err != nil {
				return err
			}
			count++
		}
		return nil
	})
	return count, err
}
