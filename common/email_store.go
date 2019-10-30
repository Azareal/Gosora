package common

import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var Emails EmailStore

type Email struct {
	UserID    int
	Email     string
	Validated bool
	Primary   bool
	Token     string
}

type EmailStore interface {
	// TODO: Add an autoincrement key
	Get(user *User, email string) (Email, error)
	GetEmailsByUser(user *User) (emails []Email, err error)
	Add(uid int, email, token string) error
	Delete(uid int, email string) error
	VerifyEmail(email string) error
}

type DefaultEmailStore struct {
	get *sql.Stmt
	getEmailsByUser *sql.Stmt
	add *sql.Stmt
	delete *sql.Stmt
	verifyEmail     *sql.Stmt
}

func NewDefaultEmailStore(acc *qgen.Accumulator) (*DefaultEmailStore, error) {
	e := "emails"
	return &DefaultEmailStore{
		get: acc.Select(e).Columns("email,validated,token").Where("uid=? AND email=?").Prepare(),
		getEmailsByUser: acc.Select(e).Columns("email,validated,token").Where("uid=?").Prepare(),
		add: acc.Insert(e).Columns("uid,email,validated,token").Fields("?,?,?,?").Prepare(),
		delete: acc.Delete(e).Where("uid=? AND email=?").Prepare(),

		// Need to fix this: Empty string isn't working, it gets set to 1 instead x.x -- Has this been fixed?
		verifyEmail: acc.Update(e).Set("validated=1,token=''").Where("email=?").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultEmailStore) Get(user *User, email string) (Email, error) {
	e := Email{UserID:user.ID, Primary:email !="" && user.Email==email}
	err := s.get.QueryRow(user.ID, email).Scan(&e.Email, &e.Validated, &e.Token)
	return e, err
}

func (s *DefaultEmailStore) GetEmailsByUser(user *User) (emails []Email, err error) {
	e := Email{UserID: user.ID}
	rows, err := s.getEmailsByUser.Query(user.ID)
	if err != nil {
		return emails, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&e.Email, &e.Validated, &e.Token)
		if err != nil {
			return emails, err
		}
		if e.Email == user.Email {
			e.Primary = true
		}
		emails = append(emails, e)
	}
	return emails, rows.Err()
}

func (s *DefaultEmailStore) Add(uid int, email string,  token string) error {
	_, err := s.add.Exec(uid, email, 0, token)
	return err
}

func (s *DefaultEmailStore) Delete(uid int, email string) error {
	_, err := s.delete.Exec(uid,email)
	return err
}

func (s *DefaultEmailStore) VerifyEmail(email string) error {
	_, err := s.verifyEmail.Exec(email)
	return err
}
