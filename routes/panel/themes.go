package panel

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
	p "github.com/Azareal/Gosora/common/phrases"
)

func Themes(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "themes", "themes")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return c.NoPermissions(w, r, user)
	}

	var pThemeList, vThemeList []*c.Theme
	for _, theme := range c.Themes {
		if theme.HideFromThemes {
			continue
		}
		if theme.ForkOf == "" {
			pThemeList = append(pThemeList, theme)
		} else {
			vThemeList = append(vThemeList, theme)
		}
	}

	pi := c.PanelThemesPage{basePage, pThemeList, vThemeList}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "panel_themes", "", "panel_themes", &pi})
}

func ThemesSetDefault(w http.ResponseWriter, r *http.Request, user c.User, uname string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return c.NoPermissions(w, r, user)
	}

	theme, ok := c.Themes[uname]
	if !ok {
		return c.LocalError("The theme isn't registered in the system", w, r, user)
	}
	if theme.Disabled {
		return c.LocalError("You must not enable this theme", w, r, user)
	}

	err := c.UpdateDefaultTheme(theme)
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/themes/", http.StatusSeeOther)
	return nil
}

func ThemesMenus(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "themes_menus", "themes")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return c.NoPermissions(w, r, user)
	}

	var menuList []c.PanelMenuListItem
	for mid, list := range c.Menus.GetAllMap() {
		name := ""
		if mid == 1 {
			name = p.GetTmplPhrase("panel_themes_menus_main")
		}
		menuList = append(menuList, c.PanelMenuListItem{
			Name:      name,
			ID:        mid,
			ItemCount: len(list.List),
		})
	}

	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "", "", "panel_themes_menus", &c.PanelMenuListPage{basePage, menuList}})
}

func ThemesMenusEdit(w http.ResponseWriter, r *http.Request, user c.User, smid string) c.RouteError {
	// TODO: Something like Menu #1 for the title?
	basePage, ferr := buildBasePage(w, r, &user, "themes_menus_edit", "themes")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return c.NoPermissions(w, r, user)
	}
	basePage.Header.AddScript("Sortable-1.4.0/Sortable.min.js")
	basePage.Header.AddScriptAsync("panel_menu_items.js")

	mid, err := strconv.Atoi(smid)
	if err != nil {
		return c.LocalError(p.GetErrorPhrase("url_id_must_be_integer"), w, r, user)
	}

	menuHold, err := c.Menus.Get(mid)
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, basePage.Header)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	var menuList []c.MenuItem
	for _, item := range menuHold.List {
		menuTmpls := map[string]c.MenuTmpl{
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

	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "", "", "panel_themes_menus_items", &c.PanelMenuPage{basePage, mid, menuList}})
}

func ThemesMenuItemEdit(w http.ResponseWriter, r *http.Request, user c.User, sitemID string) c.RouteError {
	// TODO: Something like Menu #1 for the title?
	basePage, ferr := buildBasePage(w, r, &user, "themes_menus_edit", "themes")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return c.NoPermissions(w, r, user)
	}

	itemID, err := strconv.Atoi(sitemID)
	if err != nil {
		return c.LocalError(p.GetErrorPhrase("url_id_must_be_integer"), w, r, user)
	}

	menuItem, err := c.Menus.ItemStore().Get(itemID)
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, basePage.Header)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "", "", "panel_themes_menus_item_edit", &c.PanelMenuItemPage{basePage, menuItem}})
}

func themesMenuItemSetters(r *http.Request, i c.MenuItem) c.MenuItem {
	getItem := func(name string) string {
		return c.SanitiseSingleLine(r.PostFormValue("item-" + name))
	}
	i.Name = getItem("name")
	i.HTMLID = getItem("htmlid")
	i.CSSClass = getItem("cssclass")
	i.Position = getItem("position")
	if i.Position != "left" && i.Position != "right" {
		i.Position = "left"
	}
	i.Path = getItem("path")
	i.Aria = getItem("aria")
	i.Tooltip = getItem("tooltip")
	i.TmplName = getItem("tmplname")

	switch getItem("permissions") {
	case "everyone":
		i.GuestOnly = false
		i.MemberOnly = false
		i.SuperModOnly = false
		i.AdminOnly = false
	case "guest-only":
		i.GuestOnly = true
		i.MemberOnly = false
		i.SuperModOnly = false
		i.AdminOnly = false
	case "member-only":
		i.GuestOnly = false
		i.MemberOnly = true
		i.SuperModOnly = false
		i.AdminOnly = false
	case "supermod-only":
		i.GuestOnly = false
		i.MemberOnly = true
		i.SuperModOnly = true
		i.AdminOnly = false
	case "admin-only":
		i.GuestOnly = false
		i.MemberOnly = true
		i.SuperModOnly = true
		i.AdminOnly = true
	}
	return i
}

