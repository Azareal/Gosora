package main

// Database details
var dbhost = "127.0.0.1"
var dbuser = "root"
var dbpassword = "password"
var dbname = "gosora"
var dbport = "3306" // You probably won't need to change this

// Limiters
var max_request_size = 5 * megabyte

// Misc
var default_route = route_topics
var default_group = 3
var staff_css = " background-color: #ffeaff;"
var uncategorised_forum_visible = true
var enable_emails = false
var site_email = ""
var smtp_server = ""
var siteurl = "localhost:8080"
var noavatar = "https://api.adorable.io/avatars/285/{id}@" + siteurl + ".png"
var items_per_page = 50

// Developer flag
var debug = false