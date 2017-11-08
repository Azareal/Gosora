package main

import "strconv"
import "strings"
import "sync/atomic"

// SettingBox is a map type specifically for holding the various settings admins set to toggle features on and off or to otherwise alter Gosora's behaviour from the Control Panel
type SettingBox map[string]interface{}

var settingBox atomic.Value // An atomic value pointing to a SettingBox

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

func init() {
	settingBox.Store(SettingBox(make(map[string]interface{})))
}

func LoadSettings() error {
	rows, err := stmts.getFullSettings.Query()
	if err != nil {
		return err
	}
	defer rows.Close()

	var sBox = SettingBox(make(map[string]interface{}))
	var sname, scontent, stype, sconstraints string
	for rows.Next() {
		err = rows.Scan(&sname, &scontent, &stype, &sconstraints)
		if err != nil {
			return err
		}
		errmsg := sBox.ParseSetting(sname, scontent, stype, sconstraints)
		if errmsg != "" {
			return err
		}
	}
	err = rows.Err()
	if err != nil {
		return err
	}

	settingBox.Store(sBox)
	return nil
}

// TODO: Add better support for HTML attributes (html-attribute). E.g. Meta descriptions.
func (sBox SettingBox) ParseSetting(sname string, scontent string, stype string, constraint string) string {
	var err error
	var ssBox = map[string]interface{}(sBox)
	if stype == "bool" {
		ssBox[sname] = (scontent == "1")
	} else if stype == "int" {
		ssBox[sname], err = strconv.Atoi(scontent)
		if err != nil {
			return "You were supposed to enter an integer x.x\nType mismatch in " + sname
		}
	} else if stype == "int64" {
		ssBox[sname], err = strconv.ParseInt(scontent, 10, 64)
		if err != nil {
			return "You were supposed to enter an integer x.x\nType mismatch in " + sname
		}
	} else if stype == "list" {
		cons := strings.Split(constraint, "-")
		if len(cons) < 2 {
			return "Invalid constraint! The second field wasn't set!"
		}

		con1, err := strconv.Atoi(cons[0])
		con2, err2 := strconv.Atoi(cons[1])
		if err != nil || err2 != nil {
			return "Invalid contraint! The constraint field wasn't an integer!"
		}

		value, err := strconv.Atoi(scontent)
		if err != nil {
			return "Only integers are allowed in this setting x.x\nType mismatch in " + sname
		}

		if value < con1 || value > con2 {
			return "Only integers between a certain range are allowed in this setting"
		}
		ssBox[sname] = value
	} else {
		ssBox[sname] = scontent
	}
	return ""
}
