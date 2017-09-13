package main

func init() {
	// Site Info
	site.Name = "TS"
	site.Email = ""
	site.URL = "localhost"
	site.Port = "8080" // 8080
	site.EnableSsl = false
	site.EnableEmails = false
	site.HasProxy = false // Cloudflare counts as this, if it's sitting in the middle
	config.SslPrivkey = ""
	config.SslFullchain = ""
	site.Language = "english"

	// Database details
	dbConfig.Host = "localhost"
	dbConfig.Username = "root"
	dbConfig.Password = "password"
	dbConfig.Dbname = "gosora"
	dbConfig.Port = "3306" // You probably won't need to change this

	// Limiters
	config.MaxRequestSize = 5 * megabyte

	// Caching
	config.CacheTopicUser = CACHE_STATIC
	config.UserCacheCapacity = 120  // The max number of users held in memory
	config.TopicCacheCapacity = 200 // The max number of topics held in memory

	// Email
	config.SmtpServer = ""
	config.SmtpUsername = ""
	config.SmtpPassword = ""
	config.SmtpPort = "25"

	// Misc
	config.DefaultRoute = routeTopics
	config.DefaultGroup = 3    // Should be a setting in the database
	config.ActivationGroup = 5 // Should be a setting in the database
	config.StaffCss = "staff_post"
	config.UncategorisedForumVisible = true
	config.MinifyTemplates = false
	config.MultiServer = false // Experimental: Enable Cross-Server Synchronisation and several other features

	//config.Noavatar = "https://api.adorable.io/avatars/{width}/{id}@{site_url}.png"
	config.Noavatar = "https://api.adorable.io/avatars/285/{id}@{site_url}.png"
	config.ItemsPerPage = 25

	// Developer flag
	dev.DebugMode = true
	//dev.SuperDebug = true
	//dev.Profiling = true
}
