package main

import (
	"errors"
	"net/http"
	"strings"
)

var site = &Site{Name: "Magical Fairy Land", Language: "english"}
var dbConfig = DBConfig{Host: "localhost"}
var config Config
var dev DevConfig

type Site struct {
	ShortName    string
	Name         string // ? - Move this into the settings table? Should we make a second version of this for the abbreviation shown in the navbar?
	Email        string // ? - Move this into the settings table?
	URL          string
	Port         string
	EnableSsl    bool
	EnableEmails bool
	HasProxy     bool
	Language     string // ? - Move this into the settings table?
}

type DBConfig struct {
	// Production database
	Host     string
	Username string
	Password string
	Dbname   string
	Port     string

	// Test database. Split this into a separate variable?
	TestHost     string
	TestUsername string
	TestPassword string
	TestDbname   string
	TestPort     string
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

	DefaultRoute    func(http.ResponseWriter, *http.Request, User) RouteError
	DefaultGroup    int
	ActivationGroup int
	StaffCSS        string // ? - Move this into the settings table? Might be better to implement this as Group CSS
	DefaultForum    int    // The forum posts go in by default, this used to be covered by the Uncategorised Forum, but we want to replace it with a more robust solution. Make this a setting?
	MinifyTemplates bool
	MultiServer     bool

	Noavatar     string // ? - Move this into the settings table?
	ItemsPerPage int    // ? - Move this into the settings table?
}

type DevConfig struct {
	DebugMode     bool
	SuperDebug    bool
	TemplateDebug bool
	Profiling     bool
	TestDB        bool
}

func processConfig() error {
	config.Noavatar = strings.Replace(config.Noavatar, "{site_url}", site.URL, -1)
	if site.Port != "80" && site.Port != "443" {
		site.URL = strings.TrimSuffix(site.URL, "/")
		site.URL = strings.TrimSuffix(site.URL, "\\")
		site.URL = strings.TrimSuffix(site.URL, ":")
		site.URL = site.URL + ":" + site.Port
	}
	// We need this in here rather than verifyConfig as switchToTestDB() currently overwrites the values it verifies
	if dbConfig.TestDbname == dbConfig.Dbname {
		return errors.New("Your test database can't have the same name as your production database")
	}
	if dev.TestDB {
		switchToTestDB()
	}
	return nil
}

func verifyConfig() error {
	if !fstore.Exists(config.DefaultForum) {
		return errors.New("Invalid default forum")
	}
	return nil
}

func switchToTestDB() {
	dbConfig.Host = dbConfig.TestHost
	dbConfig.Username = dbConfig.TestUsername
	dbConfig.Password = dbConfig.TestPassword
	dbConfig.Dbname = dbConfig.TestDbname
	dbConfig.Port = dbConfig.TestPort
}
