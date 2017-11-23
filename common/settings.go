package common

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"sync/atomic"

	"../query_gen/lib"
)

var SettingBox atomic.Value // An atomic value pointing to a SettingBox

// SettingMap is a map type specifically for holding the various settings admins set to toggle features on and off or to otherwise alter Gosora's behaviour from the Control Panel
type SettingMap map[string]interface{}

type SettingStore interface {
	ParseSetting(sname string, scontent string, stype string, sconstraint string) string
	BypassGet(name string) (*Setting, error)
	BypassGetAll(name string) ([]*Setting, error)
}

type OptionLabel struct {
	Label    string
	Value    int
	Selected bool
}

type Setting struct {
	Name       string
	Content    string
	Type       string
	Constraint string
}

type SettingStmts struct {
	getAll *sql.Stmt
	get    *sql.Stmt
	update *sql.Stmt
}

var settingStmts SettingStmts

func init() {
	SettingBox.Store(SettingMap(make(map[string]interface{})))
	DbInits.Add(func(acc *qgen.Accumulator) error {
		settingStmts = SettingStmts{
			getAll: acc.Select("settings").Columns("name, content, type, constraints").Prepare(),
			get:    acc.Select("settings").Columns("content, type, constraints").Where("name = ?").Prepare(),
			update: acc.Update("settings").Set("content = ?").Where("name = ?").Prepare(),
		}
		return acc.FirstError()
	})
}

func LoadSettings() error {
	var sBox = SettingMap(make(map[string]interface{}))
	settings, err := sBox.BypassGetAll()
	if err != nil {
		return err
	}

	for _, setting := range settings {
		err = sBox.ParseSetting(setting.Name, setting.Content, setting.Type, setting.Constraint)
		if err != nil {
			return err
		}
	}

	SettingBox.Store(sBox)
	return nil
}

// nolint
var ErrNotInteger = errors.New("You were supposed to enter an integer x.x")
var ErrSettingNotInteger = errors.New("Only integers are allowed in this setting x.x")
var ErrBadConstraintNotInteger = errors.New("Invalid contraint! The constraint field wasn't an integer!")
var ErrBadSettingRange = errors.New("Only integers between a certain range are allowed in this setting")

// To avoid leaking internal state to the user
// TODO: We need to add some sort of DualError interface
func SafeSettingError(err error) bool {
	return err == ErrNotInteger || err == ErrSettingNotInteger || err == ErrBadConstraintNotInteger || err == ErrBadSettingRange || err == ErrNoRows
}

// TODO: Add better support for HTML attributes (html-attribute). E.g. Meta descriptions.
func (sBox SettingMap) ParseSetting(sname string, scontent string, stype string, constraint string) (err error) {
	var ssBox = map[string]interface{}(sBox)
	switch stype {
	case "bool":
		ssBox[sname] = (scontent == "1")
	case "int":
		ssBox[sname], err = strconv.Atoi(scontent)
		if err != nil {
			return ErrNotInteger
		}
	case "int64":
		ssBox[sname], err = strconv.ParseInt(scontent, 10, 64)
		if err != nil {
			return ErrNotInteger
		}
	case "list":
		cons := strings.Split(constraint, "-")
		if len(cons) < 2 {
			return errors.New("Invalid constraint! The second field wasn't set!")
		}

		con1, err := strconv.Atoi(cons[0])
		con2, err2 := strconv.Atoi(cons[1])
		if err != nil || err2 != nil {
			return ErrBadConstraintNotInteger
		}

		value, err := strconv.Atoi(scontent)
		if err != nil {
			return ErrSettingNotInteger
		}

		if value < con1 || value > con2 {
			return ErrBadSettingRange
		}
		ssBox[sname] = value
	default:
		ssBox[sname] = scontent
	}
	return nil
}

func (sBox SettingMap) BypassGet(name string) (*Setting, error) {
	setting := &Setting{Name: name}
	err := settingStmts.get.QueryRow(name).Scan(&setting.Content, &setting.Type, &setting.Constraint)
	return setting, err
}

func (sBox SettingMap) BypassGetAll() (settingList []*Setting, err error) {
	rows, err := settingStmts.getAll.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		setting := &Setting{Name: ""}
		err := rows.Scan(&setting.Name, &setting.Content, &setting.Type, &setting.Constraint)
		if err != nil {
			return nil, err
		}
		settingList = append(settingList, setting)
	}
	return settingList, rows.Err()
}

func (sBox SettingMap) Update(name string, content string) error {
	setting, err := sBox.BypassGet(name)
	if err == ErrNoRows {
		return err
	}

	// TODO: Why is this here and not in a common function?
	if setting.Type == "bool" {
		if content == "on" || content == "1" {
			content = "1"
		} else {
			content = "0"
		}
	}

	// TODO: Make this a method or function?
	_, err = settingStmts.update.Exec(content, name)
	if err != nil {
		return err
	}

	err = sBox.ParseSetting(name, content, setting.Type, setting.Constraint)
	if err != nil {
		return err
	}
	// TODO: Do a reload instead?
	SettingBox.Store(sBox)
	return nil
}
