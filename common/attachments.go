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
	Count() int
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
	a := "attachments"
	return &DefaultAttachmentStore{
		get:         acc.Select(a).Columns("originID, sectionID, uploadedBy, path, extra").Where("attachID = ?").Prepare(),
		getByObj:    acc.Select(a).Columns("attachID, sectionID, uploadedBy, path, extra").Where("originTable = ? AND originID = ?").Prepare(),
		add:         acc.Insert(a).Columns("sectionID, sectionTable, originID, originTable, uploadedBy, path, extra").Fields("?,?,?,?,?,?,?").Prepare(),
		count:       acc.Count(a).Prepare(),
		countIn:     acc.Count(a).Where("originTable = ? and originID = ?").Prepare(),
		countInPath: acc.Count(a).Where("path = ?").Prepare(),
		move:        acc.Update(a).Set("sectionID = ?").Where("originID = ? AND originTable = ?").Prepare(),
		moveByExtra: acc.Update(a).Set("sectionID = ?").Where("originTable = ? AND extra = ?").Prepare(),
		delete:      acc.Delete(a).Where("attachID = ?").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultAttachmentStore) MiniGetList(originTable string, originID int) (alist []*MiniAttachment, err error) {
	rows, err := s.getByObj.Query(originTable, originID)
	defer rows.Close()
	for rows.Next() {
		a := &MiniAttachment{OriginID: originID}
		err := rows.Scan(&a.ID, &a.SectionID, &a.UploadedBy, &a.Path, &a.Extra)
		if err != nil {
			return nil, err
		}
		extarr := strings.Split(a.Path, ".")
		if len(extarr) < 2 {
			return nil, errors.New("corrupt attachment path")
		}
		a.Ext = extarr[len(extarr)-1]
		a.Image = ImageFileExts.Contains(a.Ext)
		alist = append(alist, a)
	}
	return alist, rows.Err()
}

func (s *DefaultAttachmentStore) BulkMiniGetList(originTable string, ids []int) (amap map[int][]*MiniAttachment, err error) {
	if len(ids) == 0 {
		return nil, sql.ErrNoRows
	}
	if len(ids) == 1 {
		res, err := s.MiniGetList(originTable, ids[0])
		return map[int][]*MiniAttachment{ids[0]: res}, err
	}

	amap = make(map[int][]*MiniAttachment)
	var buffer []*MiniAttachment
	var currentID int
	rows, err := qgen.NewAcc().Select("attachments").Columns("attachID,sectionID,originID,uploadedBy,path").Where("originTable=?").In("originID", ids).Orderby("originID ASC").Query(originTable)
	defer rows.Close()
	for rows.Next() {
		a := &MiniAttachment{}
		err := rows.Scan(&a.ID, &a.SectionID, &a.OriginID, &a.UploadedBy, &a.Path)
		if err != nil {
			return nil, err
		}
		extarr := strings.Split(a.Path, ".")
		if len(extarr) < 2 {
			return nil, errors.New("corrupt attachment path")
		}
		a.Ext = extarr[len(extarr)-1]
		a.Image = ImageFileExts.Contains(a.Ext)
		if currentID == 0 {
			currentID = a.OriginID
		}
		if a.OriginID != currentID {
			if len(buffer) > 0 {
				amap[currentID] = buffer
				currentID = a.OriginID
				buffer = nil
			}
		}
		buffer = append(buffer, a)
	}
	if len(buffer) > 0 {
		amap[currentID] = buffer
	}
	return amap, rows.Err()
}

func (s *DefaultAttachmentStore) Get(id int) (*MiniAttachment, error) {
	a := &MiniAttachment{ID: id}
	err := s.get.QueryRow(id).Scan(&a.OriginID, &a.SectionID, &a.UploadedBy, &a.Path, &a.Extra)
	if err != nil {
		return nil, err
	}
	extarr := strings.Split(a.Path, ".")
	if len(extarr) < 2 {
		return nil, errors.New("corrupt attachment path")
	}
	a.Ext = extarr[len(extarr)-1]
	a.Image = ImageFileExts.Contains(a.Ext)
	return a, nil
}

func (s *DefaultAttachmentStore) Add(sectionID int, sectionTable string, originID int, originTable string, uploadedBy int, path string, extra string) (int, error) {
	res, err := s.add.Exec(sectionID, sectionTable, originID, originTable, uploadedBy, path, extra)
	if err != nil {
		return 0, err
	}
	lid, err := res.LastInsertId()
	return int(lid), err
}

func (s *DefaultAttachmentStore) MoveTo(sectionID int, originID int, originTable string) error {
	_, err := s.move.Exec(sectionID, originID, originTable)
	return err
}

func (s *DefaultAttachmentStore) MoveToByExtra(sectionID int, originTable string, extra string) error {
	_, err := s.moveByExtra.Exec(sectionID, originTable, extra)
	return err
}

func (s *DefaultAttachmentStore) Count() (count int) {
	err := s.count.QueryRow().Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func (s *DefaultAttachmentStore) CountIn(originTable string, oid int) (count int) {
	err := s.countIn.QueryRow(originTable, oid).Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func (s *DefaultAttachmentStore) CountInPath(path string) (count int) {
	err := s.countInPath.QueryRow(path).Scan(&count)
	if err != nil {
		LogError(err)
	}
	return count
}

func (s *DefaultAttachmentStore) Delete(aid int) error {
	_, err := s.delete.Exec(aid)
	return err
}
