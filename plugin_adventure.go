// WIP - Experimental adventure plugin, this might find a new home soon, but it's here to stress test Gosora's extensibility for now
package main

import "./common"

func init() {
	common.Plugins["adventure"] = common.NewPlugin("adventure", "WIP", "Azareal", "http://github.com/Azareal", "", "", "", initAdventure, nil, deactivateAdventure, installAdventure, nil)
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