func ThemesMenuItemEditSubmit(w http.ResponseWriter, r *http.Request, user c.User, sitemID string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	js := r.PostFormValue("js") == "1"
	if !user.Perms.ManageThemes {
		return c.NoPermissionsJSQ(w, r, user, js)
	}

	itemID, err := strconv.Atoi(sitemID)
	if err != nil {
		return c.LocalErrorJSQ(p.GetErrorPhrase("id_must_be_integer"), w, r, user, js)
	}

	menuItem, err := c.Menus.ItemStore().Get(itemID)
	if err == sql.ErrNoRows {
		return c.LocalErrorJSQ("This item doesn't exist.", w, r, user, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}
	//menuItem = menuItem.Copy() // If we switch this for a pointer, we might need this as a scratchpad
	menuItem = themesMenuItemSetters(r, menuItem)

	err = menuItem.Commit()
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}
	return successRedirect("/panel/themes/menus/item/edit/"+strconv.Itoa(itemID), w, r, js)
}

func ThemesMenuItemCreateSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	js := r.PostFormValue("js") == "1"
	if !user.Perms.ManageThemes {
		return c.NoPermissionsJSQ(w, r, user, js)
	}
	smenuID := r.PostFormValue("mid")
	if smenuID == "" {
		return c.LocalErrorJSQ("No menuID provided", w, r, user, js)
	}
	menuID, err := strconv.Atoi(smenuID)
	if err != nil {
		return c.LocalErrorJSQ(p.GetErrorPhrase("id_must_be_integer"), w, r, user, js)
	}

	menuItem := c.MenuItem{MenuID: menuID}
	menuItem = themesMenuItemSetters(r, menuItem)
	itemID, err := menuItem.Create()
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}
	return successRedirect("/panel/themes/menus/item/edit/"+strconv.Itoa(itemID), w, r, js)
}

func ThemesMenuItemDeleteSubmit(w http.ResponseWriter, r *http.Request, user c.User, sitemID string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	js := r.PostFormValue("js") == "1"
	if !user.Perms.ManageThemes {
		return c.NoPermissionsJSQ(w, r, user, js)
	}

	itemID, err := strconv.Atoi(sitemID)
	if err != nil {
		return c.LocalErrorJSQ(p.GetErrorPhrase("id_must_be_integer"), w, r, user, js)
	}
	menuItem, err := c.Menus.ItemStore().Get(itemID)
	if err == sql.ErrNoRows {
		return c.LocalErrorJSQ("This item doesn't exist.", w, r, user, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}
	//menuItem = menuItem.Copy() // If we switch this for a pointer, we might need this as a scratchpad

	err = menuItem.Delete()
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}
	return successRedirect("/panel/themes/menus/", w, r, js)
}

func ThemesMenuItemOrderSubmit(w http.ResponseWriter, r *http.Request, user c.User, smid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	js := r.PostFormValue("js") == "1"
	if !user.Perms.ManageThemes {
		return c.NoPermissionsJSQ(w, r, user, js)
	}

	mid, err := strconv.Atoi(smid)
	if err != nil {
		return c.LocalErrorJSQ(p.GetErrorPhrase("id_must_be_integer"), w, r, user, js)
	}
	menuHold, err := c.Menus.Get(mid)
	if err == sql.ErrNoRows {
		return c.LocalErrorJSQ("Can't find menu", w, r, user, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	sitems := strings.TrimSuffix(strings.TrimPrefix(r.PostFormValue("items"), "{"), "}")
	//fmt.Printf("sitems: %+v\n", sitems)

	updateMap := make(map[int]int)
	for index, smiid := range strings.Split(sitems, ",") {
		miid, err := strconv.Atoi(smiid)
		if err != nil {
			return c.LocalErrorJSQ("Invalid integer in menu item list", w, r, user, js)
		}
		updateMap[miid] = index
	}
	menuHold.UpdateOrder(updateMap)

	return successRedirect("/panel/themes/menus/edit/"+strconv.Itoa(mid), w, r, js)
}

func ThemesWidgets(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "themes_widgets", "themes")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return c.NoPermissions(w, r, user)
	}
	basePage.Header.AddScript("widgets.js")

	docks := make(map[string][]c.WidgetEdit)
	for _, name := range c.GetDockList() {
		if name == "leftOfNav" || name == "rightOfNav" {
			continue
		}
		var widgets []c.WidgetEdit
		for _, widget := range c.GetDock(name) {
			data := make(map[string]string)
			err := json.Unmarshal([]byte(widget.RawBody), &data)
			if err != nil {
				return c.InternalError(err, w, r)
			}
			widgets = append(widgets, c.WidgetEdit{widget, data})
		}
		docks[name] = widgets
	}

	pi := c.PanelWidgetListPage{basePage, docks, c.WidgetEdit{&c.Widget{ID: 0, Type: "simple"}, make(map[string]string)}}
	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "", "", "panel_themes_widgets", pi})
}

