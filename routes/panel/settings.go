package panel

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

func Settings(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "settings", "settings")
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
	settingPhrases := phrases.GetAllSettingPhrases()

	var settingList []*c.PanelSetting
	for _, settingPtr := range settings {
		setting := settingPtr.Copy()
		if setting.Type == "list" {
			llist := settingPhrases[setting.Name+"_label"]
			labels := strings.Split(llist, ",")
			conv, err := strconv.Atoi(setting.Content)
			if err != nil {
				return c.LocalError("The setting '"+setting.Name+"' can't be converted to an integer", w, r, user)
			}
			setting.Content = labels[conv-1]
		} else if setting.Type == "bool" {
			if setting.Content == "1" {
				setting.Content = "Yes"
			} else {
				setting.Content = "No"
			}
		} else if setting.Type == "html-attribute" {
			setting.Type = "textarea"
		}
		settingList = append(settingList, &c.PanelSetting{setting, phrases.GetSettingPhrase(setting.Name)})
	}

	pi := c.PanelPage{basePage, tList, settingList}
	return renderTemplate("panel_settings", w, r, basePage.Header, &pi)
}

func SettingEdit(w http.ResponseWriter, r *http.Request, user c.User, sname string) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "edit_setting", "settings")
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
		llist := phrases.GetSettingPhrase(setting.Name + "_label")
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

	pSetting := &c.PanelSetting{setting, phrases.GetSettingPhrase(setting.Name)}
	pi := c.PanelSettingPage{basePage, itemList, pSetting}
	return renderTemplate("panel_setting", w, r, basePage.Header, &pi)
}

func SettingEditSubmit(w http.ResponseWriter, r *http.Request, user c.User, sname string) c.RouteError {
	headerLite, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return c.NoPermissions(w, r, user)
	}

	scontent := c.SanitiseBody(r.PostFormValue("setting-value"))
	rerr := headerLite.Settings.Update(sname, scontent)
	if rerr != nil {
		return rerr
	}

	http.Redirect(w, r, "/panel/settings/", http.StatusSeeOther)
	return nil
}
