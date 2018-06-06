package panel

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"../../common"
)

func Settings(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	header, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return common.NoPermissions(w, r, user)
	}
	header.Title = common.GetTitlePhrase("panel_settings")

	settings, err := header.Settings.BypassGetAll()
	if err != nil {
		return common.InternalError(err, w, r)
	}
	settingPhrases := common.GetAllSettingPhrases()

	var settingList []*common.PanelSetting
	for _, settingPtr := range settings {
		setting := settingPtr.Copy()
		if setting.Type == "list" {
			llist := settingPhrases[setting.Name+"_label"]
			labels := strings.Split(llist, ",")
			conv, err := strconv.Atoi(setting.Content)
			if err != nil {
				return common.LocalError("The setting '"+setting.Name+"' can't be converted to an integer", w, r, user)
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
		settingList = append(settingList, &common.PanelSetting{setting, common.GetSettingPhrase(setting.Name)})
	}

	pi := common.PanelPage{&common.BasePanelPage{header, stats, "settings", common.ReportForumID}, tList, settingList}
	return panelRenderTemplate("panel_settings", w, r, user, &pi)
}

func SettingEdit(w http.ResponseWriter, r *http.Request, user common.User, sname string) common.RouteError {
	header, stats, ferr := common.PanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return common.NoPermissions(w, r, user)
	}
	header.Title = common.GetTitlePhrase("panel_edit_setting")

	setting, err := header.Settings.BypassGet(sname)
	if err == sql.ErrNoRows {
		return common.LocalError("The setting you want to edit doesn't exist.", w, r, user)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	var itemList []common.OptionLabel
	if setting.Type == "list" {
		llist := common.GetSettingPhrase(setting.Name + "_label")
		conv, err := strconv.Atoi(setting.Content)
		if err != nil {
			return common.LocalError("The value of this setting couldn't be converted to an integer", w, r, user)
		}
		fmt.Println("llist: ", llist)

		for index, label := range strings.Split(llist, ",") {
			itemList = append(itemList, common.OptionLabel{
				Label:    label,
				Value:    index + 1,
				Selected: conv == (index + 1),
			})
		}
	} else if setting.Type == "html-attribute" {
		setting.Type = "textarea"
	}

	pSetting := &common.PanelSetting{setting, common.GetSettingPhrase(setting.Name)}
	pi := common.PanelSettingPage{&common.BasePanelPage{header, stats, "settings", common.ReportForumID}, itemList, pSetting}
	return panelRenderTemplate("panel_setting", w, r, user, &pi)
}

func SettingEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, sname string) common.RouteError {
	headerLite, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.EditSettings {
		return common.NoPermissions(w, r, user)
	}

	scontent := common.SanitiseBody(r.PostFormValue("setting-value"))
	err := headerLite.Settings.Update(sname, scontent)
	if err != nil {
		if common.SafeSettingError(err) {
			return common.LocalError(err.Error(), w, r, user)
		}
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/settings/", http.StatusSeeOther)
	return nil
}
