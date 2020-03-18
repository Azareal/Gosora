package panel

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
	p "github.com/Azareal/Gosora/common/phrases"
)

func Settings(w http.ResponseWriter, r *http.Request, user *c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, user, "settings", "settings")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return c.NoPermissions(w, r, user)
	}

	// TODO: What if the list gets too long? How should we structure this?
	settings, err := basePage.Settings.BypassGetAll()
	if err != nil {
		return c.InternalError(err, w, r)
	}
	settingPhrases := p.GetAllSettingPhrases()

	var settingList []*c.PanelSetting
	for _, settingPtr := range settings {
		s := settingPtr.Copy()
		if s.Type == "list" {
			llist := settingPhrases[s.Name+"_label"]
			labels := strings.Split(llist, ",")
			conv, err := strconv.Atoi(s.Content)
			if err != nil {
				return c.LocalError("The setting '"+s.Name+"' can't be converted to an integer", w, r, user)
			}
			s.Content = labels[conv-1]
			// TODO: Localise this
		} else if s.Type == "bool" {
			if s.Content == "1" {
				s.Content = "Yes"
			} else {
				s.Content = "No"
			}
		} else if s.Type == "html-attribute" {
			s.Type = "textarea"
		}
		settingList = append(settingList, &c.PanelSetting{s, p.GetSettingPhrase(s.Name)})
	}

	pi := c.PanelPage{basePage, tList, settingList}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "", "", "panel_settings", &pi})
}

func SettingEdit(w http.ResponseWriter, r *http.Request, user *c.User, sname string) c.RouteError {
	basePage, ferr := buildBasePage(w, r, user, "edit_setting", "settings")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return c.NoPermissions(w, r, user)
	}

	setting, err := basePage.Settings.BypassGet(sname)
	if err == sql.ErrNoRows {
		return c.LocalError("The setting you want to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	var itemList []c.OptionLabel
	if setting.Type == "list" {
		llist := p.GetSettingPhrase(setting.Name + "_label")
		conv, err := strconv.Atoi(setting.Content)
		if err != nil {
			return c.LocalError("The value of this setting couldn't be converted to an integer", w, r, user)
		}
		for index, label := range strings.Split(llist, ",") {
			itemList = append(itemList, c.OptionLabel{
				Label:    label,
				Value:    index + 1,
				Selected: conv == (index + 1),
			})
		}
	} else if setting.Type == "html-attribute" {
		setting.Type = "textarea"
	}

	pSetting := &c.PanelSetting{setting, p.GetSettingPhrase(setting.Name)}
	pi := c.PanelSettingPage{basePage, itemList, pSetting}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "", "", "panel_setting", &pi})
}

func SettingEditSubmit(w http.ResponseWriter, r *http.Request, user *c.User, name string) c.RouteError {
	headerLite, ferr := c.SimplePanelUserCheck(w, r, user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return c.NoPermissions(w, r, user)
	}

	name = c.SanitiseSingleLine(name)
	content := c.SanitiseBody(r.PostFormValue("value"))
	rerr := headerLite.Settings.Update(name, content)
	if rerr != nil {
		return rerr
	}
	// TODO: Avoid this hack
	err := c.AdminLogs.Create(name, 0, "setting", user.GetIP(), user.ID)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/settings/", http.StatusSeeOther)
	return nil
}
