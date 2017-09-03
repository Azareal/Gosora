/*
*
* Gosora Installer
* Copyright Azareal 2017 - 2018
*
 */
package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"

	"../query_gen/lib"
)

const saltLength int = 32

var db *sql.DB
var scanner *bufio.Scanner

var dbAdapter = "mysql"
var dbHost string
var dbUsername string
var dbPassword string
var dbName string
var dbPort string
var siteName, siteURL, serverPort string

var defaultAdapter = "mysql"
var defaultHost = "localhost"
var defaultUsername = "root"
var defaultDbname = "gosora"
var defaultSiteName = "Site Name"
var defaultsiteURL = "localhost"
var defaultServerPort = "80" // 8080's a good one, if you're testing and don't want it to clash with port 80

// func() error, removing type to satisfy lint
var initDatabase = _initMysql
var tableDefs = _tableDefsMysql
var initialData = _initialDataMysql

func main() {
	// Capture panics rather than immediately closing the window on Windows
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println(r)
			debug.PrintStack()
			pressAnyKey()
			return
		}
	}()

	scanner = bufio.NewScanner(os.Stdin)
	fmt.Println("Welcome to Gosora's Installer")
	fmt.Println("We're going to take you through a few steps to help you get started :)")
	if !getDatabaseDetails() {
		err := scanner.Err()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Something went wrong!")
		}
		fmt.Println("Aborting installation...")
		pressAnyKey()
		return
	}

	if !getSiteDetails() {
		err := scanner.Err()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Something went wrong!")
		}
		fmt.Println("Aborting installation...")
		pressAnyKey()
		return
	}

	err := initDatabase()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		pressAnyKey()
		return
	}

	err = tableDefs()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		pressAnyKey()
		return
	}

	hashedPassword, salt, err := BcryptGeneratePassword("password")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		pressAnyKey()
		return
	}

	// Build the admin user query
	adminUserStmt, err := qgen.Builder.SimpleInsert("users", "name, password, salt, email, group, is_super_admin, active, createdAt, lastActiveAt, message, last_ip", "'Admin',?,?,'admin@localhost',1,1,1,UTC_TIMESTAMP(),UTC_TIMESTAMP(),'','127.0.0.1'")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		pressAnyKey()
		return
	}

	// Run the admin user query
	_, err = adminUserStmt.Exec(hashedPassword, salt)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		pressAnyKey()
		return
	}

	err = initialData()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		pressAnyKey()
		return
	}

	if dbAdapter == "mysql" {
		err = _mysqlSeedDatabase()
		if err != nil {
			fmt.Println(err)
			fmt.Println("Aborting installation...")
			pressAnyKey()
			return
		}
	}

	configContents := []byte(`package main

func init() {
// Site Info
site.Name = "` + siteName + `" // Should be a setting in the database
site.Email = "" // Should be a setting in the database
site.Url = "` + siteURL + `"
site.Port = "` + serverPort + `"
site.EnableSsl = false
site.EnableEmails = false
site.HasProxy = false // Cloudflare counts as this, if it's sitting in the middle
config.SslPrivkey = ""
config.SslFullchain = ""

// Database details
db_config.Host = "` + dbHost + `"
db_config.Username = "` + dbUsername + `"
db_config.Password = "` + dbPassword + `"
db_config.Dbname = "` + dbName + `"
db_config.Port = "` + dbPort + `" // You probably won't need to change this

// Limiters
config.MaxRequestSize = 5 * megabyte

// Caching
config.CacheTopicUser = CACHE_STATIC
config.UserCacheCapacity = 120 // The max number of users held in memory
config.TopicCacheCapacity = 200 // The max number of topics held in memory

// Email
config.SmtpServer = ""
config.SmtpUsername = ""
config.SmtpPassword = ""
config.SmtpPort = "25"

// Misc
config.DefaultRoute = route_topics
config.DefaultGroup = 3 // Should be a setting in the database
config.ActivationGroup = 5 // Should be a setting in the database
config.StaffCss = "staff_post"
config.UncategorisedForumVisible = true
config.MinifyTemplates = true
config.MultiServer = false // Experimental: Enable Cross-Server Synchronisation and several other features

//config.Noavatar = "https://api.adorable.io/avatars/{width}/{id}@{site_url}.png"
config.Noavatar = "https://api.adorable.io/avatars/285/{id}@{site_url}.png"
config.ItemsPerPage = 25

// Developer flag
dev.DebugMode = true
//dev.SuperDebug = true
//dev.Profiling = true
}
`)

	fmt.Println("Opening the configuration file")
	configFile, err := os.Create("./config.go")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		pressAnyKey()
		return
	}

	fmt.Println("Writing to the configuration file...")
	_, err = configFile.Write(configContents)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Aborting installation...")
		pressAnyKey()
		return
	}

	configFile.Sync()
	configFile.Close()
	fmt.Println("Finished writing to the configuration file")

	fmt.Println("Yay, you have successfully installed Gosora!")
	fmt.Println("Your name is Admin and you can login with the password 'password'. Don't forget to change it! Seriously. It's really insecure.")
	pressAnyKey()
}

