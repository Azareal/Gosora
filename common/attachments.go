package common

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/Azareal/Gosora/query_gen"
)

var Attachments AttachmentStore

type MiniAttachment struct {
	ID         int
	SectionID  int
	OriginID   int
	UploadedBy int
	Path       string

	Image bool
	Ext   string
}

type AttachmentStore interface {
	Get(id int) (*MiniAttachment, error)
	MiniTopicGet(id int) (alist []*MiniAttachment, err error)
	Add(sectionID int, sectionTable string, originID int, originTable string, uploadedBy int, path string) (int, error)
	GlobalCount() int
	CountIn(originTable string, oid int) int
	CountInPath(path string) int
	Delete(aid int) error
}

type DefaultAttachmentStore struct {
	get         *sql.Stmt
	getByTopic  *sql.Stmt
	add         *sql.Stmt
	count       *sql.Stmt
	countIn     *sql.Stmt
	countInPath *sql.Stmt
	delete      *sql.Stmt
}

func NewDefaultAttachmentStore() (*DefaultAttachmentStore, error) {
	acc := qgen.NewAcc()
	return &DefaultAttachmentStore{
		get:         acc.Select("attachments").Columns("originID, sectionID, uploadedBy, path").Where("attachID = ?").Prepare(),
		getByTopic:  acc.Select("attachments").Columns("attachID, sectionID, uploadedBy, path").Where("originTable = 'topics' AND originID = ?").Prepare(),
		add:         acc.Insert("attachments").Columns("sectionID, sectionTable, originID, originTable, uploadedBy, path").Fields("?,?,?,?,?,?").Prepare(),
		count:       acc.Count("attachments").Prepare(),
		countIn:     acc.Count("attachments").Where("originTable = ? and originID = ?").Prepare(),
		countInPath: acc.Count("attachments").Where("path = ?").Prepare(),
		delete:      acc.Delete("attachments").Where("attachID = ?").Prepare(),
	}, acc.FirstError()
}

// TODO: Make this more generic so we can use it for reply attachments too
func (store *DefaultAttachmentStore) MiniTopicGet(id int) (alist []*MiniAttachment, err error) {
	rows, err := store.getByTopic.Query(id)
	defer rows.Close()
	for rows.Next() {
		attach := &MiniAttachment{OriginID: id}
		err := rows.Scan(&attach.ID, &attach.SectionID, &attach.UploadedBy, &attach.Path)
		if err != nil {
			return nil, err
		}
		extarr := strings.Split(attach.Path, ".")
		if len(extarr) < 2 {
			return nil, errors.New("corrupt attachment path")
		}
		attach.Ext = extarr[len(extarr)-1]
		attach.Image = ImageFileExts.Contains(attach.Ext)
		alist = append(alist, attach)
	}
	return alist, rows.Err()
}

func (store *DefaultAttachmentStore) Get(id int) (*MiniAttachment, error) {
	attach := &MiniAttachment{ID: id}
	err := store.get.QueryRow(id).Scan(&attach.OriginID, &attach.SectionID, &attach.UploadedBy, &attach.Path)
	if err != nil {
		return nil, err
	}
	extarr := strings.Split(attach.Path, ".")
	if len(extarr) < 2 {
		return nil, errors.New("corrupt attachment path")
	}
	attach.Ext = extarr[len(extarr)-1]
	attach.Image = ImageFileExts.Contains(attach.Ext)
	return attach, nil
}

func (store *DefaultAttachmentStore) Add(sectionID int, sectionTable string, originID int, originTable string, uploadedBy int, path string) (int, error) {
	res, err := store.add.Exec(sectionID, sectionTable, originID, originTable, uploadedBy, path)
	if err != nil {
		return 0, err
	}
	lid, err := res.LastInsertId()
	return int(lid), err
}

func (store *DefaultAttachmentStore) GlobalCount() (count int) {
	err := store.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func (store *DefaultAttachmentStore) CountIn(originTable string, oid int) (count int) {
	err := store.countIn.QueryRow(originTable, oid).Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func (store *DefaultAttachmentStore) CountInPath(path string) (count int) {
	err := store.countInPath.QueryRow(path).Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func (store *DefaultAttachmentStore) Delete(aid int) error {
	_, err := store.delete.Exec(aid)
	return err
}
