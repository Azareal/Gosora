/*
*
* Gosora Installer
* Copyright Azareal 2017 - 2018
*
 */
package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"

	"./install"
)

var scanner *bufio.Scanner

var siteShortName string
var siteName string
var siteURL string
var serverPort string

var defaultAdapter = "mysql"
var defaultHost = "localhost"
var defaultUsername = "root"
var defaultDbname = "gosora"
var defaultSiteShortName = "SN"
var defaultSiteName = "Site Name"
var defaultsiteURL = "localhost"
var defaultServerPort = "80" // 8080's a good one, if you're testing and don't want it to clash with port 80

func main() {
	// Capture panics instead of closing the window at a superhuman speed before the user can read the message on Windows
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
	adap, ok := handleDatabaseDetails()
	if !ok {
		err := scanner.Err()
		if err != nil {
			fmt.Println(err)
		} else {
			err = errors.New("Something went wrong!")
		}
		abortError(err)
		return
	}

	if !getSiteDetails() {
		err := scanner.Err()
		if err != nil {
			fmt.Println(err)
		} else {
			err = errors.New("Something went wrong!")
		}
		abortError(err)
		return
	}

	err := adap.InitDatabase()
	if err != nil {
		abortError(err)
		return
	}

	err = adap.TableDefs()
	if err != nil {
		abortError(err)
		return
	}

	err = adap.CreateAdmin()
	if err != nil {
		abortError(err)
		return
	}

	err = adap.InitialData()
	if err != nil {
		abortError(err)
		return
	}

	configContents := []byte(`package main

import "./common"

func init() {
	// Site Info
	common.Site.ShortName = "` + siteShortName + `" // This should be less than three letters to fit in the navbar
	common.Site.Name = "` + siteName + `"
	common.Site.Email = ""
	common.Site.URL = "` + siteURL + `"
	common.Site.Port = "` + serverPort + `"
	common.Site.EnableSsl = false
	common.Site.EnableEmails = false
	common.Site.HasProxy = false // Cloudflare counts as this, if it's sitting in the middle
	common.Config.SslPrivkey = ""
	common.Config.SslFullchain = ""
	common.Site.Language = "english"

	// Database details
	common.DbConfig.Host = "` + adap.DBHost() + `"
	common.DbConfig.Username = "` + adap.DBUsername() + `"
	common.DbConfig.Password = "` + adap.DBPassword() + `"
	common.DbConfig.Dbname = "` + adap.DBName() + `"
	common.DbConfig.Port = "` + adap.DBPort() + `" // You probably won't need to change this

	// Test Database details
    common.DbConfig.TestHost = ""
    common.DbConfig.TestUsername = ""
    common.DbConfig.TestPassword = ""
    common.DbConfig.TestDbname = "" // The name of the test database, leave blank to disable. DON'T USE YOUR PRODUCTION DATABASE FOR THIS. LEAVE BLANK IF YOU DON'T KNOW WHAT THIS MEANS.
	common.DbConfig.TestPort = ""

	// Limiters
	common.Config.MaxRequestSize = 5 * common.Megabyte

	// Caching
	common.Config.CacheTopicUser = common.CACHE_STATIC
	common.Config.UserCacheCapacity = 120 // The max number of users held in memory
	common.Config.TopicCacheCapacity = 200 // The max number of topics held in memory

	// Email
	common.Config.SMTPServer = ""
	common.Config.SMTPUsername = ""
	common.Config.SMTPPassword = ""
	common.Config.SMTPPort = "25"

	// Misc
	common.Config.DefaultRoute = "routes.TopicList"
	common.Config.DefaultGroup = 3 // Should be a setting in the database
	common.Config.ActivationGroup = 5 // Should be a setting in the database
	common.Config.StaffCSS = "staff_post"
	common.Config.DefaultForum = 2
	common.Config.MinifyTemplates = true
	common.Config.BuildSlugs = true
	common.Config.ServerCount = 1 // Experimental: Enable Cross-Server Synchronisation and several other features

	//common.Config.Noavatar = "https://api.adorable.io/avatars/{width}/{id}@{site_url}.png"
	common.Config.Noavatar = "https://api.adorable.io/avatars/285/{id}@{site_url}.png"
	common.Config.ItemsPerPage = 25

	// Developer flags
	common.Dev.DebugMode = true
	//common.Dev.SuperDebug = true
	//common.Dev.TemplateDebug = true
	//common.Dev.Profiling = true
	//common.Dev.TestDB = true
}
`)

	fmt.Println("Opening the configuration file")
	configFile, err := os.Create("./config.go")
	if err != nil {
		abortError(err)
		return
	}

	fmt.Println("Writing to the configuration file...")
	_, err = configFile.Write(configContents)
	if err != nil {
		abortError(err)
		return
	}

	configFile.Sync()
	configFile.Close()
	fmt.Println("Finished writing to the configuration file")

	fmt.Println("Yay, you have successfully installed Gosora!")
	fmt.Println("Your name is Admin and you can login with the password 'password'. Don't forget to change it! Seriously. It's really insecure.")
	pressAnyKey()
}

func abortError(err error) {
	fmt.Println(err)
	fmt.Println("Aborting installation...")
	pressAnyKey()
}

func handleDatabaseDetails() (adap install.InstallAdapter, ok bool) {
	var dbAdapter string
	var dbHost string
	var dbUsername string
	var dbPassword string
	var dbName string
	// TODO: Let the admin set the database port?
	//var dbPort string

	for {
		fmt.Println("Which database adapter do you wish to use? mysql, mssql, or mysql? Default: mysql")
		if !scanner.Scan() {
			return nil, false
		}
		dbAdapter := scanner.Text()
		if dbAdapter == "" {
			dbAdapter = defaultAdapter
		}
		adap, ok = install.Lookup(dbAdapter)
		if ok {
			break
		}
		fmt.Println("That adapter doesn't exist")
	}
	fmt.Println("Set database adapter to " + dbAdapter)

	fmt.Println("Database Host? Default: " + defaultHost)
	if !scanner.Scan() {
		return nil, false
	}
	dbHost = scanner.Text()
	if dbHost == "" {
		dbHost = defaultHost
	}
	fmt.Println("Set database host to " + dbHost)

	fmt.Println("Database Username? Default: " + defaultUsername)
	if !scanner.Scan() {
		return nil, false
	}
	dbUsername = scanner.Text()
	if dbUsername == "" {
		dbUsername = defaultUsername
	}
	fmt.Println("Set database username to " + dbUsername)

	fmt.Println("Database Password? Default: ''")
	if !scanner.Scan() {
		return nil, false
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
		return nil, false
	}
	dbName = scanner.Text()
	if dbName == "" {
		dbName = defaultDbname
	}
	fmt.Println("Set database name to " + dbName)

	adap.SetConfig(dbHost, dbUsername, dbPassword, dbName, adap.DefaultPort())
	return adap, true
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

	// ? - We could compute this based on the first letter of each word in the site's name, if it's name spans multiple words. I'm not sure how to do this for single word names.
	fmt.Println("Can we have a short abbreviation for your site? Default: " + defaultSiteShortName)
	if !scanner.Scan() {
		return false
	}
	siteShortName = scanner.Text()
	if siteShortName == "" {
		siteShortName = defaultSiteShortName
	}
	fmt.Println("Set the short name to " + siteShortName)

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
