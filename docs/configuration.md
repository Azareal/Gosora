# Configuration

For configuring the system, Gosora has a file called `config/config.json` which you can tweak to change various behaviours, it also has a few settings in the Setting Manager in the Control Panel.

The configuration file has five categories you may be familiar with from poring through it's contents. Site, Config, Database, Dev and Plugin.

Site is for critical settings.

Config is for lower priority yet still important settings.

Database contains the credentials for the database (you will be able to pass these via parameters to the binary in a future version).

Dev is for a few flags which help out with the development of Gosora.

Plugin which you may not have run into is a category in which plugins can define their own custom configuration settings.

An example of what the file might look like: https://github.com/Azareal/Gosora/blob/master/config/config_example.json

# Site

ShortName - A two or three letter abbreviation of your site's name. Intended for compact spaces where the full name is too long to squeeze in.

Name - The name of your site, as appears in the title, some theme headers and search engine search results.

Email - The email address you want to show up in the From: field when Gosora sends emails. May be left blank, if emails are disabled.

URL - The URL for your site. Please leave out the `http://` or `https://` and the `/` at the end.

Port - The port you want Gosora to listen on. This will usually be 443 for HTTPS and 80 for HTTP. Gosora will try to bind to both, if you're on HTTPS to redirect users from the HTTP site to the HTTPS one.

EnableSsl - Determines whether HTTPS is enabled.

EnableEmails - Determines whether the SMTP mail subsystem is enabled. The experimental plugin sendmail also allows you to send emails without SMTP in a similar style to some languages like PHP, although it only works on Linux and has some issues.

HasProxy - Brittle, but lets you set whether you're sitting behind a proxy like Cloudflare. Unknown effects with reverse-proxies like Nginx.

