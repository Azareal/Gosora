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
	Extra      string

	Image bool
	Ext   string
}

type AttachmentStore interface {
	Get(id int) (*MiniAttachment, error)
	MiniGetList(originTable string, originID int) (alist []*MiniAttachment, err error)
	BulkMiniGetList(originTable string, ids []int) (amap map[int][]*MiniAttachment, err error)
	Add(sectionID int, sectionTable string, originID int, originTable string, uploadedBy int, path string, extra string) (int, error)
	MoveTo(sectionID int, originID int, originTable string) error
	MoveToByExtra(sectionID int, originTable string, extra string) error
	GlobalCount() int
	CountIn(originTable string, oid int) int
	CountInPath(path string) int
	Delete(aid int) error
}

type DefaultAttachmentStore struct {
	get         *sql.Stmt
	getByObj    *sql.Stmt
	add         *sql.Stmt
	count       *sql.Stmt
	countIn     *sql.Stmt
	countInPath *sql.Stmt
	move        *sql.Stmt
	moveByExtra *sql.Stmt
	delete      *sql.Stmt
}

func NewDefaultAttachmentStore(acc *qgen.Accumulator) (*DefaultAttachmentStore, error) {
	return &DefaultAttachmentStore{
		get:         acc.Select("attachments").Columns("originID, sectionID, uploadedBy, path, extra").Where("attachID = ?").Prepare(),
		getByObj:    acc.Select("attachments").Columns("attachID, sectionID, uploadedBy, path, extra").Where("originTable = ? AND originID = ?").Prepare(),
		add:         acc.Insert("attachments").Columns("sectionID, sectionTable, originID, originTable, uploadedBy, path, extra").Fields("?,?,?,?,?,?,?").Prepare(),
		count:       acc.Count("attachments").Prepare(),
		countIn:     acc.Count("attachments").Where("originTable = ? and originID = ?").Prepare(),
		countInPath: acc.Count("attachments").Where("path = ?").Prepare(),
		move:        acc.Update("attachments").Set("sectionID = ?").Where("originID = ? AND originTable = ?").Prepare(),
		moveByExtra: acc.Update("attachments").Set("sectionID = ?").Where("originTable = ? AND extra = ?").Prepare(),
		delete:      acc.Delete("attachments").Where("attachID = ?").Prepare(),
	}, acc.FirstError()
}

func (store *DefaultAttachmentStore) MiniGetList(originTable string, originID int) (alist []*MiniAttachment, err error) {
	rows, err := store.getByObj.Query(originTable, originID)
	defer rows.Close()
	for rows.Next() {
		attach := &MiniAttachment{OriginID: originID}
		err := rows.Scan(&attach.ID, &attach.SectionID, &attach.UploadedBy, &attach.Path, &attach.Extra)
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

func (store *DefaultAttachmentStore) BulkMiniGetList(originTable string, ids []int) (amap map[int][]*MiniAttachment, err error) {
	if len(ids) == 0 {
		return nil, sql.ErrNoRows
	}
	if len(ids) == 1 {
		res, err := store.MiniGetList(originTable, ids[0])
		return map[int][]*MiniAttachment{ids[0]: res}, err
	}

	amap = make(map[int][]*MiniAttachment)
	var buffer []*MiniAttachment
	var currentID int
	rows, err := qgen.NewAcc().Select("attachments").Columns("attachID, sectionID, originID, uploadedBy, path").Where("originTable = ?").In("originID", ids).Orderby("originID ASC").Query(originTable)
	defer rows.Close()
	for rows.Next() {
		attach := &MiniAttachment{}
		err := rows.Scan(&attach.ID, &attach.SectionID, &attach.OriginID, &attach.UploadedBy, &attach.Path)
		if err != nil {
			return nil, err
		}
		extarr := strings.Split(attach.Path, ".")
		if len(extarr) < 2 {
			return nil, errors.New("corrupt attachment path")
		}
		attach.Ext = extarr[len(extarr)-1]
		attach.Image = ImageFileExts.Contains(attach.Ext)
		if attach.ID != currentID {
			if len(buffer) > 0 {
				amap[currentID] = buffer
				buffer = nil
			}
		}
		buffer = append(buffer, attach)
	}
	return amap, rows.Err()
}

func (store *DefaultAttachmentStore) Get(id int) (*MiniAttachment, error) {
	attach := &MiniAttachment{ID: id}
	err := store.get.QueryRow(id).Scan(&attach.OriginID, &attach.SectionID, &attach.UploadedBy, &attach.Path, &attach.Extra)
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

func (store *DefaultAttachmentStore) Add(sectionID int, sectionTable string, originID int, originTable string, uploadedBy int, path string, extra string) (int, error) {
	res, err := store.add.Exec(sectionID, sectionTable, originID, originTable, uploadedBy, path, extra)
	if err != nil {
		return 0, err
	}
	lid, err := res.LastInsertId()
	return int(lid), err
}

func (store *DefaultAttachmentStore) MoveTo(sectionID int, originID int, originTable string) error {
	_, err := store.move.Exec(sectionID, originID, originTable)
	return err
}

func (store *DefaultAttachmentStore) MoveToByExtra(sectionID int, originTable string, extra string) error {
	_, err := store.moveByExtra.Exec(sectionID, originTable, extra)
	return err
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
