package main

import c "github.com/Azareal/Gosora/common"
import "github.com/oschwald/geoip2-golang"

var geoipDB *geoip.DB
var geoipDBLocation = "geoip_db.mmdb"

func init() {
	c.Plugins.Add(&c.Plugin{UName: "geoip", Name: "Geoip", Author: "Azareal", Init: initGeoip, Deactivate: deactivateGeoip})
}

func initGeoip(plugin *c.Plugin) (err error) {
	geoipDB, err = geoip2.Open(geoipDBLocation)
	return err
}

func deactivateGeoip(plugin *c.Plugin) {
	geoipDB.Close()
}
