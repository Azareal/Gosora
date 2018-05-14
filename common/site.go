package common

import (
	"errors"
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
	Host         string
	Port         string
	EnableSsl    bool
	EnableEmails bool
	HasProxy     bool
	Language     string
}

type dbConfig struct {
	// Production database
	Adapter  string
	Host     string
	Username string
	Password string
	Dbname   string
	Port     string

	// Test database. Split this into a separate variable?
	TestAdapter  string
	TestHost     string
	TestUsername string
	TestPassword string
	TestDbname   string
	TestPort     string
}

type config struct {
	SslPrivkey   string
	SslFullchain string

	MaxRequestSize     int64
	CacheTopicUser     int
	UserCacheCapacity  int
	TopicCacheCapacity int

	SMTPServer   string
	SMTPUsername string
	SMTPPassword string
	SMTPPort     string
	//SMTPEnableTLS bool

	DefaultRoute    string
	DefaultGroup    int
	ActivationGroup int
	StaffCSS        string // ? - Move this into the settings table? Might be better to implement this as Group CSS
	DefaultForum    int    // The forum posts go in by default, this used to be covered by the Uncategorised Forum, but we want to replace it with a more robust solution. Make this a setting?
	MinifyTemplates bool
	BuildSlugs      bool // TODO: Make this a setting?
	ServerCount     int

	Noavatar            string // ? - Move this into the settings table?
	ItemsPerPage        int    // ? - Move this into the settings table?
	MaxTopicTitleLength int
	MaxUsernameLength   int
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
	Site.Host = Site.URL
	if Site.Port != "80" && Site.Port != "443" {
		Site.URL = strings.TrimSuffix(Site.URL, "/")
		Site.URL = strings.TrimSuffix(Site.URL, "\\")
		Site.URL = strings.TrimSuffix(Site.URL, ":")
		Site.URL = Site.URL + ":" + Site.Port
	}

	// ? Find a way of making these unlimited if zero? It might rule out some optimisations, waste memory, and break layouts
	if Config.MaxTopicTitleLength == 0 {
		Config.MaxTopicTitleLength = 100
	}
	if Config.MaxUsernameLength == 0 {
		Config.MaxUsernameLength = 100
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
	if Config.ServerCount < 1 {
		return errors.New("You can't have less than one server")
	}
	if Config.MaxTopicTitleLength > 100 {
		return errors.New("The max topic title length cannot be over 100 as that's unable to fit in the database row")
	}
	if Config.MaxUsernameLength > 100 {
		return errors.New("The max username length cannot be over 100 as that's unable to fit in the database row")
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