Language - The language you want to use. Defaults to english. Please consult [internationalisation](https://github.com/Azareal/Gosora/blob/master/docs/internationalisation.md) for details.

# Config

SslPrivkey - The path to the SSL private key. Example: ./cert/test.key

SslFullchain - The path to the fullchain SSL certificate. Example: ./cert/test.cert

SMTPServer - The domain for the SMTP server. May be left blank if EnableEmails is false.

SMTPUsername - The username for the SMTP server.

SMTPPassword - The password for the SMTP server.

SMTPPort - The port for the SMTP server, usually 25 or 465 for full TLS.

SMTPEnableTLS - Enable TLS to fully encrypt the connection between Gosora and the SMTP server.

Search - The type of search system to use. Options: disabled, sql (default)

MaxRequestSizeStr - The maximum size that a request made to Gosora can be. This includes uploads. Example: 5MB

UserCache - The type of user cache you want to use. You can leave this blank to disable this feature or use `static` for a small in-memory cache.

TopicCache - The type of topic cache you want to use. You can leave this blank to disable this feature or use `static` for a small in-memory cache.

ReplyCache - The type of reply cache you want to use. You can leave this blank to disable this feature or use `static` for a small in-memory cache.

UserCacheCapacity - The maximum number of users you want in the in-memory user cache, if enabled in the UserCache setting.

TopicCacheCapacity - The maximum number of topics you want in the in-memory topic cache, if enabled in the TopicCache setting.

ReplyCacheCapacity - The maximum number of replies you want in the in-memory reply cache, if enabled in the ReplyCache setting.

DefaultPath - The route you want the homepage `/` to default to. Examples: `/topics/` or `/forums/`

DefaultGroup - The group you want users to be moved to once they're activated. Example: 3

ActivationGroup - The group you want users to be placed in while they're awaiting activation. Example: 5

StaffCSS - Classes you want applied to the postbits for staff posts. This setting is deprecated and will likely be replaced with a more generic mechanism in the near future.

DefaultForum - The default forum for the drop-down in the quick topic creator. Please note that FID 1 is reserved for the default reports forum. Example: 2

MinifyTemplates - Whether you want the HTML pages to be minified prior to being send to the client.

BuildSlugs - Whether you want the title appear in the URL. For instance: `/topic/traffic-in-paris.5` versus `/topic/5`

ServerCount - The number of instances you're running. This setting is currently experimental.

PostIPCutoff - The number of days which need to pass before the IP data for a post is automatically deleted. 0 defaults to whatever the current default is, currently 180 and -1 disables this feature. Default: 0

LogPruneCutoff - The number of days which need to pass before the login and registration logs are pruned. 0 defaults to whatever the current default is, currently 365 and -1 disables this feature. Default: 0

DisableLiveTopicList - This switch allows you to disable the live topic list. Default: false

DisableJSAntispam - This switch lets you disable the JS anti-spam feature. It may be useful if you primarily get users who for one reason or another have decided to disable JavaScript. Default: false

LooseHost - Disable host header checks in the router. This may be useful when using a reverse-proxy like Nginx / Apache to stop it white-screening. Default: false

LoosePort - Disable port match checks in the router. This may be useful when using a revere-proxy like Nginx / Apache to stop it white-screening. Default: false

DisableServerPush - This switch lets you disable the HTTP/2 server push feature. Default: false

EnableCDNPush - This switch lets you enable the HTTP/2 CDN Server Push feature. This operates by sending a Link header on every request and may also work with reverse-proxies like Nginx for doing HTTP/2 server pushes.

DisableNoavatarRange - This switch lets you disable the noavatar algorithm which maps IDs to a set ranging from 0 to 50 for better cacheability. Default: false

DisableDefaultNoavatar - This switch lets you disable the default noavatar algorithm which may intercept noavatars for increased efficiency. Default: false

RefNoTrack - This switch disables tracking the referrers of users who click from another site to your site and the referrers of any requests to resources from other sites as-well.

RefNoRef - This switch makes it so that if a user clicks on a link, then the incoming site won't know which site they're coming from.

NoAvatar - The default avatar to use for users when they don't have their own. The default for this may change in the near future to better utilise HTTP/2. Example: https://api.adorable.io/avatars/{width}/{id}.png

ItemsPerPage - The number of posts, topics, etc. you want on each page.

MaxTopicTitleLength - The maximum length that a topic can be. Please note that this measures the number of bytes and may differ from language to language with it being equal to a letter in English and being two bytes in others.

MaxUsernameLength - The maximum length that a user's name can be. Please note that this measures the number of bytes and may differ from language to language with it being equal to a letter in English and being two bytes in others.

ReadTimeout - The number of seconds that we are allowed to take to fully read a request. Defaults to 8.

WriteTimeout - The number of seconds that a route is allowed to run for before the request is automatically terminated. Defaults to 10.

IdleTimeout - The number of seconds that a Keep-Alive connection will be kept open before being closed. You might to tweak this, if you use Cloudflare or similar. Defaults to 120.

Related: https://support.cloudflare.com/hc/en-us/articles/212794707-General-Best-Practices-for-Load-Balancing-at-your-origin-with-Cloudflare


# Database

Adapter - The name of the database adapter. `mysql` and `mssql` are options, although mssql may not work properly in the latest version of Gosora. PgSQL support is in the works.

Host - The host of the database you wish to connect to. Example: localhost

Username - The username for the database you wish to connect to. Example: root

Password - The password for the database you wish to connect to. Example: password

Dbname - The name of the database you want to use. Example: gosora

Port - The port the database is listening on. Usually 3306 for MySQL.

TestAdapter - A test version of Adapter. Only used for testing purposes.

TestHost - A test version of Host. Only used for testing purposes.

TestUsername - A test version of Username. Only used for testing purposes.

TestPassword - A test version of Password. Only used for testing purposes.

TestDbname - A test version of Dbname. Only used for testing purposes.

TestPort - A test version of Port. Only used for testing purposes.

# Dev

DebugMode - Outputs a basic level of information about what Gosora is doing to help ease debugging.

SuperDebug - Outputs a detailed level of information about what Gosora is doing to help ease debugging. Warning: This may cause severe slowdowns in Gosora.

TemplateDebug - Output a detailed level of information about what Gosora is doing during the template transpilation step. Warning: Large amounts of information will be dumped into the logs.

NoFsnotify - Whether you want to disable the file watcher which automatically reloads assets whenever they change.