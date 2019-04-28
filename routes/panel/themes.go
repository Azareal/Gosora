package panel

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	c "github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/phrases"
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
	return renderTemplate("panel_themes", w, r, basePage.Header, &pi)
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
		var name = ""
		if mid == 1 {
			name = phrases.GetTmplPhrase("panel_themes_menus_main")
		}
		menuList = append(menuList, c.PanelMenuListItem{
			Name:      name,
			ID:        mid,
			ItemCount: len(list.List),
		})
	}

	pi := c.PanelMenuListPage{basePage, menuList}
	return renderTemplate("panel_themes_menus", w, r, basePage.Header, &pi)
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
		return c.LocalError(phrases.GetErrorPhrase("url_id_must_be_integer"), w, r, user)
	}

	menuHold, err := c.Menus.Get(mid)
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, basePage.Header)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	var menuList []c.MenuItem
	for _, item := range menuHold.List {
		var menuTmpls = map[string]c.MenuTmpl{
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

	pi := c.PanelMenuPage{basePage, mid, menuList}
	return renderTemplate("panel_themes_menus_items", w, r, basePage.Header, &pi)
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
		return c.LocalError(phrases.GetErrorPhrase("url_id_must_be_integer"), w, r, user)
	}

	menuItem, err := c.Menus.ItemStore().Get(itemID)
	if err == sql.ErrNoRows {
		return c.NotFound(w, r, basePage.Header)
	} else if err != nil {
		return c.InternalError(err, w, r)
	}

	pi := c.PanelMenuItemPage{basePage, menuItem}
	return renderTemplate("panel_themes_menus_item_edit", w, r, basePage.Header, &pi)
}

func themesMenuItemSetters(r *http.Request, menuItem c.MenuItem) c.MenuItem {
	var getItem = func(name string) string {
		return c.SanitiseSingleLine(r.PostFormValue("item-" + name))
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

func ThemesMenuItemEditSubmit(w http.ResponseWriter, r *http.Request, user c.User, sitemID string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	isJs := (r.PostFormValue("js") == "1")
	if !user.Perms.ManageThemes {
		return c.NoPermissionsJSQ(w, r, user, isJs)
	}

	itemID, err := strconv.Atoi(sitemID)
	if err != nil {
		return c.LocalErrorJSQ(phrases.GetErrorPhrase("id_must_be_integer"), w, r, user, isJs)
	}

	menuItem, err := c.Menus.ItemStore().Get(itemID)
	if err == sql.ErrNoRows {
		return c.LocalErrorJSQ("This item doesn't exist.", w, r, user, isJs)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, isJs)
	}
	//menuItem = menuItem.Copy() // If we switch this for a pointer, we might need this as a scratchpad
	menuItem = themesMenuItemSetters(r, menuItem)

	err = menuItem.Commit()
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, isJs)
	}
	return successRedirect("/panel/themes/menus/item/edit/"+strconv.Itoa(itemID), w, r, isJs)
}

func ThemesMenuItemCreateSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}

	isJs := (r.PostFormValue("js") == "1")
	if !user.Perms.ManageThemes {
		return c.NoPermissionsJSQ(w, r, user, isJs)
	}
	smenuID := r.PostFormValue("mid")
	if smenuID == "" {
		return c.LocalErrorJSQ("No menuID provided", w, r, user, isJs)
	}
	menuID, err := strconv.Atoi(smenuID)
	if err != nil {
		return c.LocalErrorJSQ(phrases.GetErrorPhrase("id_must_be_integer"), w, r, user, isJs)
	}

	menuItem := c.MenuItem{MenuID: menuID}
	menuItem = themesMenuItemSetters(r, menuItem)
	itemID, err := menuItem.Create()
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, isJs)
	}
	return successRedirect("/panel/themes/menus/item/edit/"+strconv.Itoa(itemID), w, r, isJs)
}

func ThemesMenuItemDeleteSubmit(w http.ResponseWriter, r *http.Request, user c.User, sitemID string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	isJs := (r.PostFormValue("js") == "1")
	if !user.Perms.ManageThemes {
		return c.NoPermissionsJSQ(w, r, user, isJs)
	}

	itemID, err := strconv.Atoi(sitemID)
	if err != nil {
		return c.LocalErrorJSQ(phrases.GetErrorPhrase("id_must_be_integer"), w, r, user, isJs)
	}
	menuItem, err := c.Menus.ItemStore().Get(itemID)
	if err == sql.ErrNoRows {
		return c.LocalErrorJSQ("This item doesn't exist.", w, r, user, isJs)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, isJs)
	}
	//menuItem = menuItem.Copy() // If we switch this for a pointer, we might need this as a scratchpad

	err = menuItem.Delete()
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, isJs)
	}
	return successRedirect("/panel/themes/menus/", w, r, isJs)
}

