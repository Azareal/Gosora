// WIP - Experimental adventure plugin, this might find a new home soon, but it's here to stress test Gosora's extensibility for now
package main

import "github.com/Azareal/Gosora/common"

func init() {
	common.Plugins.Add(&common.Plugin{
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

func initAdventure() error {
	return nil
}

// TODO: Change the signature to return an error?
func deactivateAdventure() {
}

func installAdventure() error {
	return nil
}
