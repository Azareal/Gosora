# Gosora ![Build Status](https://travis-ci.org/Azareal/Gosora.svg?branch=master) [![Azareal's Discord Chat](https://img.shields.io/badge/style-Invite-7289DA.svg?style=flat&label=Discord)](https://discord.gg/eyYvtTf)

A super fast forum software written in Go. You can talk to us on our Discord chat!

The initial code-base was forked from one of my side projects, but has now gone far beyond that. We've moved along in a development and the software should be somewhat stable for general use.

Features may break from time to time, however I will generally try to warn of the biggest offenders in advance, so that you can tread with caution around certain commits, the upcoming v0.1 will undergo even more rigorous testing.

File an issue or open a topic on the forum, if there's something you want and you very well might find it landing in the software fairly quickly.

For plugin and theme developers, things are a little dicier, as the internal APIs and ways of writing themes are in constant flux, however some stability in that area should be coming fairly soon.

If you like this software, please give it a star and give us some feedback :)

If you dislike it, please give us some feedback on how to make it better! We're always looking for feedback. We love hearing your opinions. If there's something missing or something doesn't look quite right, don't worry! We plan to add many, many things in the run up to v0.1!


# Features
Standard Forum Functionality. All of the little things you would expect of any forum software. E.g. Common Moderation features, modlogs, theme system, avatars, bbcode parser, markdown parser, report system, per-forum permissions, group permissions and so on.

Custom Pages. There are some rough edges

Emojis. Allow your users to express themselves without resorting to serving tons upon tons of image files.

In-memory static file, forum and group caches. We have a slightly more dynamic cache for users and topics.

A profile system, including profile comments and moderation tools for the profile owner.

A template engine which compiles templates down to machine code. Over forty times faster than the standard template library `html/templates`, although it does remove some of the hand holding to achieve this. Compatible with templates written for `html/templates`, so you don't need to learn any new templating language.

A plugin system. We have a number of APIs and hooks for plugins, however they're currently subject to change and don't cover as much of the software as we'd like yet.

A responsive design. Looks great on mobile phones, tablets, laptops, desktops and more!

Other modern features like alerts, likes, advanced dashboard with live stats (CPU, RAM, online user count, and so on), etc.


# Requirements

Go 1.10 or newer - You will need to install this. Pick the .msi, if you want everything sorted out for you rather than having to go around updating the environment settings. https://golang.org/doc/install

Git - You may need this for downloading updates via the updater. You might already have this installed on your server. More to come on this here. https://git-scm.com/downloads

MySQL Database - You will need to setup a MySQL Database somewhere. A MariaDB Database works equally well and is much faster than MySQL. You could use something like WNMP / XAMPP which have a little PHP script called PhpMyAdmin for managing MySQL databases or you could install MariaDB directly.

