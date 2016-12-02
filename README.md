# Grosolo

A super fast forum software written in Go.

The initial code-base was forked from one of my side projects, and converted from the web framework it was using.



# Dependencies

Go. The programming language this program is written in, and the compiler which it requires. You will need to install this.

MySQL Database. You will need to setup a MySQL Database somewhere. A MariaDB Database works equally well, and is much faster than MySQL.


# Installation Instructions

**Run the following commands:**

go install github.com/go-sql-driver/mysql

go install golang.org/x/crypto/bcrypt

go run errors.go main.go pages.go post.go routes.go topic.go user.go utils.go


The last command is run whenever you want to start-up an instance of the software.

# TO-DO

Oh my, you caught me right at the start of this project. There's nothing to see here yet, asides from the absolute basics. You might want to look again later!


More moderation features.

Fix the login system. It broke during the switch in frameworks.

Add an alert system.

Add a report feature.

Add a complex permissions system.

Add a settings system.

Add a plugin system.
