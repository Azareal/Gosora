package main

import "github.com/oschwald/geoip2-golang"

var geoip_db *geoip.DB
var geoip_db_location string = "geoip_db.mmdb"

func init() {
	plugins["geoip"] = NewPlugin("geoip","Geoip","Azareal","http://github.com/Azareal","","","",init_geoip,nil,deactivate_geoip,nil,nil)
}

func init_geoip() (err error) {
	geoip_db, err = geoip2.Open(geoip_db_location)
	return err
}

func deactivate_geoip() {
	geoip_db.Close()
}
