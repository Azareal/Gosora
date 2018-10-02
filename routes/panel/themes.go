package panel

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"../../common"
)

//routePanelThemes
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

//routePanelThemesSetDefault
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

//routePanelThemesMenus
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
			name = common.GetTmplPhrase("panel_themes_menus_main")
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

//routePanelThemesMenusEdit
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
		return common.LocalError(common.GetErrorPhrase("url_id_must_be_integer"), w, r, user)
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

//routePanelThemesMenuItemEdit
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
		return common.LocalError(common.GetErrorPhrase("url_id_must_be_integer"), w, r, user)
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

//routePanelThemesMenuItemEditSubmit
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
		return common.LocalErrorJSQ(common.GetErrorPhrase("id_must_be_integer"), w, r, user, isJs)
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

//routePanelThemesMenuItemCreateSubmit
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
		return common.LocalErrorJSQ(common.GetErrorPhrase("id_must_be_integer"), w, r, user, isJs)
	}

	menuItem := common.MenuItem{MenuID: menuID}
	menuItem = themesMenuItemSetters(r, menuItem)
	itemID, err := menuItem.Create()
	if err != nil {
		return common.InternalErrorJSQ(err, w, r, isJs)
	}
	return successRedirect("/panel/themes/menus/item/edit/"+strconv.Itoa(itemID), w, r, isJs)
}

//routePanelThemesMenuItemDeleteSubmit
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
		return common.LocalErrorJSQ(common.GetErrorPhrase("id_must_be_integer"), w, r, user, isJs)
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

//routePanelThemesMenuItemOrderSubmit
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
		return common.LocalErrorJSQ(common.GetErrorPhrase("id_must_be_integer"), w, r, user, isJs)
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
