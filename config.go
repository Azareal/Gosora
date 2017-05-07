package main

// Site Info
var site_name = "Test Install" // Should be a setting in the database
var site_url = "localhost:8080"
var server_port = "8080"
var enable_ssl = false
var ssl_privkey = ""
var ssl_fullchain = ""

// Database details
var dbhost = "localhost"
var dbuser = "root"
var dbpassword = "password"
var dbname = "gosora"
var dbport = "3306" // You probably won't need to change this

// Limiters
var max_request_size = 5 * megabyte

// Caching
var cache_topicuser = CACHE_STATIC
var user_cache_capacity = 100 // The max number of users held in memory
var topic_cache_capacity = 100 // The max number of topics held in memory

// Email
var site_email = "" // Should be a setting in the database
var smtp_server = ""
var smtp_username = ""
var smtp_password = ""
var smtp_port = "25"
var enable_emails = false

// Misc
var default_route = route_topics
var default_group = 3 // Should be a setting in the database
var activation_group = 5 // Should be a setting in the database
var staff_css = " background-color: #ffeaff;"
var uncategorised_forum_visible = true
var minify_templates = true

//var noavatar = "https://api.adorable.io/avatars/{width}/{id}@{site_url}.png"
var noavatar = "https://api.adorable.io/avatars/285/{id}@" + site_url + ".png"
var items_per_page = 25

// Developer flags
var debug = false
var super_debug = false
var profiling = false
