package main

// Database details
var dbhost = "localhost"
var dbuser = "root"
var dbpassword = "password"
var dbname = "gosora"
var dbport = "3306" // You probably won't need to change this

// Limiters
var max_request_size = 5 * megabyte

// Misc
var default_route = route_topics
var default_group = 3 // Should be a setting
var activation_group = 5 // Should be a setting
var staff_css = " background-color: #ffeaff;"
var uncategorised_forum_visible = true
var enable_emails = false
var site_name = "Test Install" // Should be a setting
var site_email = "" // Should be a setting
var smtp_server = ""
//var noavatar = "https://api.adorable.io/avatars/{width}/{id}@{site_url}.png"
var noavatar = "https://api.adorable.io/avatars/285/{id}@" + site_url + ".png"
var items_per_page = 25

var site_url = "localhost:8080"
var server_port = "8080"
var enable_ssl = false
var ssl_privkey = ""
var ssl_fullchain = ""

// Developer flag
var debug = false
var profiling = false
