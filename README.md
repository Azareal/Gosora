# Gosora [![Azareal's Discord Chat](https://img.shields.io/badge/style-Invite-7289DA.svg?style=flat&label=Discord)](https://discord.gg/eyYvtTf)

A super fast forum software written in Go. You can talk to us on our Discord chat!

The initial code-base was forked from one of my side projects, but has now gone far beyond that. We're still fairly early in development, so the code-base might change at an incredible rate. We plan to stop making as many breaking changes once we release the first alpha.

If you like this software, please give it a star and give us some feedback :)

If you dislike it, please give us some feedback on how to make it better! We're always looking for feedback. We love hearing your opinions. If there's something missing or something doesn't look quite right, don't worry! We plan to change many things before the alpha!


# Features
Basic Forum Functionality. All of the little things you would expect of any forum software. E.g. Common Moderation features, modlogs, theme system, avatars, bbcode parser, markdown parser, report system, per-forum permissions, group permissions and so on.

Custom Pages. Under development. The Control Panel portion is incomplete, but you can create them by hand today. They're basically html/templates templates in the /pages/ folder.

Emojis. Allow your users to express themselves without resorting to serving tons upon tons of image files.

In-memory static file, forum and group caches. We have a slightly more dynamic cache for users and topics.

A profile system, including profile comments and moderation tools for the profile owner.

A template engine which compiles templates down to machine code. Over thirty times faster than html/templates, although it does remove some of the hand holding to achieve this. Compatible with templates written for html/templates, you don't need to learn any new templating language.

A plugin system. More on this to come.

A responsive design. Looks great on mobile phones, tablets, laptops, desktops and more!

Other modern features like alerts, likes, advanced dashboard with live stats (CPU, RAM, online user count, and so on), etc.


# Dependencies

Go 1.9 - You will need to install this. Pick the .msi, if you want everything sorted out for you rather than having to go around updating the environment settings. https://golang.org/doc/install

MySQL Database - You will need to setup a MySQL Database somewhere. A MariaDB Database works equally well and is much faster than MySQL. You could use something like WNMP / XAMPP which have a little PHP script called PhpMyAdmin for managing MySQL databases or you could install MariaDB directly.

