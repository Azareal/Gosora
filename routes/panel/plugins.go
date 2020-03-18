package panel

import (
	"errors"
	"log"
	"net/http"

	c "github.com/Azareal/Gosora/common"
)

func Plugins(w http.ResponseWriter, r *http.Request, user *c.User) c.RouteError {
	basePage, ferr := buildBasePage(w, r, user, "plugins", "plugins")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManagePlugins {
		return c.NoPermissions(w, r, user)
	}

	var pluginList []interface{}
	for _, plugin := range c.Plugins {
		pluginList = append(pluginList, plugin)
	}

	return renderTemplate("panel", w, r, basePage.Header, c.Panel{basePage, "", "", "panel_plugins", c.PanelPage{basePage, pluginList, nil}})
}

// TODO: Abstract more of the plugin activation / installation / deactivation logic, so we can test all that more reliably and easily
func PluginsActivate(w http.ResponseWriter, r *http.Request, user *c.User, uname string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManagePlugins {
		return c.NoPermissions(w, r, user)
	}

	plugin, ok := c.Plugins[uname]
	if !ok {
		return c.LocalError("The plugin isn't registered in the system", w, r, user)
	}
	if plugin.Installable && !plugin.Installed {
		return c.LocalError("You can't activate this plugin without installing it first", w, r, user)
	}

	active, err := plugin.BypassActive()
	hasPlugin, err2 := plugin.InDatabase()
	if err != nil || err2 != nil {
		return c.InternalError(err, w, r)
	}

	if plugin.Activate != nil {
		err = plugin.Activate(plugin)
		if err != nil {
			return c.LocalError(err.Error(), w, r, user)
		}
	}

	if hasPlugin {
		if active {
			return c.LocalError("The plugin is already active", w, r, user)
		}
		err = plugin.SetActive(true)
	} else {
		err = plugin.AddToDatabase(true, false)
	}
	if err != nil {
		return c.InternalError(err, w, r)
	}

	log.Printf("Activating plugin '%s'", plugin.Name)
	err = plugin.Init(plugin)
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	err = c.AdminLogs.CreateExtra("activate", 0, "plugin", user.GetIP(), user.ID, c.SanitiseSingleLine(plugin.Name))
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/plugins/", http.StatusSeeOther)
	return nil
}

func PluginsDeactivate(w http.ResponseWriter, r *http.Request, user *c.User, uname string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManagePlugins {
		return c.NoPermissions(w, r, user)
	}

	plugin, ok := c.Plugins[uname]
	if !ok {
		return c.LocalError("The plugin isn't registered in the system", w, r, user)
	}
	log.Printf("plugin: %+v\n", plugin)

	active, err := plugin.BypassActive()
	if err != nil {
		return c.InternalError(err, w, r)
	} else if !active {
		return c.LocalError("The plugin you're trying to deactivate isn't active", w, r, user)
	}

	err = plugin.SetActive(false)
	if err != nil {
		return c.InternalError(err, w, r)
	}
	if plugin.Deactivate != nil {
		plugin.Deactivate(plugin)
	}
	err = c.AdminLogs.CreateExtra("deactivate", 0, "plugin", user.GetIP(), user.ID, c.SanitiseSingleLine(plugin.Name))
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/plugins/", http.StatusSeeOther)
	return nil
}

func PluginsInstall(w http.ResponseWriter, r *http.Request, user *c.User, uname string) c.RouteError {
	_, ferr := c.SimplePanelUserCheck(w, r, user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManagePlugins {
		return c.NoPermissions(w, r, user)
	}

	plugin, ok := c.Plugins[uname]
	if !ok {
		return c.LocalError("The plugin isn't registered in the system", w, r, user)
	}
	if !plugin.Installable {
		return c.LocalError("This plugin is not installable", w, r, user)
	}
	if plugin.Installed {
		return c.LocalError("This plugin has already been installed", w, r, user)
	}

	active, err := plugin.BypassActive()
	hasPlugin, err2 := plugin.InDatabase()
	if err != nil || err2 != nil {
		return c.InternalError(err, w, r)
	}
	if active {
		return c.InternalError(errors.New("An uninstalled plugin is still active"), w, r)
	}

	if plugin.Install != nil {
		err = plugin.Install(plugin)
		if err != nil {
			return c.LocalError(err.Error(), w, r, user)
		}
	}

	if plugin.Activate != nil {
		err = plugin.Activate(plugin)
		if err != nil {
			return c.LocalError(err.Error(), w, r, user)
		}
	}

	if hasPlugin {
		err = plugin.SetInstalled(true)
		if err != nil {
			return c.InternalError(err, w, r)
		}
		err = plugin.SetActive(true)
	} else {
		err = plugin.AddToDatabase(true, true)
	}
	if err != nil {
		return c.InternalError(err, w, r)
	}

	log.Printf("Installing plugin '%s'", plugin.Name)
	err = plugin.Init(plugin)
	if err != nil {
		return c.LocalError(err.Error(), w, r, user)
	}
	err = c.AdminLogs.CreateExtra("install", 0, "plugin", user.GetIP(), user.ID, c.SanitiseSingleLine(plugin.Name))
	if err != nil {
		return c.InternalError(err, w, r)
	}

	http.Redirect(w, r, "/panel/plugins/", http.StatusSeeOther)
	return nil
}
