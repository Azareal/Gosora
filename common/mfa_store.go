package common

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/Azareal/Gosora/query_gen"
)

var MFAstore MFAStore
var ErrMFAScratchIndexOutOfBounds = errors.New("That MFA scratch index is out of bounds")

type MFAItemStmts struct {
	update *sql.Stmt
	delete *sql.Stmt
}

var mfaItemStmts MFAItemStmts

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		mfaItemStmts = MFAItemStmts{
			update: acc.Update("users_2fa_keys").Set("scratch1 = ?, scratch2, scratch3 = ?, scratch3 = ?, scratch4 = ?, scratch5 = ?, scratch6 = ?, scratch7 = ?, scratch8 = ?").Where("uid = ?").Prepare(),
			delete: acc.Delete("users_2fa_keys").Where("uid = ?").Prepare(),
		}
		return acc.FirstError()
	})
}

type MFAItem struct {
	UID     int
	Secret  string
	Scratch []string
}

func (item *MFAItem) BurnScratch(index int) error {
	if index < 0 || len(item.Scratch) <= index {
		return ErrMFAScratchIndexOutOfBounds
	}
	newScratch, err := mfaCreateScratch()
	if err != nil {
		return err
	}
	item.Scratch[index] = newScratch

	_, err = mfaItemStmts.update.Exec(item.Scratch[0], item.Scratch[1], item.Scratch[2], item.Scratch[3], item.Scratch[4], item.Scratch[5], item.Scratch[6], item.Scratch[7], item.UID)
	return err
}

func (item *MFAItem) Delete() error {
	_, err := mfaItemStmts.delete.Exec(item.UID)
	return err
}

func mfaCreateScratch() (string, error) {
	code, err := GenerateStd32SafeString(8)
	return strings.Replace(code, "=", "", -1), err
}

type MFAStore interface {
	Get(id int) (*MFAItem, error)
	Create(secret string, uid int) (err error)
}

type SQLMFAStore struct {
	get    *sql.Stmt
	create *sql.Stmt
}

func NewSQLMFAStore(acc *qgen.Accumulator) (*SQLMFAStore, error) {
	return &SQLMFAStore{
		get:    acc.Select("users_2fa_keys").Columns("secret, scratch1, scratch2, scratch3, scratch4, scratch5, scratch6, scratch7, scratch8").Where("uid = ?").Prepare(),
		create: acc.Insert("users_2fa_keys").Columns("uid, secret, scratch1, scratch2, scratch3, scratch4, scratch5, scratch6, scratch7, scratch8, createdAt").Fields("?,?,?,?,?,?,?,?,?,?,UTC_TIMESTAMP()").Prepare(),
	}, acc.FirstError()
}

// TODO: Write a test for this
func (store *SQLMFAStore) Get(id int) (*MFAItem, error) {
	item := MFAItem{UID: id, Scratch: make([]string, 8)}
	err := store.get.QueryRow(id).Scan(&item.Secret, &item.Scratch[0], &item.Scratch[1], &item.Scratch[2], &item.Scratch[3], &item.Scratch[4], &item.Scratch[5], &item.Scratch[6], &item.Scratch[7])
	return &item, err

}

// TODO: Write a test for this
func (store *SQLMFAStore) Create(secret string, uid int) (err error) {
	var params = make([]interface{}, 10)
	params[0] = uid
	params[1] = secret
	for i := 2; i < len(params); i++ {
		code, err := mfaCreateScratch()
		if err != nil {
			return err
		}
		params[i] = code
	}

	_, err = store.create.Exec(params...)
	return err
}