Download the .msi installer from [MariaDB](https://mariadb.com/downloads) and run that. You may want to set it up as a service to avoid running it every-time the computer starts up.

Instructions on how to set MariaDB up on Linux: https://downloads.mariadb.org/mariadb/repositories/

We recommend changing the root password (that is the password for the user 'root'). Remember that password, you will need it for the installation process. Of course, we would advise using a user other than root for maximum security, although that adds additional steps to the process of getting everything setup.

It's entirely possible that your host might already have MySQL, so you might be able to skip this step, particularly if it's a managed VPS or a shared host (contrary to popular belief, it is possible, although the ecosystem in this regard is extremely immature). Or they might have a quicker and easier method of setting up MySQL.


# Installation Instructions

*Linux*

cd to the directory / folder the code is in. In other words, type cd followed by the location of the code and it should jump there.

Run ./install-gosora-linux

Follow the instructions shown on the screen.

*Windows*

Run install.bat

Follow the instructions shown on the screen.


# Run the program

*Linux*

In the same directory you installed it, you simply have to type: ./run-gosora-linux

*Windows*

Run run.bat

*Updating Dependencies*

You can update the dependencies which Gosora relies on by running update-deps.bat on Windows or ./update-deps-linux on Linux. These dependencies do not include Go or MySQL.

We're also looking into ways to distribute ready made executables for Windows. While this is not a complicated endeavour, the configuration settings currently get built with the rest of the program for speed, and we will likely have to change this.

With the introduction of the new settings system, we will begin moving some of the less critical settings out of the configuration file, and will likely have a config.xml or config.ini in the future to store the critical settings in.

You might have to run run.bat or ./run-gosora-linux twice after changing a template to make sure the templates are compiled properly. We'll be resolving this issue while rolling out the new compiled templates system to the rest of the routes.

Several important features for saving memory in the templates system may have to be implemented before the new compiled template system is rolled out to every route. These features are coming fairly soon, but not before the higher priority items.


# Advanced Installation

An example of running the commands directly on Windows.

Linux is similar, however you might need to use cd and mv a bit more like in the shell files due to the differences in go build across platforms. Additionally, Linux doesn't require `StackExchange/wmi` or ``/x/sys/windows`

```bash
git clone https://github.com/Azareal/Gosora

go get -u github.com/go-sql-driver/mysql

go get -u golang.org/x/crypto/bcrypt

go get -u github.com/StackExchange/wmi

go get -u github.com/Azareal/gopsutil

go get -u github.com/gorilla/websocket

go get -u gopkg.in/sourcemap.v1

go get -u github.com/robertkrimen/otto

go get -u github.com/lib/pq

go get -u github.com/denisenkom/go-mssqldb

go get -u github.com/fsnotify/fsnotify


go generate

go build ./router_gen

router_gen.exe

go build ./query_gen

query_gen.exe

go build -o gosora.exe

go build ./install

install.exe

gosora.exe
```

I'm looking into minimising the number of go gets for the advanced build and to maybe remove the platform and database engine specific dependencies if possible for those who don't need them.


# How do I install plugins?

For the default plugins like Markdown and Helloworld, you can find them in the Plugin Manager of your Control Panel. For ones which aren't included by default, you will need to drag them from your /extend/ directory and into the / directory (the root directory of your Gosora installation, where the executable and most of the main Go files are).

You will then need to recompile Gosora in order to link the plugin code with Gosora's code. For plugins not written in Gosora (e.g. JavaScript), you do not need to move them from the /extend/ directory, they will automatically show up in your Control Panel ready to be installed.

Experimental plugins aka the ones in the /experimental/ folder (e.g. plugin_sendmail) are similar but different. You will have to move native plugins (ones written in Go) to the root directory of your installation and will have to move experimental plugins written in other languages into the /extend/ directory.

We're looking for ways to clean-up the plugin system so that all of them (except the experimental ones) are housed in /extend/, however we've encountered some problems with Go's packaging system. We plan to fix this issue in the future.


# Images
![Shadow Theme](https://github.com/Azareal/Gosora/blob/master/images/shadow.png)

![Shadow Quick Topic](https://github.com/Azareal/Gosora/blob/master/images/quick-topics.png)

![Tempra Simple Theme](https://github.com/Azareal/Gosora/blob/master/images/tempra-simple.png)

![Tempra Simple Topic List](https://github.com/Azareal/Gosora/blob/master/images/topic-list.png)

![Tempra Simple Mobile](https://github.com/Azareal/Gosora/blob/master/images/tempra-simple-mobile-375px.png)

![Tempra Cursive Theme](https://github.com/Azareal/Gosora/blob/master/images/tempra-cursive.png)

![Tempra Conflux Theme](https://github.com/Azareal/Gosora/blob/master/images/tempra-conflux.png)

![Tempra Conflux Mobile](https://github.com/Azareal/Gosora/blob/master/images/tempra-conflux-mobile-320px.png)

![Cosora Prototype WIP](https://github.com/Azareal/Gosora/blob/master/images/cosora-wip.png)

More images in the /images/ folder. Beware though, some of them are *really* outdated.

# Dependencies (a few of these like Riot aren't currently in use, but we anticipate that we'll need some sort of search engine library in the very immediate future)

* Go 1.9

* MariaDB (or any other MySQL compatible database engine). We'll allow other database engines in the future.

* github.com/go-sql-driver/mysql For interfacing with MariaDB.

* golang.org/x/crypto/bcrypt For hashing passwords.

* github.com/Azareal/gopsutil For pulling information on CPU and memory usage. I've temporarily forked this, as we were having stability issues with the latest build.

  * github.com/StackExchange/wmi Dependency for gopsutil on Windows.

  * golang.org/x/sys/windows Also a dependency for gopsutil on Windows. This isn't needed at the moment, as I've rolled things back to an older more stable build.

* github.com/gorilla/websocket Needed for Gosora's Optional WebSockets Module.

* github.com/robertkrimen/otto Needed for the upcoming JS plugin type.

  * gopkg.in/sourcemap.v1 Dependency for Otto.

* github.com/lib/pq For interfacing with PostgreSQL. You will be able to pick this instead of MariaDB soon.

* ithub.com/denisenkom/go-mssqldb For interfacing with MSSQL. You will be able to pick this instead of MSSQL soon.

* github.com/go-ego/riot A search engine library.

* github.com/bamiaux/rez An image resizer (e.g. for spitting out thumbnails)

* github.com/fsnotify/fsnotify A library for watching events on the file system.

# Bundled Plugins

There are several plugins which are bundled with the software by default. These cover various common tasks which aren't common enough to clutter the core with or which have competing implementation methods (E.g. plugin_markdown vs plugin_bbcode for post mark-up).

* Hey There / Skeleton / Hey There (JS Version) - Example plugins for helping you learn how to develop plugins.

* BBCode - A plugin in early development for converting BBCode Tags into HTML.

* Markdown - An extremely simple plugin for converting Markdown into HTML.

* Social Groups - A WIP plugin which lets users create their own little discussion areas which they can administrate / moderate on their own.

# Developers

There are a few things you'll need to know before running the more developer oriented features like the tests or the benchmarks.

The benchmarks are currently being rewritten as they're currently extremely serial which can lead to severe slow-downs when run on a home computer due to the benchmarks being run on the one core everything else is being run on (Browser, OS, etc.) and the tests not taking parallelism into account.
