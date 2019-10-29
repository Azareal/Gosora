package common

import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var Emails EmailStore

type EmailStore interface {
	GetEmailsByUser(user *User) (emails []Email, err error)
	VerifyEmail(email string) error
}

type DefaultEmailStore struct {
	getEmailsByUser *sql.Stmt
	verifyEmail     *sql.Stmt
}

func NewDefaultEmailStore(acc *qgen.Accumulator) (*DefaultEmailStore, error) {
	return &DefaultEmailStore{
		getEmailsByUser: acc.Select("emails").Columns("email,validated,token").Where("uid=?").Prepare(),

		// Need to fix this: Empty string isn't working, it gets set to 1 instead x.x -- Has this been fixed?
		verifyEmail: acc.Update("emails").Set("validated = 1, token = ''").Where("email = ?").Prepare(),
	}, acc.FirstError()
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

func (s *DefaultEmailStore) VerifyEmail(email string) error {
	_, err := s.verifyEmail.Exec(email)
	return err
}
