/* Work in progress. Check back later! */
package main

import "log"
import "errors"
import "database/sql"
import "./query_gen/lib"

var err_noforum = errors.New("This forum doesn't exist")

type ForumStore interface
{
	Get(int) (*Forum, error)
	CascadeGet(int) (*Forum, error)
	BypassGet(int) (*Forum, error)
	//Update(Forum) error
	//CascadeUpdate(Forum) error
	//Delete(int) error
	//CascadeDelete(int) error
	//QuickCreate(string, string, bool, string) (*Forum, error)
	Exists(int) bool
}

type StaticForumStore struct
{
	get *sql.Stmt
	get_all *sql.Stmt
}

func NewStaticForumStore() *StaticForumStore {
	get_stmt, err := qgen.Builder.SimpleSelect("forums","name, desc, active, preset, topicCount, lastTopic, lastTopicID, lastReplyer, lastReplyerID, lastTopicTime","fid = ?","","")
	if err != nil {
		log.Fatal(err)
	}
	get_all_stmt, err := qgen.Builder.SimpleSelect("forums","fid, name, desc, active, preset, topicCount, lastTopic, lastTopicID, lastReplyer, lastReplyerID, lastTopicTime","","fid ASC","")
	if err != nil {
		log.Fatal(err)
	}
	return &StaticForumStore{
		get: get_stmt,
		get_all: get_all_stmt,
	}
}

func (sfs *StaticForumStore) Get(id int) (*Forum, error) {
	if !((id <= forumCapCount) && (id >= 0) && forums[id].Name!="") {
		return nil, err_noforum
	}
	return &forums[id], nil
}

func (sfs *StaticForumStore) CascadeGet(id int) (*Forum, error) {
	if !((id <= forumCapCount) && (id >= 0) && forums[id].Name!="") {
		return nil, err_noforum
	}
	return &forums[id], nil
}

func (sfs *StaticForumStore) BypassGet(id int) (*Forum, error) {
	var forum Forum = Forum{ID:id}
	err := sfs.get.QueryRow(id).Scan(&forum.Name, &forum.Desc, &forum.Active, &forum.Preset, &forum.TopicCount, &forum.LastTopic, &forum.LastTopicID, &forum.LastReplyer, &forum.LastReplyerID, &forum.LastTopicTime)
	if err != nil {
		return nil, err
	}
	return &forum, nil
}

func (sfs *StaticForumStore) Exists(id int) bool {
	return (id <= forumCapCount) && (id >= 0) && forums[id].Name != ""
}
