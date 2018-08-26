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

For Ubuntu, you can consult: https://tecadmin.net/install-go-on-ubuntu/
You will also want to run `ln -s /usr/local/go/bin/go` (replace /usr/local with where ever you put Go), so that go becomes visible to other users.

If you followed the instructions above, you can update to the latest version of Go simply by deleting the `/go/` folder and replacing it with a `/go/` folder for the latest version of Go.

Git - You may need this for downloading updates via the updater. You might already have this installed on your server, if the `git` commands don't work, then install this. https://git-scm.com/downloads

MySQL Database - You will need to setup a MySQL Database somewhere. A MariaDB Database works equally well and is much faster than MySQL. You could use something like WNMP / XAMPP which have a little PHP script called PhpMyAdmin for managing MySQL databases or you could install MariaDB directly.

Download the .msi installer from [MariaDB](https://mariadb.com/downloads) and run that. You may want to set it up as a service to avoid running it every-time the computer starts up.

Instructions on how to set MariaDB up on Linux: https://downloads.mariadb.org/mariadb/repositories/

We recommend changing the root password (that is the password for the user 'root'). Remember that password, you will need it for the installation process. Of course, we would advise using a user other than root for maximum security, although that adds additional steps to the process of getting everything setup.

You might also want to run `mysql_secure_installation` to further harden (aka make it more secure) MySQL / MariaDB.

If you're using Ubuntu, you might want to look at: https://www.itzgeek.com/how-tos/linux/ubuntu-how-tos/install-mariadb-on-ubuntu-16-04.html

It's entirely possible that your host already has MySQL installed and ready to go, so you might be able to skip this step, particularly if it's a managed VPS or a shared host. Or they might have a quicker and easier method of setting up MySQL.


# How to download

For Linux, you can skip down to the Installation section as it covers this.

On Windows, you might want to try the [GosoraBootstrapper](https://github.com/Azareal/GosoraBootstrapper), if you can't find the command prompt or otherwise can't follow those instructions. It's just a matter of double-clicking on the bat file there and it'll download the rest of the files for you.

# Installation

Consult [installation](https://github.com/Azareal/Gosora/blob/master/docs/installation.md) for instructions on how to install Gosora.

# Updating

Consult [updating](https://github.com/Azareal/Gosora/blob/master/docs/updating.md) for instructions on how to update Gosora.


# Running the program

*Linux*

If you have setup a service, you can run:

`./pre-run-linux`

`service gosora start`

You can then, check Gosora's current status (to see if it started up properly) with:

`service gosora status`

And you can stop it with:

`service gosora stop`

If you haven't setup a service, you can run `./run-linux`, although you will be responsible for finding a way to run it in the background, so that it doesn't close when the terminal does.

One method might be to use: https://serverfault.com/questions/34750/is-it-possible-to-detach-a-process-from-its-terminal-or-i-should-have-used-s

*Windows*

Run `run.bat`, e.g. double-clicking on it.


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
