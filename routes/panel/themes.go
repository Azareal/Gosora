package panel

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
)

func Themes(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "themes", "themes")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return common.NoPermissions(w, r, user)
	}

	var pThemeList, vThemeList []*common.Theme
	for _, theme := range common.Themes {
		if theme.HideFromThemes {
			continue
		}
		if theme.ForkOf == "" {
			pThemeList = append(pThemeList, theme)
		} else {
			vThemeList = append(vThemeList, theme)
		}
	}

	pi := common.PanelThemesPage{basePage, pThemeList, vThemeList}
	return renderTemplate("panel_themes", w, r, user, &pi)
}

func ThemesSetDefault(w http.ResponseWriter, r *http.Request, user common.User, uname string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return common.NoPermissions(w, r, user)
	}

	theme, ok := common.Themes[uname]
	if !ok {
		return common.LocalError("The theme isn't registered in the system", w, r, user)
	}
	if theme.Disabled {
		return common.LocalError("You must not enable this theme", w, r, user)
	}

	err := common.UpdateDefaultTheme(theme)
	if err != nil {
		return common.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/themes/", http.StatusSeeOther)
	return nil
}

func ThemesMenus(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "themes_menus", "themes")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return common.NoPermissions(w, r, user)
	}

	var menuList []common.PanelMenuListItem
	for mid, list := range common.Menus.GetAllMap() {
		var name = ""
		if mid == 1 {
			name = phrases.GetTmplPhrase("panel_themes_menus_main")
		}
		menuList = append(menuList, common.PanelMenuListItem{
			Name:      name,
			ID:        mid,
			ItemCount: len(list.List),
		})
	}

	pi := common.PanelMenuListPage{basePage, menuList}
	return renderTemplate("panel_themes_menus", w, r, user, &pi)
}

func ThemesMenusEdit(w http.ResponseWriter, r *http.Request, user common.User, smid string) common.RouteError {
	// TODO: Something like Menu #1 for the title?
	basePage, ferr := buildBasePage(w, r, &user, "themes_menus_edit", "themes")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return common.NoPermissions(w, r, user)
	}
	basePage.Header.AddScript("Sortable-1.4.0/Sortable.min.js")

	mid, err := strconv.Atoi(smid)
	if err != nil {
		return common.LocalError(phrases.GetErrorPhrase("url_id_must_be_integer"), w, r, user)
	}

	menuHold, err := common.Menus.Get(mid)
	if err == sql.ErrNoRows {
		return common.NotFound(w, r, basePage.Header)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	var menuList []common.MenuItem
	for _, item := range menuHold.List {
		var menuTmpls = map[string]common.MenuTmpl{
			item.TmplName: menuHold.Parse(item.Name, []byte("{{.Name}}")),
		}
		var renderBuffer [][]byte
		var variableIndices []int
		renderBuffer, _ = menuHold.ScanItem(menuTmpls, item, renderBuffer, variableIndices)

		var out string
		for _, renderItem := range renderBuffer {
			out += string(renderItem)
		}
		item.Name = out
		if item.Name == "" {
			item.Name = "???"
		}
		menuList = append(menuList, item)
	}

	pi := common.PanelMenuPage{basePage, mid, menuList}
	return renderTemplate("panel_themes_menus_items", w, r, user, &pi)
}

func ThemesMenuItemEdit(w http.ResponseWriter, r *http.Request, user common.User, sitemID string) common.RouteError {
	// TODO: Something like Menu #1 for the title?
	basePage, ferr := buildBasePage(w, r, &user, "themes_menus_edit", "themes")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return common.NoPermissions(w, r, user)
	}

	itemID, err := strconv.Atoi(sitemID)
	if err != nil {
		return common.LocalError(phrases.GetErrorPhrase("url_id_must_be_integer"), w, r, user)
	}

	menuItem, err := common.Menus.ItemStore().Get(itemID)
	if err == sql.ErrNoRows {
		return common.NotFound(w, r, basePage.Header)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	pi := common.PanelMenuItemPage{basePage, menuItem}
	return renderTemplate("panel_themes_menus_item_edit", w, r, user, &pi)
}

