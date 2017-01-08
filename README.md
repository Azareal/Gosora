# Gosora

A super fast forum software written in Go.

The initial code-base was forked from one of my side projects, but has now gone far beyond that.

Discord Server: https://discord.gg/eyYvtTf

If you like this software, please give it a star and give us some feedback :)

If you dislike it, please give us some feedback on how to make it better! We're always looking for feedback. We love hearing your opinions. If there's something missing or something doesn't look quite right, don't worry! We plan to change many things before the alpha!


# Features
Basic Forum Functionality. All of the little things you would expect of any forum software. E.g. Moderation, Custom Themes, Avatars, and so on.

Custom Pages. Under development. Mainly the Control Panel portion to come, but you can create them by hand today.

Emojis. Allow your users to express themselves without resorting to serving tons upon tons of image files.

In-memory static file, forum and group caches. We're pondering over extending this solution over to topics, users, etc. to some extent.

A profile system including profile comments and moderation tools for the profile owner.

A template engine which compiles templates down into machine code. Over ten times faster than html/templates. Compatible with templates written for html/templates, you don't need to learn any new templating language.

A plugin system. Under development.

A responsive design. Looks great on mobile phones, tablets, laptops, desktops and more!


# Dependencies

Go 1.7. The programming language this program is written in, and the compiler which it requires. You will need to install this. https://golang.org/doc/install

MySQL Database. You will need to setup a MySQL Database somewhere. A MariaDB Database works equally well, and is much faster than MySQL.
If you're curious about how to install this, you might want to try the WNMP or XAMPP bundles on Windows.
Instructions on how to do so on Linux: https://downloads.mariadb.org/mariadb/repositories/


# Installation Instructions

**Run the following commands:**

go get -u github.com/go-sql-driver/mysql

go get -u golang.org/x/crypto/bcrypt

Tweak the config.go file and put your database details in there. Import data.sql into the same database. Comment out the first line (put /* and */ around it), if you've already made a database, and don't want the script to generate it for you.

Set the password column of your user account in the database to what you want your password to be. The system will encrypt your password when you login for the first time.

You can run these commands again at any time to update these dependencies to their latest versions.

# Run the program

*Linux*

cd to the directory / folder the code is in.

go build

./gosora


*Windows*

Open up cmd.exe

cd to the directory / folder the code is in. E.g. `cd /Users/Blah/Documents/gosora`

go build

./gosora.exe


Alternatively, you could run the run.bat batch file on Windows.

We're also looking into ways to distribute ready made executables for Windows. While this is not a complicated endeavour, the configuration settings currently get built with the rest of the program for speed, and we will likely have to change this.

With the introduction of the new settings system, we will begin moving some of the less critical settings out of the configuration file, and will likely have a config.xml or config.ini in the future to store the critical settings in.

You might have to go build, run the executable, and then go build and then run the executable again to make sure the templates are compiled properly. We'll be resolving this issue while we roll out the new compiled templates system to the rest of the routes.

Several important features for saving memory in the templates system may have to be implemented before the new compiled template system is rolled out to every route. These features are coming fairly soon, but not before the other more high priority items.


# How do I install plugins?

For the default plugins like Markdown and Helloworld, you can find them in the Plugin Manager of your Control Panel. For ones which aren't included by default, you will need to drag them from your /extend/ directory and into the / directory (the root directory of your Gosora installation, where the executable and most of the main Go files are).

You will then need to recompile Gosora in order to link the plugin code with Gosora's code. For plugins not written in Gosora (e.g. JavaScript), you do not need to move them from the /extend/ directory, they will automatically show up in your Control Panel ready to be installed.

Experimental plugins aka the ones in the /experimental/ folder (e.g. plugin_sendmail) are similar but different. You will have to move native plugins (ones written in Go) to the root directory of your installation and will have to move experimental plugins written in other languages into the /extend/ directory.

We're looking for ways to clean-up the plugin system so that all of them (except the experimental ones) are housed in /extend/, however we've encountered some problems with Go's packaging system. We plan to fix this issue in the future.


# Images
![Tempra Simple Theme](https://github.com/Azareal/Gosora/blob/master/images/tempra-simple.png)

![Tempra Conflux Theme](https://github.com/Azareal/Gosora/blob/master/images/tempra-conflux.png)

![Cosmo Conflux Theme](https://github.com/Azareal/Gosora/blob/master/images/cosmo-conflux.png)

![Cosmo Theme](https://github.com/Azareal/Gosora/blob/master/images/cosmo.png)


# TO-DO

Oh my, you caught me right at the start of this project. There's nothing to see here yet, asides from the absolute basics. You might want to look again later!


The various little features which somehow got stuck in the net. Don't worry, I'll get to them!

More moderation features. E.g. Move, Approval Queue (Posts made by users in certain usergroups will need to be approved by a moderator before they're publically visible), etc.

Add a simple anti-spam measure. I have quite a few ideas in mind, but it'll take a while to implement the more advanced ones, so I'd like to put off some of those to a later date and focus on the basics. E.g. CAPTCHAs, hidden fields, etc.

Add an alert system.

Add per-forum permissions to finish up the foundations of the permissions system.

Add a *better* plugin system. E.g. Allow for plugins written in Javascript and ones written in Go. Also, we need to add many, many, many more plugin hooks.

I will need to ponder over implementing an even faster router. We don't need one immediately, although it would be nice if we could get one in the near future. It really depends. Ideally, it would be one which can easily integrate with the current structure without much work, although I'm not beyond making some alterations to faciliate it, assuming that we don't get too tightly bound to that specific router.

Allow themes to define their own templates.

Add a friend system.

Add more administration features.

Add more features for improving user engagement. I have quite a few of these in mind, but I'm mostly occupied with implementing the essentials right now.

Add a widget system.

Add support for multi-factor authentication.

Add support for secondary emails for users.
