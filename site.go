package main

import "net/http"

var site *Site = &Site{Name:"Magical Fairy Land"}
var db_config DB_Config = DB_Config{Host:"localhost"}
var config Config
var dev DevConfig

type Site struct
{
	Name string
	Email string
	Url string
	Port string
	EnableSsl bool
	EnableEmails bool
}

type DB_Config struct
{
	Host string
	Username string
	Password string
	Dbname string
	Port string
}

type Config struct
{
	SslPrivkey string
	SslFullchain string

	MaxRequestSize int
	CacheTopicUser int
	UserCacheCapacity int
	TopicCacheCapacity int

	SmtpServer string
	SmtpUsername string
	SmtpPassword string
	SmtpPort string

	DefaultRoute func(http.ResponseWriter, *http.Request, User)
	DefaultGroup int
	ActivationGroup int
	StaffCss string
	UncategorisedForumVisible bool
	MinifyTemplates bool
	MultiServer bool

	Noavatar string
	ItemsPerPage int
}

type DevConfig struct
{
	DebugMode bool
	SuperDebug bool
	Profiling bool
}
