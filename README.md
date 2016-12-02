# Grosolo

A super fast forum software written in Go.

The initial code-base was forked from one of my side projects, and converted from the web framework it was using.


# Features
Basic Forum Functionality

Custom Pages


# Dependencies

Go. The programming language this program is written in, and the compiler which it requires. You will need to install this.

MySQL Database. You will need to setup a MySQL Database somewhere. A MariaDB Database works equally well, and is much faster than MySQL.


# Installation Instructions

**Run the following commands:**

go install github.com/go-sql-driver/mysql

go install golang.org/x/crypto/bcrypt

Tweak the config.go file and put your database details in there. Import data.sql into the same database. Comment out the first line (put /* and */ around it), if you've already made a database, and don't want the script to generate it for you.

Set the password column of your user account in the database to what you want your password to be. The system will encrypt your password when you login for the first time.


# Run the program

go run errors.go main.go pages.go reply.go routes.go topic.go user.go utils.go config.go

Alternatively, you could run the run.bat batch file on Windows.


# TO-DO

Oh my, you caught me right at the start of this project. There's nothing to see here yet, asides from the absolute basics. You might want to look again later!


More moderation features.

Fix the bug where errors are sent off in raw HTML rather than formatted HTML.

Fix the custom pages.

Add emails as a requirement for registration and add a simple anti-spam measure.

Add an alert system.

Add a report feature.

Add a complex permissions system.

Add a settings system.

Add a plugin system.

Revamp the system for serving static files to make it much faster.

Tweak the CSS to make it responsive.