func widgetsParseInputs(r *http.Request, widget *c.Widget) (*c.WidgetEdit, error) {
	data := make(map[string]string)
	widget.Enabled = r.FormValue("wenabled") == "1"
	widget.Location = r.FormValue("wlocation")
	if widget.Location == "" {
		return nil, errors.New("You need to specify a location for this widget.")
	}
	widget.Side = r.FormValue("wside")
	if !c.HasDock(widget.Side) {
		return nil, errors.New("The widget dock you specified doesn't exist.")
	}

	wtype := r.FormValue("wtype")
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
	case "wol", "wol_context", "search_and_filter":
		widget.Type = wtype // ? - Are we sure we should be directly assigning user provided data even if it's validated?
	default:
		return nil, errors.New("Unknown widget type")
	}

	return &c.WidgetEdit{widget, data}, nil
}

// ThemesWidgetsEditSubmit is an action which is triggered when someone sends an update request for a widget
func ThemesWidgetsEditSubmit(w http.ResponseWriter, r *http.Request, user c.User, swid string) c.RouteError {
	//fmt.Println("in ThemesWidgetsEditSubmit")
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	js := r.PostFormValue("js") == "1"
	if !user.Perms.ManageThemes {
		return c.NoPermissionsJSQ(w, r, user, js)
	}

	wid, err := strconv.Atoi(swid)
	if err != nil {
		return c.LocalErrorJSQ(p.GetErrorPhrase("id_must_be_integer"), w, r, user, js)
	}

	widget, err := c.Widgets.Get(wid)
	if err == sql.ErrNoRows {
		return c.NotFoundJSQ(w, r, nil, js)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	ewidget, err := widgetsParseInputs(r, widget.Copy())
	if err != nil {
		return c.LocalErrorJSQ(err.Error(), w, r, user, js)
	}

	err = ewidget.Commit()
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	return successRedirect("/panel/themes/widgets/", w, r, js)
}

// ThemesWidgetsCreateSubmit is an action which is triggered when someone sends a create request for a widget
func ThemesWidgetsCreateSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	//fmt.Println("in ThemesWidgetsCreateSubmit")
	js := r.PostFormValue("js") == "1"
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return c.NoPermissionsJSQ(w, r, user, js)
	}

	ewidget, err := widgetsParseInputs(r, &c.Widget{})
	if err != nil {
		return c.LocalErrorJSQ(err.Error(), w, r, user, js)
	}

	err = ewidget.Create()
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, js)
	}

	return successRedirect("/panel/themes/widgets/", w, r, js)
}

func ThemesWidgetsDeleteSubmit(w http.ResponseWriter, r *http.Request, user c.User, swid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	js := r.PostFormValue("js") == "1"
	if !user.Perms.ManageThemes {
		return c.NoPermissionsJSQ(w, r, user, js)
	}

	wid, err := strconv.Atoi(swid)
	if err != nil {
		return c.LocalErrorJSQ(p.GetErrorPhrase("id_must_be_integer"), w, r, user, js)
	}
	widget, err := c.Widgets.Get(wid)
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, nil)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	err = widget.Delete()
	if err != nil {
		return c.InternalError(err, w, r)
	}

	return successRedirect("/panel/themes/widgets/", w, r, js)
}