func themesMenuItemSetters(r *http.Request, menuItem common.MenuItem) common.MenuItem {
	var getItem = func(name string) string {
		return common.SanitiseSingleLine(r.PostFormValue("item-" + name))
	}
	menuItem.Name = getItem("name")
	menuItem.HTMLID = getItem("htmlid")
	menuItem.CSSClass = getItem("cssclass")
	menuItem.Position = getItem("position")
	if menuItem.Position != "left" && menuItem.Position != "right" {
		menuItem.Position = "left"
	}
	menuItem.Path = getItem("path")
	menuItem.Aria = getItem("aria")
	menuItem.Tooltip = getItem("tooltip")
	menuItem.TmplName = getItem("tmplname")

	switch getItem("permissions") {
	case "everyone":
		menuItem.GuestOnly = false
		menuItem.MemberOnly = false
		menuItem.SuperModOnly = false
		menuItem.AdminOnly = false
	case "guest-only":
		menuItem.GuestOnly = true
		menuItem.MemberOnly = false
		menuItem.SuperModOnly = false
		menuItem.AdminOnly = false
	case "member-only":
		menuItem.GuestOnly = false
		menuItem.MemberOnly = true
		menuItem.SuperModOnly = false
		menuItem.AdminOnly = false
	case "supermod-only":
		menuItem.GuestOnly = false
		menuItem.MemberOnly = true
		menuItem.SuperModOnly = true
		menuItem.AdminOnly = false
	case "admin-only":
		menuItem.GuestOnly = false
		menuItem.MemberOnly = true
		menuItem.SuperModOnly = true
		menuItem.AdminOnly = true
	}
	return menuItem
}

func ThemesMenuItemEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, sitemID string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	isJs := (r.PostFormValue("js") == "1")
	if !user.Perms.ManageThemes {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	itemID, err := strconv.Atoi(sitemID)
	if err != nil {
		return common.LocalErrorJSQ(phrases.GetErrorPhrase("id_must_be_integer"), w, r, user, isJs)
	}

	menuItem, err := common.Menus.ItemStore().Get(itemID)
	if err == sql.ErrNoRows {
		return common.LocalErrorJSQ("This item doesn't exist.", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	//menuItem = menuItem.Copy() // If we switch this for a pointer, we might need this as a scratchpad
	menuItem = themesMenuItemSetters(r, menuItem)

	err = menuItem.Commit()
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	return successRedirect("/panel/themes/menus/item/edit/"+strconv.Itoa(itemID), w, r, isJs)
}

func ThemesMenuItemCreateSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	isJs := (r.PostFormValue("js") == "1")
	if !user.Perms.ManageThemes {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}
	smenuID := r.PostFormValue("mid")
	if smenuID == "" {
		return common.LocalErrorJSQ("No menuID provided", w, r, user, isJs)
	}
	menuID, err := strconv.Atoi(smenuID)
	if err != nil {
		return common.LocalErrorJSQ(phrases.GetErrorPhrase("id_must_be_integer"), w, r, user, isJs)
	}

	menuItem := common.MenuItem{MenuID: menuID}
	menuItem = themesMenuItemSetters(r, menuItem)
	itemID, err := menuItem.Create()
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	return successRedirect("/panel/themes/menus/item/edit/"+strconv.Itoa(itemID), w, r, isJs)
}

func ThemesMenuItemDeleteSubmit(w http.ResponseWriter, r *http.Request, user common.User, sitemID string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	isJs := (r.PostFormValue("js") == "1")
	if !user.Perms.ManageThemes {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	itemID, err := strconv.Atoi(sitemID)
	if err != nil {
		return common.LocalErrorJSQ(phrases.GetErrorPhrase("id_must_be_integer"), w, r, user, isJs)
	}
	menuItem, err := common.Menus.ItemStore().Get(itemID)
	if err == sql.ErrNoRows {
		return common.LocalErrorJSQ("This item doesn't exist.", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	//menuItem = menuItem.Copy() // If we switch this for a pointer, we might need this as a scratchpad

	err = menuItem.Delete()
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	return successRedirect("/panel/themes/menus/", w, r, isJs)
}

func ThemesMenuItemOrderSubmit(w http.ResponseWriter, r *http.Request, user common.User, smid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	isJs := (r.PostFormValue("js") == "1")
	if !user.Perms.ManageThemes {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	mid, err := strconv.Atoi(smid)
	if err != nil {
		return common.LocalErrorJSQ(phrases.GetErrorPhrase("id_must_be_integer"), w, r, user, isJs)
	}
	menuHold, err := common.Menus.Get(mid)
	if err == sql.ErrNoRows {
		return common.LocalErrorJSQ("Can't find menu", w, r, user, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	sitems := strings.TrimSuffix(strings.TrimPrefix(r.PostFormValue("items"), "{"), "}")
	//fmt.Printf("sitems: %+v\n", sitems)

	var updateMap = make(map[int]int)
	for index, smiid := range strings.Split(sitems, ",") {
		miid, err := strconv.Atoi(smiid)
		if err != nil {
			return common.LocalErrorJSQ("Invalid integer in menu item list", w, r, user, isJs)
		}
		updateMap[miid] = index
	}
	menuHold.UpdateOrder(updateMap)

	return successRedirect("/panel/themes/menus/edit/"+strconv.Itoa(mid), w, r, isJs)
}

func ThemesWidgets(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "themes_widgets", "themes")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return common.NoPermissions(w, r, user)
	}
	basePage.Header.AddScript("widgets.js")

	var docks = make(map[string][]common.WidgetEdit)
	for _, name := range common.GetDockList() {
		var widgets []common.WidgetEdit
		for _, widget := range common.GetDock(name) {
			var data = make(map[string]string)
			err := json.Unmarshal([]byte(widget.RawBody), &data)
			if err != nil {
				return common.InternalError(err, w, r)
			}
			widgets = append(widgets, common.WidgetEdit{widget, data})
		}
		docks[name] = widgets
	}

	pi := common.PanelWidgetListPage{basePage, docks, common.WidgetEdit{&common.Widget{ID: 0, Type: "simple"}, make(map[string]string)}}
	return renderTemplate("panel_themes_widgets", w, r, user, &pi)
}

func widgetsParseInputs(r *http.Request, widget *common.Widget) (*common.WidgetEdit, error) {
	var data = make(map[string]string)
	widget.Enabled = (r.FormValue("wenabled") == "1")
	widget.Location = r.FormValue("wlocation")
	if widget.Location == "" {
		return nil, errors.New("You need to specify a location for this widget.")
	}
	widget.Side = r.FormValue("wside")
	if !common.HasDock(widget.Side) {
		return nil, errors.New("The widget dock you specified doesn't exist.")
	}

	var wtype = r.FormValue("wtype")
	switch wtype {
	case "simple", "about":
		data["Name"] = r.FormValue("wname")
		if data["Name"] == "" {
			return nil, errors.New("You need to specify a title for this widget.")
		}
		data["Text"] = r.FormValue("wtext")
		if data["Text"] == "" {
			return nil, errors.New("You need to fill in the body for this widget.")
		}
		widget.Type = wtype // ? - Are we sure we should be directly assigning user provided data even if it's validated?
	case "wol", "search_and_filter":
		widget.Type = wtype // ? - Are we sure we should be directly assigning user provided data even if it's validated?
	default:
		return nil, errors.New("Unknown widget type")
	}

	return &common.WidgetEdit{widget, data}, nil
}

// ThemesWidgetsEditSubmit is an action which is triggered when someone sends an update request for a widget
func ThemesWidgetsEditSubmit(w http.ResponseWriter, r *http.Request, user common.User, swid string) common.RouteError {
	fmt.Println("in ThemesWidgetsEditSubmit")
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	isJs := (r.PostFormValue("js") == "1")
	if !user.Perms.ManageThemes {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	wid, err := strconv.Atoi(swid)
	if err != nil {
		return common.LocalErrorJSQ(phrases.GetErrorPhrase("id_must_be_integer"), w, r, user, isJs)
	}

	widget, err := common.Widgets.Get(wid)
	if err == sql.ErrNoRows {
		return common.NotFoundJSQ(w, r, nil, isJs)
	} else if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	ewidget, err := widgetsParseInputs(r, widget.Copy())
	if err != nil {
		return common.LocalErrorJSQ(err.Error(), w, r, user, isJs)
	}

	err = ewidget.Commit()
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	return successRedirect("/panel/themes/widgets/", w, r, isJs)
}

// ThemesWidgetsCreateSubmit is an action which is triggered when someone sends a create request for a widget
func ThemesWidgetsCreateSubmit(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	fmt.Println("in ThemesWidgetsCreateSubmit")
	isJs := (r.PostFormValue("js") == "1")
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	ewidget, err := widgetsParseInputs(r, &common.Widget{})
	if err != nil {
		return common.LocalErrorJSQ(err.Error(), w, r, user, isJs)
	}

	err = ewidget.Create()
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}

	return successRedirect("/panel/themes/widgets/", w, r, isJs)
}

func ThemesWidgetsDeleteSubmit(w http.ResponseWriter, r *http.Request, user common.User, swid string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	isJs := (r.PostFormValue("js") == "1")
	if !user.Perms.ManageThemes {
		return common.NoPermissionsJSQ(w, r, user, isJs)
	}

	wid, err := strconv.Atoi(swid)
	if err != nil {
		return common.LocalErrorJSQ(phrases.GetErrorPhrase("id_must_be_integer"), w, r, user, isJs)
	}
	widget, err := common.Widgets.Get(wid)
	if err == sql.ErrNoRows {
		return common.NotFound(w, r, nil)
	} else if err != nil {
		return common.InternalError(err, w, r)
	}

	err = widget.Delete()
	if err != nil {
		return common.InternalError(err, w, r)
	}

	return successRedirect("/panel/themes/widgets/", w, r, isJs)
}
