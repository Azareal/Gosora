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

/*
	type PasswordReset struct {
		Email string `q:"email"`
		Uid int `q:"uid"`
		Validated bool `q:"validated"`
		Token string `q:"token"`
		CreatedAt time.Time `q:"createdAt"`
	}
*/

func NewDefaultPasswordResetter(acc *qgen.Accumulator) (*DefaultPasswordResetter, error) {
	pr := "password_resets"
	return &DefaultPasswordResetter{
		getTokens: acc.Select(pr).Columns("token").Where("uid = ?").Prepare(),
		create:    acc.Insert(pr).Columns("email, uid, validated, token, createdAt").Fields("?,?,0,?,UTC_TIMESTAMP()").Prepare(),
		//create: acc.Insert(pr).Cols("email,uid,validated=0,token,createdAt=UTC_TIMESTAMP()").Prep(),
		delete:    acc.Delete(pr).Where("uid=?").Prepare(),
		//model:  acc.Model(w).Cols("email,uid,validated=0,token").Key("uid").CreatedAt("createdAt").Prep(),
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

	success := false
	for rows.Next() {
		var rtoken string
		if err := rows.Scan(&rtoken); err != nil {
			return err
		}
		if subtle.ConstantTimeCompare([]byte(token), []byte(rtoken)) == 1 {
			success = true
		}
	}
	if err = rows.Err(); err != nil {
		return err
	}

	if !success {
		return ErrBadResetToken
	}
	return nil
}
