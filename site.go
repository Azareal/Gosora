package main

import "net/http"

var site = &Site{Name: "Magical Fairy Land", Language: "english"}
var dbConfig = DBConfig{Host: "localhost"}
var config Config
var dev DevConfig

type Site struct {
	Name         string // ? - Move this into the settings table?
	Email        string // ? - Move this into the settings table?
	URL          string
	Port         string
	EnableSsl    bool
	EnableEmails bool
	HasProxy     bool
	Language     string // ? - Move this into the settings table?
}

type DBConfig struct {
	Host     string
	Username string
	Password string
	Dbname   string
	Port     string
}

type Config struct {
	SslPrivkey   string
	SslFullchain string

	MaxRequestSize     int
	CacheTopicUser     int
	UserCacheCapacity  int
	TopicCacheCapacity int

	SMTPServer   string
	SMTPUsername string
	SMTPPassword string
	SMTPPort     string

	DefaultRoute              func(http.ResponseWriter, *http.Request, User)
	DefaultGroup              int
	ActivationGroup           int
	StaffCSS                  string // ? - Move this into the settings table? Might be better to implement this as Group CSS
	UncategorisedForumVisible bool
	MinifyTemplates           bool
	MultiServer               bool

	Noavatar     string // ? - Move this into the settings table?
	ItemsPerPage int    // ? - Move this into the settings table?
}

type DevConfig struct {
	DebugMode  bool
	SuperDebug bool
	Profiling  bool
}
