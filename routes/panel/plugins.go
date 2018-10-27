package panel

import (
	"errors"
	"log"
	"net/http"

	"github.com/Azareal/Gosora/common"
)

func Plugins(w http.ResponseWriter, r *http.Request, user common.User) common.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "plugins", "plugins")
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManagePlugins {
		return common.NoPermissions(w, r, user)
	}

	var pluginList []interface{}
	for _, plugin := range common.Plugins {
		pluginList = append(pluginList, plugin)
	}

	pi := common.PanelPage{basePage, pluginList, nil}
	return renderTemplate("panel_plugins", w, r, user, &pi)
}

// TODO: Abstract more of the plugin activation / installation / deactivation logic, so we can test all that more reliably and easily
func PluginsActivate(w http.ResponseWriter, r *http.Request, user common.User, uname string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManagePlugins {
		return common.NoPermissions(w, r, user)
	}

	plugin, ok := common.Plugins[uname]
	if !ok {
		return common.LocalError("The plugin isn't registered in the system", w, r, user)
	}
	if plugin.Installable && !plugin.Installed {
		return common.LocalError("You can't activate this plugin without installing it first", w, r, user)
	}

	active, err := plugin.BypassActive()
	hasPlugin, err2 := plugin.InDatabase()
	if err != nil || err2 != nil {
		return common.InternalError(err, w, r)
	}

	if plugin.Activate != nil {
		err = plugin.Activate()
		if err != nil {
			return common.LocalError(err.Error(), w, r, user)
		}
	}

	if hasPlugin {
		if active {
			return common.LocalError("The plugin is already active", w, r, user)
		}
		err = plugin.SetActive(true)
	} else {
		err = plugin.AddToDatabase(true, false)
	}
	if err != nil {
		return common.InternalError(err, w, r)
	}

	log.Printf("Activating plugin '%s'", plugin.Name)
	err = plugin.Init()
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	http.Redirect(w, r, "/panel/plugins/", http.StatusSeeOther)
	return nil
}

func PluginsDeactivate(w http.ResponseWriter, r *http.Request, user common.User, uname string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManagePlugins {
		return common.NoPermissions(w, r, user)
	}

	plugin, ok := common.Plugins[uname]
	if !ok {
		return common.LocalError("The plugin isn't registered in the system", w, r, user)
	}
	log.Printf("plugin: %+v\n", plugin)

	active, err := plugin.BypassActive()
	if err != nil {
		return common.InternalError(err, w, r)
	} else if !active {
		return common.LocalError("The plugin you're trying to deactivate isn't active", w, r, user)
	}

	err = plugin.SetActive(false)
	if err != nil {
		return common.InternalError(err, w, r)
	}
	if plugin.Deactivate != nil {
		plugin.Deactivate()
	}

	http.Redirect(w, r, "/panel/plugins/", http.StatusSeeOther)
	return nil
}

func PluginsInstall(w http.ResponseWriter, r *http.Request, user common.User, uname string) common.RouteError {
	_, ferr := common.SimplePanelUserCheck(w, r, &user)
	if ferr != nil {
		return ferr
	}
	if !user.Perms.ManagePlugins {
		return common.NoPermissions(w, r, user)
	}

	plugin, ok := common.Plugins[uname]
	if !ok {
		return common.LocalError("The plugin isn't registered in the system", w, r, user)
	}
	if !plugin.Installable {
		return common.LocalError("This plugin is not installable", w, r, user)
	}
	if plugin.Installed {
		return common.LocalError("This plugin has already been installed", w, r, user)
	}

	active, err := plugin.BypassActive()
	hasPlugin, err2 := plugin.InDatabase()
	if err != nil || err2 != nil {
		return common.InternalError(err, w, r)
	}
	if active {
		return common.InternalError(errors.New("An uninstalled plugin is still active"), w, r)
	}

	if plugin.Install != nil {
		err = plugin.Install()
		if err != nil {
			return common.LocalError(err.Error(), w, r, user)
		}
	}

	if plugin.Activate != nil {
		err = plugin.Activate()
		if err != nil {
			return common.LocalError(err.Error(), w, r, user)
		}
	}

	if hasPlugin {
		err = plugin.SetInstalled(true)
		if err != nil {
			return common.InternalError(err, w, r)
		}
		err = plugin.SetActive(true)
	} else {
		err = plugin.AddToDatabase(true, true)
	}
	if err != nil {
		return common.InternalError(err, w, r)
	}

	log.Printf("Installing plugin '%s'", plugin.Name)
	err = plugin.Init()
	if err != nil {
		return common.LocalError(err.Error(), w, r, user)
	}

	http.Redirect(w, r, "/panel/plugins/", http.StatusSeeOther)
	return nil
}
