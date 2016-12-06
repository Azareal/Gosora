package main

// Database details
var dbhost = "127.0.0.1"
var dbuser = "root"
var dbpassword = "password"
var dbname = "grosolo"
var dbport = "3306" // You probably won't need to change this

// Limiters
var max_request_size = 5 * megabyte

// Misc
var default_route = route_topics
var staff_css = "background-color: #ffeaff;background-position: left;"
var uncategorised_forum_visible = true