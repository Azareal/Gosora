# Gosora

A super fast forum software written in Go.

The initial code-base was forked from one of my side projects, but has now gone far beyond that.

Discord Server: https://discord.gg/eyYvtTf


# Features
Basic Forum Functionality

Custom Pages

Emojis

In-memory static file, forum and group caches.

A profile system including profile comments and moderation tools for the profile owner.

A template engine which compiles templates down into machine code. Over ten times faster than html/templates.

A plugin system.


# Dependencies

Go 1.7. The programming language this program is written in, and the compiler which it requires. You will need to install this. https://golang.org/doc/install

MySQL Database. You will need to setup a MySQL Database somewhere. A MariaDB Database works equally well, and is much faster than MySQL.
If you're curious about how to install this, you might want to try the WNMP or XAMPP bundles on Windows.
Instructions on how to do so on Linux: https://downloads.mariadb.org/mariadb/repositories/


# Installation Instructions

**Run the following commands:**

go get github.com/go-sql-driver/mysql

go install github.com/go-sql-driver/mysql

go get golang.org/x/crypto/bcrypt

go install golang.org/x/crypto/bcrypt

Tweak the config.go file and put your database details in there. Import data.sql into the same database. Comment out the first line (put /* and */ around it), if you've already made a database, and don't want the script to generate it for you.

Set the password column of your user account in the database to what you want your password to be. The system will encrypt your password when you login for the first time.

Add -u after go get to update those libraries, if you've already got them installed.

# Run the program

*Linux*

cd to the directory / folder the code is in.

go build

./gosora


*Windows*

Open up cmd.exe

cd to the directory / folder the code is in. E.g. cd /Users/Blah/Documents/gosora

go build

./gosora.exe


Alternatively, you could run the run.bat batch file on Windows.

We're also looking into ways to distribute ready made executables for Windows. While this is not a complicated endeavour, the configuration settings currently get built with the rest of the program for speed, and we will likely have to change this.

With the introduction of the new settings system, we will begin moving some of the less critical settings out of the configuration file, and will likely have a config.xml or config.ini in the future to store the critical settings in.


# TO-DO

Oh my, you caught me right at the start of this project. There's nothing to see here yet, asides from the absolute basics. You might want to look again later!


More moderation features.

Fix the custom pages.

Add emails as a requirement for registration and add a simple anti-spam measure.

Add an alert system.

Add a complex permissions system.

Add a plugin system.

Tweak the CSS to make it responsive.

Nest the moderation routes to possibly speed routing up a little...?

Add a friend system.

Add more administration features.
