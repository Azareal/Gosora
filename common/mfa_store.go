package common

import (
	"database/sql"
	"errors"
	"strings"

	qgen "github.com/Azareal/Gosora/query_gen"
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
			update: acc.Update("users_2fa_keys").Set("scratch1=?,scratch2=?,scratch3=?,scratch4=?,scratch5=?,scratch6=?,scratch7=?,scratch8=?").Where("uid=?").Prepare(),
			delete: acc.Delete("users_2fa_keys").Where("uid=?").Prepare(),
		}
		return acc.FirstError()
	})
}

type MFAItem struct {
	UID     int
	Secret  string
	Scratch []string
}

func (i *MFAItem) BurnScratch(index int) error {
	if index < 0 || len(i.Scratch) <= index {
		return ErrMFAScratchIndexOutOfBounds
	}
	newScratch, err := mfaCreateScratch()
	if err != nil {
		return err
	}
	i.Scratch[index] = newScratch

	_, err = mfaItemStmts.update.Exec(i.Scratch[0], i.Scratch[1], i.Scratch[2], i.Scratch[3], i.Scratch[4], i.Scratch[5], i.Scratch[6], i.Scratch[7], i.UID)
	return err
}

func (i *MFAItem) Delete() error {
	_, err := mfaItemStmts.delete.Exec(i.UID)
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
		get:    acc.Select("users_2fa_keys").Columns("secret,scratch1,scratch2,scratch3,scratch4,scratch5,scratch6,scratch7,scratch8").Where("uid=?").Prepare(),
		create: acc.Insert("users_2fa_keys").Columns("uid,secret,scratch1,scratch2,scratch3,scratch4,scratch5,scratch6,scratch7,scratch8,createdAt").Fields("?,?,?,?,?,?,?,?,?,?,UTC_TIMESTAMP()").Prepare(),
	}, acc.FirstError()
}

// TODO: Write a test for this
func (s *SQLMFAStore) Get(id int) (*MFAItem, error) {
	i := MFAItem{UID: id, Scratch: make([]string, 8)}
	err := s.get.QueryRow(id).Scan(&i.Secret, &i.Scratch[0], &i.Scratch[1], &i.Scratch[2], &i.Scratch[3], &i.Scratch[4], &i.Scratch[5], &i.Scratch[6], &i.Scratch[7])
	return &i, err

}

// TODO: Write a test for this
func (s *SQLMFAStore) Create(secret string, uid int) (err error) {
	params := make([]interface{}, 10)
	params[0] = uid
	params[1] = secret
	for i := 2; i < len(params); i++ {
		code, err := mfaCreateScratch()
		if err != nil {
			return err
		}
		params[i] = code
	}

	_, err = s.create.Exec(params...)
	return err
}