func ThemesMenuItemOrderSubmit(w http.ResponseWriter, r *http.Request, user c.User, smid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	isJs := (r.PostFormValue("js") == "1")
	if !user.Perms.ManageThemes {
		return c.NoPermissionsJSQ(w, r, user, isJs)
	}

	mid, err := strconv.Atoi(smid)
	if err != nil {
		return c.LocalErrorJSQ(phrases.GetErrorPhrase("id_must_be_integer"), w, r, user, isJs)
	}
	menuHold, err := c.Menus.Get(mid)
	if err == sql.ErrNoRows {
		return c.LocalErrorJSQ("Can't find menu", w, r, user, isJs)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, isJs)
	}

	sitems := strings.TrimSuffix(strings.TrimPrefix(r.PostFormValue("items"), "{"), "}")
	//fmt.Printf("sitems: %+v\n", sitems)

	var updateMap = make(map[int]int)
	for index, smiid := range strings.Split(sitems, ",") {
		miid, err := strconv.Atoi(smiid)
		if err != nil {
			return c.LocalErrorJSQ("Invalid integer in menu item list", w, r, user, isJs)
		}
		updateMap[miid] = index
	}
	menuHold.UpdateOrder(updateMap)

	return successRedirect("/panel/themes/menus/edit/"+strconv.Itoa(mid), w, r, isJs)
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

	var docks = make(map[string][]c.WidgetEdit)
	for _, name := range c.GetDockList() {
		if name == "leftOfNav" || name == "rightOfNav" {
			continue
		}
		var widgets []c.WidgetEdit
		for _, widget := range c.GetDock(name) {
			var data = make(map[string]string)
			err := json.Unmarshal([]byte(widget.RawBody), &data)
			if err != nil {
				return c.InternalError(err, w, r)
			}
			widgets = append(widgets, c.WidgetEdit{widget, data})
		}
		docks[name] = widgets
	}

	pi := c.PanelWidgetListPage{basePage, docks, c.WidgetEdit{&c.Widget{ID: 0, Type: "simple"}, make(map[string]string)}}
	return renderTemplate("panel_themes_widgets", w, r, basePage.Header, &pi)
}

func widgetsParseInputs(r *http.Request, widget *c.Widget) (*c.WidgetEdit, error) {
	var data = make(map[string]string)
	widget.Enabled = (r.FormValue("wenabled") == "1")
	widget.Location = r.FormValue("wlocation")
	if widget.Location == "" {
		return nil, errors.New("You need to specify a location for this widget.")
	}
	widget.Side = r.FormValue("wside")
	if !c.HasDock(widget.Side) {
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
	isJs := (r.PostFormValue("js") == "1")
	if !user.Perms.ManageThemes {
		return c.NoPermissionsJSQ(w, r, user, isJs)
	}

	wid, err := strconv.Atoi(swid)
	if err != nil {
		return c.LocalErrorJSQ(phrases.GetErrorPhrase("id_must_be_integer"), w, r, user, isJs)
	}

	widget, err := c.Widgets.Get(wid)
	if err == sql.ErrNoRows {
		return c.NotFoundJSQ(w, r, nil, isJs)
	} else if err != nil {
		return c.InternalErrorJSQ(err, w, r, isJs)
	}

	ewidget, err := widgetsParseInputs(r, widget.Copy())
	if err != nil {
		return c.LocalErrorJSQ(err.Error(), w, r, user, isJs)
	}

	err = ewidget.Commit()
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, isJs)
	}

	return successRedirect("/panel/themes/widgets/", w, r, isJs)
}

// ThemesWidgetsCreateSubmit is an action which is triggered when someone sends a create request for a widget
func ThemesWidgetsCreateSubmit(w http.ResponseWriter, r *http.Request, user c.User) c.RouteError {
	//fmt.Println("in ThemesWidgetsCreateSubmit")
	isJs := (r.PostFormValue("js") == "1")
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManageThemes {
		return c.NoPermissionsJSQ(w, r, user, isJs)
	}

	ewidget, err := widgetsParseInputs(r, &c.Widget{})
	if err != nil {
		return c.LocalErrorJSQ(err.Error(), w, r, user, isJs)
	}

	err = ewidget.Create()
	if err != nil {
		return c.InternalErrorJSQ(err, w, r, isJs)
	}

	return successRedirect("/panel/themes/widgets/", w, r, isJs)
}

func ThemesWidgetsDeleteSubmit(w http.ResponseWriter, r *http.Request, user c.User, swid string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	isJs := (r.PostFormValue("js") == "1")
	if !user.Perms.ManageThemes {
		return c.NoPermissionsJSQ(w, r, user, isJs)
	}

	wid, err := strconv.Atoi(swid)
	if err != nil {
		return c.LocalErrorJSQ(phrases.GetErrorPhrase("id_must_be_integer"), w, r, user, isJs)
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

	return successRedirect("/panel/themes/widgets/", w, r, isJs)
}
