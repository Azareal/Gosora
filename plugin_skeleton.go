package main

import "github.com/Azareal/Gosora/common"

func init() {
	/*
		The UName field should match the name in the URL minus plugin_ and the file extension. The same name as the map index. Please choose a unique name which won't clash with any other plugins.

		The Name field is for the friendly name of the plugin shown to the end-user.

		The Author field is the author of this plugin. The one who created it.

		The URL field is for the URL pointing to the location where you can download this plugin.

		The Settings field points to the route for managing the settings for this plugin. Coming soon.

		The Tag field is for providing a tiny snippet of information separate from the description.

		The Type field is for the type of the plugin. This gets changed to "go" automatically and we would suggest leaving "".

		The Init field is for the initialisation handler which is called by the software to run this plugin. This expects a function. You should add your hooks, init logic, initial queries, etc. in said function.

		The Activate field is for the handler which is called by the software when the admin hits the Activate button in the control panel. This is separate from the Init handler which is called upon the start of the server and upon activation. Use nil if you don't have a handler for this.

		The Deactivate field is for the handler which is called by the software when the admin hits the Deactivate button in the control panel. You should clean-up any resources you have allocated, remove any hooks, close any statements, etc. within this handler.

		The Installation field is for one-off installation logic such as creating tables. You will need to run the separate uninstallation function for that.

		That Uninstallation field which is currently unused is for not only deactivating this plugin, but for purging any data associated with it such a new tables or data produced by the end-user.
	*/
	common.Plugins.Add(&common.Plugin{UName: "skeleton", Name: "Skeleton", Author: "Azareal", Init: initSkeleton, Activate: activateSkeleton, Deactivate: deactivateSkeleton})
}

func initSkeleton(plugin *common.Plugin) error { return nil }

// Any errors encountered while trying to activate the plugin are reported back to the admin and the activation is aborted
func activateSkeleton(plugin *common.Plugin) error { return nil }

func deactivateSkeleton(plugin *common.Plugin) {}
