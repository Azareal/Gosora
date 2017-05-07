package main
import "strconv"
import "strings"

var settingLabels map[string]string

type OptionLabel struct
{
	Label string
	Value int
	Selected bool
}

type Setting struct
{
	Name string
	Content string
	Type string
	Constraint string
}

func init() {
	settingLabels = make(map[string]string)
	settingLabels["activation_type"] = "Activate All,Email Activation,Admin Approval"
}

func parseSetting(sname string, scontent string, stype string, constraint string) string {
	var err error
	if stype == "bool" {
		settings[sname] = (scontent == "1")
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
	} else if stype == "list" {
		cons := strings.Split(constraint,"-")
		if len(cons) < 2 {
			return "Invalid constraint! The second field wasn't set!"
		}
		
		con1, err := strconv.Atoi(cons[0])
		if err != nil {
			return "Invalid contraint! The constraint field wasn't an integer!"
		}
		con2, err := strconv.Atoi(cons[1])
		if err != nil {
			return "Invalid contraint! The constraint field wasn't an integer!"
		}
		
		value, err  := strconv.Atoi(scontent)
		if err != nil {
			return "Only integers are allowed in this setting x.x\nType mismatch in " + sname
		}
		
		if value < con1 || value > con2 {
			return "Only integers between a certain range are allowed in this setting"
		}
		settings[sname] = value
	} else {
		settings[sname] = scontent
	}
	return ""
}