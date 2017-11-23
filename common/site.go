package common

import (
	"errors"
	"net/http"
	"strings"
)

// Site holds the basic settings which should be tweaked when setting up a site, we might move them to the settings table at some point
var Site = &site{Name: "Magical Fairy Land", Language: "english"}

// DbConfig holds the database configuration
var DbConfig = dbConfig{Host: "localhost"}

// Config holds the more technical settings
var Config config

// Dev holds build flags and other things which should only be modified during developers or to gather additional test data
var Dev devConfig

type site struct {
	ShortName    string
	Name         string
	Email        string
	URL          string
	Port         string
	EnableSsl    bool
	EnableEmails bool
	HasProxy     bool
	Language     string
}

type dbConfig struct {
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

type config struct {
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

type devConfig struct {
	DebugMode     bool
	SuperDebug    bool
	TemplateDebug bool
	Profiling     bool
	TestDB        bool
}

func ProcessConfig() error {
	Config.Noavatar = strings.Replace(Config.Noavatar, "{site_url}", Site.URL, -1)
	if Site.Port != "80" && Site.Port != "443" {
		Site.URL = strings.TrimSuffix(Site.URL, "/")
		Site.URL = strings.TrimSuffix(Site.URL, "\\")
		Site.URL = strings.TrimSuffix(Site.URL, ":")
		Site.URL = Site.URL + ":" + Site.Port
	}
	// We need this in here rather than verifyConfig as switchToTestDB() currently overwrites the values it verifies
	if DbConfig.TestDbname == DbConfig.Dbname {
		return errors.New("Your test database can't have the same name as your production database")
	}
	if Dev.TestDB {
		SwitchToTestDB()
	}
	return nil
}

func VerifyConfig() error {
	if !Forums.Exists(Config.DefaultForum) {
		return errors.New("Invalid default forum")
	}
	return nil
}

func SwitchToTestDB() {
	DbConfig.Host = DbConfig.TestHost
	DbConfig.Username = DbConfig.TestUsername
	DbConfig.Password = DbConfig.TestPassword
	DbConfig.Dbname = DbConfig.TestDbname
	DbConfig.Port = DbConfig.TestPort
}