Download the .msi installer from [MariaDB](https://mariadb.com/downloads) and run that. You may want to set it up as a service to avoid running it every-time the computer starts up.

Instructions on how to set MariaDB up on Linux: https://downloads.mariadb.org/mariadb/repositories/

We recommend changing the root password (that is the password for the user 'root'). Remember that password, you will need it for the installation process. Of course, we would advise using a user other than root for maximum security, although that adds additional steps to the process of getting everything setup.

You might also want to run `mysql_secure_installation` to further harden (aka make it more secure) MySQL / MariaDB.

It's entirely possible that your host already has MySQL installed and ready to go, so you might be able to skip this step, particularly if it's a managed VPS or a shared host (contrary to popular belief, it is possible, although the ecosystem in this regard is extremely immature). Or they might have a quicker and easier method of setting up MySQL.


# How to download

For Linux, you can skip down to the How to install section.

On Windows, you might want to try the [GosoraBootstrapper](https://github.com/Azareal/GosoraBootstrapper), if you can't find the command prompt or otherwise can't follow those instructions. It's just a matter of double-clicking on the bat file there and it'll download the rest of the files for you.


# How to install

*Linux*

First, you will need to jump to the place where you want to put the code, we will use `/home/gosora` here, but if you want to use something else, then you'll have to modify the service file with your own path (but *never* in a folder where the files are automatically served by a webserver).

If you place it in `/www/`, `/public_html/` or any similar folder, then there's a chance that your server might be compromised.

You can navigate to it by typing the following six commands into the console and hitting enter:

cd /home/

git clone https://github.com/Azareal/Gosora

mv Gosora gosora

cd gosora

chmod 755 ./install-linux

./install-linux

Follow the instructions shown on the screen.

You will also want to setup a service to manage Gosora more easily, although this will require administrator priviledges on the machine:

chmod 755 ./pre-run-linux

mv ./gosora_example.service /lib/systemd/system/gosora.service

systemctl daemon-reload

*Windows*

Run install.bat, e.g. double-click on it. You will also have to start-up MySQL, which if you're using Wnmp or friends is just a matter of opening that program and starting the MySQL process via it.

Follow the instructions shown on the screen.


# Running the program

*Linux*

If you have setup a service, you can run:

./pre-run-linux

service gosora start

You can then, check Gosora's current status (to see if it started up properly) with:

service gosora status

And you can stop it with:

service gosora stop

If you haven't setup a service, you can run `./run-linux`, although you will be responsible for finding a way to run it in the background, so that it doesn't close when the terminal does.

*Windows*

Run run.bat, e.g. double-clicking on it.

*Updating Dependencies*

You can update the dependencies which Gosora relies on by running update-deps.bat on Windows or ./update-deps-linux on Linux. These dependencies do not include Go or MySQL, those have to be updated separately.

We're also looking into ways to distribute ready made executables for Windows. While this is not a complicated endeavour, the configuration settings currently get built with the rest of the program for speed, and we will likely have to change this.

With the introduction of the new settings system, we will begin moving some of the less critical settings out of the configuration file, and will likely have a config.xml or config.ini in the future to store the critical settings in.

You'll need to restart the server every-time you change a template, e.g. with `run.bat` or killing the process and running `./run-linux` or via `./pre-run-linux` followed by `service gosora restart`

Several important features for saving memory in the templates system may have to be implemented before the new compiled template system is rolled out to every route. These features are coming fairly soon, but not before the higher priority items.


# Advanced Installation

This section explains how to set things up without running the batch or shell files. For Windows, you will likely have to open up cmd.exe (the app called Command Prompt in Win10) to run these commands inside or something similar, while with Linux you would likely use the Terminal or console.

Linux is similar, however you might need to use cd and mv a bit more like in the shell files due to the differences in go build across platforms. Additionally, Linux doesn't require `StackExchange/wmi` or `/x/sys/windows`

You also need to substitute the `gosora.exe` bits for `./Gosora` on Linux. For more info, you might want to take a gander inside the `./run-linux` and `./install-linux` shell files to see how they're implemented.

If you want to skip typing all the `go get`s, you can run `./update-deps.bat` (Windows) or `./update-deps-linux` to do that for you.

```bash
git clone https://github.com/Azareal/Gosora

go get -u github.com/go-sql-driver/mysql

go get -u golang.org/x/crypto/bcrypt

go get -u golang.org/x/crypto/argon2

go get -u github.com/StackExchange/wmi

go get -u github.com/Azareal/gopsutil

go get -u github.com/gorilla/websocket

go get -u gopkg.in/sourcemap.v1

go get -u github.com/robertkrimen/otto

go get -u github.com/esimov/caire

go get -u github.com/lib/pq

go get -u github.com/denisenkom/go-mssqldb

go get -u github.com/fsnotify/fsnotify

go get -u github.com/pkg/errors

rm -f template_*.go

rm -f gen_*.go

rm -f tmpl_client/template_*.go

rm -f ./Gosora

go generate

go build ./router_gen

router_gen.exe

go build ./query_gen

query_gen.exe

go build -o gosora.exe

go build ./install

install.exe

gosora.exe -build-templates

gosora.exe
```

I'm looking into minimising the number of go gets for the advanced build and to maybe remove the platform and database engine specific dependencies if possible for those who don't need them.


# Updating the software

The update system is currently under development, however if you have Git installed, then you can run `dev-update.bat` or `dev-update-linux` to update your instance to the latest commit and to update the associated database schema, etc.

In addition to this, you can update the dependencies without updating Gosora by running `update-deps.bat` or `./update-deps-linux` (.bat is for Windows, the other for Linux as the names would suggest).

If you want to manually patch Gosora rather than relying on the above scripts to do it, you'll first have to create a copy of `./schema/schema.json` named `./schema/lastSchema.json`, and then, you'll overwrite the files with the new ones.

After that, you'll need to run `go build ./patcher` on Windows or the following code block on Linux:
```
cd ./patcher
go build -o Patcher
mv ./Patcher ..
```

Once you've done that, you just need to run `patcher.exe` (Windows) or `./Patcher` to apply the latest patches to the database, etc.


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

![Cosora Prototype WIP](https://github.com/Azareal/Gosora/blob/master/images/cosora-wip.png)

More images in the /images/ folder. Beware though, some of them are *really* outdated. Also, keep in mind that a new theme is in the works.

# Dependencies 

These are the libraries and pieces of software which Gosora relies on to function, an "ingredients" list so to speak.

A few of these like Rez aren't currently in use, but are things we think we'll need in the very near future and want to have those things ready, so that we can quickly slot them in.

* Go 1.10+

* MariaDB (or any other MySQL compatible database engine). We'll allow other database engines in the future.

* github.com/go-sql-driver/mysql For interfacing with MariaDB.

* golang.org/x/crypto/bcrypt For hashing passwords.
* golang.org/x/crypto/argon2 For hashing passwords.

* github.com/Azareal/gopsutil For pulling information on CPU and memory usage. I've temporarily forked this, as we were having stability issues with the latest build.

  * github.com/StackExchange/wmi Dependency for gopsutil on Windows.

  * golang.org/x/sys/windows Also a dependency for gopsutil on Windows. This isn't needed at the moment, as I've rolled things back to an older more stable build.

* github.com/gorilla/websocket Needed for Gosora's Optional WebSockets Module.

* github.com/robertkrimen/otto Needed for the upcoming JS plugin type.

  * gopkg.in/sourcemap.v1 Dependency for Otto.

* github.com/lib/pq For interfacing with PostgreSQL. You will be able to pick this instead of MariaDB soon.

* ithub.com/denisenkom/go-mssqldb For interfacing with MSSQL. You will be able to pick this instead of MSSQL soon.

* github.com/bamiaux/rez An image resizer (e.g. for spitting out thumbnails)

  * github.com/esimov/caire The other image resizer, slower but may be useful for covering cases Rez does not. A third faster one we might point to at some point is probably Discord's Lilliput, however it requires a C Compiler and we don't want to add that as a dependency at this time.

* github.com/fsnotify/fsnotify A library for watching events on the file system.

* github.com/pkg/errors Some helpers to make it easier for us to track down bugs.

* More items to come here, our dependencies are going through a lot of changes, and I'll be documenting those soon ;)

# Bundled Plugins

There are several plugins which are bundled with the software by default. These cover various common tasks which aren't common enough to clutter the core with or which have competing implementation methods (E.g. plugin_markdown vs plugin_bbcode for post mark-up).

* Hey There / Skeleton / Hey There (JS Version) - Example plugins for helping you learn how to develop plugins.

* BBCode - A plugin in early development for converting BBCode Tags into HTML.

* Markdown - An extremely simple plugin for converting Markdown into HTML.

* Social Groups - An extremely unstable WIP plugin which lets users create their own little discussion areas which they can administrate / moderate on their own.

# Developers

There are a few things you'll need to know before running the more developer oriented features like the tests or the benchmarks.

The benchmarks are currently being rewritten as they're currently extremely serial which can lead to severe slow-downs when run on a home computer due to the benchmarks being run on the one core everything else is being run on (Browser, OS, etc.) and the tests not taking parallelism into account.
