package common

// EXPERIMENTAL
import (
	"errors"
)

var StatStore StatStoreInt

type StatStoreInt interface {
	LookupInt(name string, duration int, unit string) (int, error)
}

type DefaultStatStore struct {
}

func NewDefaultStatStore() *DefaultStatStore {
	return &DefaultStatStore{}
}

func (store *DefaultStatStore) LookupInt(name string, duration int, unit string) (int, error) {
	switch name {
	case "postCount":
		return store.countTable("replies", duration, unit)
	}
	return 0, errors.New("The requested stat doesn't exist")
}

func (store *DefaultStatStore) countTable(table string, duration int, unit string) (stat int, err error) {
	/*counter := qgen.NewAcc().Count("replies").DateCutoff("createdAt", 1, "day").Prepare()
	if acc.FirstError() != nil {
		return 0, acc.FirstError()
	}
	err := counter.QueryRow().Scan(&stat)*/
	return stat, err
}

//stmts.todaysPostCount, err = db.Prepare("select count(*) from replies where createdAt BETWEEN (utc_timestamp() - interval 1 day) and utc_timestamp()")