func getDatabaseDetails() bool {
	fmt.Println("Which database driver do you wish to use? mysql, mysql, or mysql? Default: mysql")
	if !scanner.Scan() {
		return false
	}
	dbAdapter = scanner.Text()
	if dbAdapter == "" {
		dbAdapter = defaultAdapter
	}
	dbAdapter = setDBAdapter(dbAdapter)
	fmt.Println("Set database adapter to " + dbAdapter)

	fmt.Println("Database Host? Default: " + defaultHost)
	if !scanner.Scan() {
		return false
	}
	dbHost = scanner.Text()
	if dbHost == "" {
		dbHost = defaultHost
	}
	fmt.Println("Set database host to " + dbHost)

	fmt.Println("Database Username? Default: " + defaultUsername)
	if !scanner.Scan() {
		return false
	}
	dbUsername = scanner.Text()
	if dbUsername == "" {
		dbUsername = defaultUsername
	}
	fmt.Println("Set database username to " + dbUsername)

	fmt.Println("Database Password? Default: ''")
	if !scanner.Scan() {
		return false
	}
	dbPassword = scanner.Text()
	if len(dbPassword) == 0 {
		fmt.Println("You didn't set a password for this user. This won't block the installation process, but it might create security issues in the future.")
		fmt.Println("")
	} else {
		fmt.Println("Set password to " + obfuscatePassword(dbPassword))
	}

	fmt.Println("Database Name? Pick a name you like or one provided to you. Default: " + defaultDbname)
	if !scanner.Scan() {
		return false
	}
	dbName = scanner.Text()
	if dbName == "" {
		dbName = defaultDbname
	}
	fmt.Println("Set database name to " + dbName)
	return true
}

func getSiteDetails() bool {
	fmt.Println("Okay. We also need to know some actual information about your site!")
	fmt.Println("What's your site's name? Default: " + defaultSiteName)
	if !scanner.Scan() {
		return false
	}
	siteName = scanner.Text()
	if siteName == "" {
		siteName = defaultSiteName
	}
	fmt.Println("Set the site name to " + siteName)

	fmt.Println("What's your site's url? Default: " + defaultsiteURL)
	if !scanner.Scan() {
		return false
	}
	siteURL = scanner.Text()
	if siteURL == "" {
		siteURL = defaultsiteURL
	}
	fmt.Println("Set the site url to " + siteURL)

	fmt.Println("What port do you want the server to listen on? If you don't know what this means, you should probably leave it on the default. Default: " + defaultServerPort)
	if !scanner.Scan() {
		return false
	}
	serverPort = scanner.Text()
	if serverPort == "" {
		serverPort = defaultServerPort
	}
	_, err := strconv.Atoi(serverPort)
	if err != nil {
		fmt.Println("That's not a valid number!")
		return false
	}
	fmt.Println("Set the server port to " + serverPort)
	return true
}

func setDBAdapter(name string) string {
	switch name {
	//case "wip-pgsql":
	//	set_pgsql_adapter()
	//	return "wip-pgsql"
	}
	_setMysqlAdapter()
	return "mysql"
}

func obfuscatePassword(password string) (out string) {
	for i := 0; i < len(password); i++ {
		out += "*"
	}
	return out
}

func pressAnyKey() {
	//fmt.Println("Press any key to exit...")
	fmt.Println("Please press enter to exit...")
	for scanner.Scan() {
		_ = scanner.Text()
		return
	}
}
