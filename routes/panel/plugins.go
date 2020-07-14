package panel

import (
	"errors"
	"log"
	"net/http"

	c "github.com/Azareal/Gosora/common"
)

func Plugins(w http.ResponseWriter, r *http.Request, u *c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, u, "plugins", "plugins")
	if ferr != nil {
		return ferr
	}
	if !u.Perms.ManagePlugins {
		return c.NoPermissions(w, r, u)
	}

	var pluginList []interface{}
	for _, plugin := range c.Plugins {
		pluginList = append(pluginList, plugin)
	}

	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "", "", "panel_plugins", c.PanelPage{basePage, pluginList, nil}})
}

// TODO: Abstract more of the plugin activation / installation / deactivation logic, so we can test all that more reliably and easily
func PluginsActivate(w http.ResponseWriter, r *http.Request, u *c.User, uname string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}
	if !u.Perms.ManagePlugins {
		return c.NoPermissions(w, r, u)
	}

	pl, ok := c.Plugins[uname]
	if !ok {
		return c.LocalError("The plugin isn't registered in the system", w, r, u)
	}
	if pl.Installable && !pl.Installed {
		return c.LocalError("You can't activate this plugin without installing it first", w, r, u)
	}

	active, err := pl.BypassActive()
	hasPlugin, err2 := pl.InDatabase()
	if err != nil || err2 != nil {
		return c.InternalError(err, w, r)
	}

	if pl.Activate != nil {
		err = pl.Activate(pl)
		if err != nil {
			return c.LocalError(err.Error(), w, r, u)
		}
	}

	if hasPlugin {
		if active {
			return c.LocalError("The plugin is already active", w, r, u)
		}
		err = pl.SetActive(true)
	} else {
		err = pl.AddToDatabase(true, false)
	}
	if err != nil {
		return c.InternalError(err, w, r)
	}

	log.Printf("Activating plugin '%s'", pl.Name)
	err = pl.Init(pl)
	if err != nil {
		return c.LocalError(err.Error(), w, r, u)
	}
	err = c.AdminLogs.CreateExtra("activate", 0, "plugin", u.GetIP(), u.ID, c.SanitiseSingleLine(pl.Name))
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/plugins/", http.StatusSeeOther)
	return nil
}

func PluginsDeactivate(w http.ResponseWriter, r *http.Request, u *c.User, uname string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}
	if !u.Perms.ManagePlugins {
		return c.NoPermissions(w, r, u)
	}

	pl, ok := c.Plugins[uname]
	if !ok {
		return c.LocalError("The plugin isn't registered in the system", w, r, u)
	}
	log.Printf("plugin: %+v\n", pl)

	active, err := pl.BypassActive()
	if err != nil {
		return c.InternalError(err, w, r)
	} else if !active {
		return c.LocalError("The plugin you're trying to deactivate isn't active", w, r, u)
	}

	err = pl.SetActive(false)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	if pl.Deactivate != nil {
		pl.Deactivate(pl)
	}
	err = c.AdminLogs.CreateExtra("deactivate", 0, "plugin", u.GetIP(), u.ID, c.SanitiseSingleLine(pl.Name))
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/plugins/", http.StatusSeeOther)
	return nil
}

func PluginsInstall(w http.ResponseWriter, r *http.Request, u *c.User, uname string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, u)
	if ferr != nil {
		return ferr
	}
	if !u.Perms.ManagePlugins {
		return c.NoPermissions(w, r, u)
	}

	pl, ok := c.Plugins[uname]
	if !ok {
		return c.LocalError("The plugin isn't registered in the system", w, r, u)
	}
	if !pl.Installable {
		return c.LocalError("This plugin is not installable", w, r, u)
	}
	if pl.Installed {
		return c.LocalError("This plugin has already been installed", w, r, u)
	}

	active, err := pl.BypassActive()
	hasPlugin, err2 := pl.InDatabase()
	if err != nil || err2 != nil {
		return c.InternalError(err, w, r)
	}
	if active {
		return c.InternalError(errors.New("An uninstalled plugin is still active"), w, r)
	}

	if pl.Install != nil {
		err = pl.Install(pl)
		if err != nil {
			return c.LocalError(err.Error(), w, r, u)
		}
	}

	if pl.Activate != nil {
		err = pl.Activate(pl)
		if err != nil {
			return c.LocalError(err.Error(), w, r, u)
		}
	}

	if hasPlugin {
		err = pl.SetInstalled(true)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		err = pl.SetActive(true)
	} else {
		err = pl.AddToDatabase(true, true)
	}
	if err != nil {
		return c.InternalError(err, w, r)
	}

	log.Printf("Installing plugin '%s'", pl.Name)
	err = pl.Init(pl)
	if err != nil {
		return c.LocalError(err.Error(), w, r, u)
	}
	err = c.AdminLogs.CreateExtra("install", 0, "plugin", u.GetIP(), u.ID, c.SanitiseSingleLine(pl.Name))
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/plugins/", http.StatusSeeOther)
	return nil
}
