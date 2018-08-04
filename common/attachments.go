package common

import (
	"database/sql"

	"../query_gen/lib"
)

var Attachments AttachmentStore

type AttachmentStore interface {
	Add(sectionID int, sectionTable string, originID int, originTable string, uploadedBy int, path string) error
}

type DefaultAttachmentStore struct {
	add *sql.Stmt
}

func NewDefaultAttachmentStore() (*DefaultAttachmentStore, error) {
	acc := qgen.NewAcc()
	return &DefaultAttachmentStore{
		add: acc.Insert("attachments").Columns("sectionID, sectionTable, originID, originTable, uploadedBy, path").Fields("?,?,?,?,?,?").Prepare(),
	}, acc.FirstError()
}

func (store *DefaultAttachmentStore) Add(sectionID int, sectionTable string, originID int, originTable string, uploadedBy int, path string) error {
	_, err := store.add.Exec(sectionID, sectionTable, originID, originTable, uploadedBy, path)
	return err
}
