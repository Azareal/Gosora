package main

import "./common"

func init() {
	// Site Info
	common.Site.ShortName = "Ts" // This should be less than three letters to fit in the navbar
	common.Site.Name = "Test Site"
	common.Site.Email = ""
	common.Site.URL = "localhost"
	common.Site.Port = "8080" // 8080
	common.Site.EnableSsl = false
	common.Site.EnableEmails = false
	common.Site.HasProxy = false // Cloudflare counts as this, if it's sitting in the middle
	common.Config.SslPrivkey = ""
	common.Config.SslFullchain = ""
	common.Site.Language = "english"

	// Database details
	common.DbConfig.Host = "localhost"
	common.DbConfig.Username = "root"
	common.DbConfig.Password = "password"
	common.DbConfig.Dbname = "gosora"
	common.DbConfig.Port = "3306" // You probably won't need to change this

	// MySQL Test Database details
	common.DbConfig.TestHost = "localhost"
	common.DbConfig.TestUsername = "root"
	common.DbConfig.TestPassword = ""
	common.DbConfig.TestDbname = "gosora_test" // The name of the test database, leave blank to disable. DON'T USE YOUR PRODUCTION DATABASE FOR THIS. LEAVE BLANK IF YOU DON'T KNOW WHAT THIS MEANS.
	common.DbConfig.TestPort = "3306"

	// Limiters
	common.Config.MaxRequestSize = 5 * common.Megabyte

	// Caching
	common.Config.CacheTopicUser = common.CACHE_STATIC
	common.Config.UserCacheCapacity = 120  // The max number of users held in memory
	common.Config.TopicCacheCapacity = 200 // The max number of topics held in memory

	// Email
	common.Config.SMTPServer = ""
	common.Config.SMTPUsername = ""
	common.Config.SMTPPassword = ""
	common.Config.SMTPPort = "25"

	// Misc
	common.Config.DefaultRoute = routeTopics
	common.Config.DefaultGroup = 3    // Should be a setting in the database
	common.Config.ActivationGroup = 5 // Should be a setting in the database
	common.Config.StaffCSS = "staff_post"
	common.Config.DefaultForum = 2
	common.Config.MinifyTemplates = true
	common.Config.MultiServer = false // Experimental: Enable Cross-Server Synchronisation and several other features

	//common.Config.Noavatar = "https://api.adorable.io/avatars/{width}/{id}@{site_url}.png"
	common.Config.Noavatar = "https://api.adorable.io/avatars/285/{id}@{site_url}.png"
	common.Config.ItemsPerPage = 25

	// Developer flags
	//common.Dev.DebugMode = true
	//common.Dev.SuperDebug = true
	//common.Dev.TemplateDebug = true
	//common.Dev.Profiling = true
	//common.Dev.TestDB = true
}
