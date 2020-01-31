package common

import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var Subscriptions SubscriptionStore

// ? Should we have a subscription store for each zone? topic, forum, etc?
type SubscriptionStore interface {
	Add(uid, elementID int, elementType string) error
	Delete(uid, targetID int, targetType string) error
	DeleteResource(targetID int, targetType string) error
}

type DefaultSubscriptionStore struct {
	add            *sql.Stmt
	delete         *sql.Stmt
	deleteResource *sql.Stmt
}

func NewDefaultSubscriptionStore() (*DefaultSubscriptionStore, error) {
	acc := qgen.NewAcc()
	ast := "activity_subscriptions"
	return &DefaultSubscriptionStore{
		add:            acc.Insert(ast).Columns("user, targetID, targetType, level").Fields("?,?,?,2").Prepare(),
		delete:         acc.Delete(ast).Where("user=? AND targetID=? AND targetType=?").Prepare(),
		deleteResource: acc.Delete(ast).Where("targetID=? AND targetType=?").Prepare(),
	}, acc.FirstError()
}

func (s *DefaultSubscriptionStore) Add(uid, elementID int, elementType string) error {
	_, err := s.add.Exec(uid, elementID, elementType)
	return err
}

// TODO: Add a primary key to the activity subscriptions table
func (s *DefaultSubscriptionStore) Delete(uid, targetID int, targetType string) error {
	_, err := s.delete.Exec(uid, targetID, targetType)
	return err
}

func (s *DefaultSubscriptionStore) DeleteResource(targetID int, targetType string) error {
	_, err := s.deleteResource.Exec(targetID, targetType)
	return err
}
