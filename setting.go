package main
import "strconv"

type Setting struct
{
	Name string
	Content string
	Type string
}

func parseSetting(sname string, scontent string, stype string) string {
	var err error
	if stype == "bool" {
		if scontent == "1" {
			settings[sname] = true
		} else {
			settings[sname] = false
		}
	} else if stype == "int" {
		settings[sname], err = strconv.Atoi(scontent)
		if err != nil {
			return "You were supposed to enter an integer x.x\nType mismatch in " + sname
		}
	} else if stype == "int64" {
		settings[sname], err = strconv.ParseInt(scontent, 10, 64)
		if err != nil {
			return "You were supposed to enter an integer x.x\nType mismatch in " + sname
		}
	} else {
		settings[sname] = scontent
	}
	return ""
}