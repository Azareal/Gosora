package common

import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var Subscriptions SubscriptionStore

// ? Should we have a subscription store for each zone? topic, forum, etc?
type SubscriptionStore interface {
	Add(uid int, elementID int, elementType string) error
}

type DefaultSubscriptionStore struct {
	add *sql.Stmt
}

func NewDefaultSubscriptionStore() (*DefaultSubscriptionStore, error) {
	acc := qgen.NewAcc()
	return &DefaultSubscriptionStore{
		add: acc.Insert("activity_subscriptions").Columns("user, targetID, targetType, level").Fields("?,?,?,2").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultSubscriptionStore) Add(uid int, elementID int, elementType string) error {
	_, err := s.add.Exec(uid, elementID, elementType)
	return err
}
