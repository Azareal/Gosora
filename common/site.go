package common

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

// Site holds the basic settings which should be tweaked when setting up a site, we might move them to the settings table at some point
var Site = &site{Name: "Magical Fairy Land", Language: "english"}

// DbConfig holds the database configuration
var DbConfig = &dbConfig{Host: "localhost"}

// Config holds the more technical settings
var Config = new(config)

// Dev holds build flags and other things which should only be modified during developers or to gather additional test data
var Dev = new(devConfig)

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

	MaxRequestSize int // Alias, do not modify, will be overwritten
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
	HashAlgo     string // Defaults to bcrypt, and in the future, possibly something stronger

	MaxRequestSizeStr  string
	MaxRequestSize     int
	UserCache          string
	UserCacheCapacity  int
	TopicCache         string
	TopicCacheCapacity int

	SMTPServer   string
	SMTPUsername string
	SMTPPassword string
	SMTPPort     string
	//SMTPEnableTLS bool

	Search string

	DefaultPath     string
	DefaultGroup    int    // Should be a setting in the database
	ActivationGroup int    // Should be a setting in the database
	StaffCSS        string // ? - Move this into the settings table? Might be better to implement this as Group CSS
	DefaultForum    int    // The forum posts go in by default, this used to be covered by the Uncategorised Forum, but we want to replace it with a more robust solution. Make this a setting?
	MinifyTemplates bool
	BuildSlugs      bool // TODO: Make this a setting?
	ServerCount     int

	DisableLiveTopicList bool
	DisableJSAntispam    bool

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
	NoFsnotify    bool // Super Experimental!
}

// configHolder is purely for having a big struct to unmarshal data into
type configHolder struct {
	Site     *site
	Config   *config
	Database *dbConfig
	Dev      *devConfig
}

func LoadConfig() error {
	data, err := ioutil.ReadFile("./config/config.json")
	if err != nil {
		return err
	}

	var config configHolder
	err = json.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	Site = config.Site
	Config = config.Config
	DbConfig = config.Database
	Dev = config.Dev

	return nil
}

func ProcessConfig() (err error) {
	Config.Noavatar = strings.Replace(Config.Noavatar, "{site_url}", Site.URL, -1)
	guestAvatar = GuestAvatar{buildNoavatar(0, 200), buildNoavatar(0, 48)}
	Site.Host = Site.URL
	if Site.Port != "80" && Site.Port != "443" {
		Site.URL = strings.TrimSuffix(Site.URL, "/")
		Site.URL = strings.TrimSuffix(Site.URL, "\\")
		Site.URL = strings.TrimSuffix(Site.URL, ":")
		Site.URL = Site.URL + ":" + Site.Port
	}
	if Config.DefaultPath == "" {
		Config.DefaultPath = "/topics/"
	}

	// TODO: Bump the size of max request size up, if it's too low
	Config.MaxRequestSize, err = strconv.Atoi(Config.MaxRequestSizeStr)
	if err != nil {
		reqSizeStr := Config.MaxRequestSizeStr
		if len(reqSizeStr) < 3 {
			return errors.New("Invalid unit for MaxRequestSizeStr")
		}

		quantity, err := strconv.Atoi(reqSizeStr[:len(reqSizeStr)-2])
		if err != nil {
			return errors.New("Unable to convert quantity to integer in MaxRequestSizeStr, found " + reqSizeStr[:len(reqSizeStr)-2])
		}
		unit := reqSizeStr[len(reqSizeStr)-2:]

		// TODO: Make it a named error just in case new errors are added in here in the future
		Config.MaxRequestSize, err = FriendlyUnitToBytes(quantity, unit)
		if err != nil {
			return errors.New("Unable to recognise unit for MaxRequestSizeStr, found " + unit)
		}
	}
	if Dev.DebugMode {
		log.Print("Set MaxRequestSize to ", Config.MaxRequestSize)
	}
	if Config.MaxRequestSize <= 0 {
		log.Fatal("MaxRequestSize should not be zero or below")
	}
	Site.MaxRequestSize = Config.MaxRequestSize

	// ? Find a way of making these unlimited if zero? It might rule out some optimisations, waste memory, and break layouts
	if Config.MaxTopicTitleLength == 0 {
		Config.MaxTopicTitleLength = 100
	}
	if Config.MaxUsernameLength == 0 {
		Config.MaxUsernameLength = 100
	}
	GuestUser.Avatar, GuestUser.MicroAvatar = BuildAvatar(0, "")

	if Config.HashAlgo != "" {
		// TODO: Set the alternate hash algo, e.g. argon2
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
