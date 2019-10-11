// WIP - Experimental adventure plugin, this might find a new home soon, but it's here to stress test Gosora's extensibility for now
package extend

import c "github.com/Azareal/Gosora/common"

func init() {
	c.Plugins.Add(&c.Plugin{
		UName:      "adventure",
		Name:       "Adventure",
		Tag:        "WIP",
		Author:     "Azareal",
		URL:        "https://github.com/Azareal",
		Init:       initAdventure,
		Deactivate: deactivateAdventure,
		Install:    installAdventure,
	})
}

func initAdventure(plugin *c.Plugin) error {
	return nil
}

// TODO: Change the signature to return an error?
func deactivateAdventure(plugin *c.Plugin) {
}

func installAdventure(plugin *c.Plugin) error {
	return nil
}
