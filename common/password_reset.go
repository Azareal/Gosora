package common

import (
	"crypto/subtle"
	"database/sql"
	"errors"

	"github.com/Azareal/Gosora/query_gen"
)

var PasswordResetter *DefaultPasswordResetter
var ErrBadResetToken = errors.New("This reset token has expired.")

type DefaultPasswordResetter struct {
	getTokens *sql.Stmt
	create    *sql.Stmt
	delete    *sql.Stmt
}

func NewDefaultPasswordResetter(acc *qgen.Accumulator) (*DefaultPasswordResetter, error) {
	return &DefaultPasswordResetter{
		getTokens: acc.Select("password_resets").Columns("token").Where("uid = ?").Prepare(),
		create:    acc.Insert("password_resets").Columns("email, uid, validated, token, createdAt").Fields("?,?,0,?,UTC_TIMESTAMP()").Prepare(),
		delete:    acc.Delete("password_resets").Where("uid =?").Prepare(),
	}, acc.FirstError()
}

func (r *DefaultPasswordResetter) Create(email string, uid int, token string) error {
	_, err := r.create.Exec(email, uid, token)
	return err
}

func (r *DefaultPasswordResetter) FlushTokens(uid int) error {
	_, err := r.delete.Exec(uid)
	return err
}

func (r *DefaultPasswordResetter) ValidateToken(uid int, token string) error {
	rows, err := r.getTokens.Query(uid)
	if err != nil {
		return err
	}
	defer rows.Close()

	var success = false
	for rows.Next() {
		var rtoken string
		err := rows.Scan(&rtoken)
		if err != nil {
			return err
		}
		if subtle.ConstantTimeCompare([]byte(token), []byte(rtoken)) == 1 {
			success = true
		}
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	if !success {
		return ErrBadResetToken
	}
	return nil
}
