# Updating Gosora (Windows)

The update system is currently under development, but you can run `dev-update.bat` to update your instance to the latest commit and to update the associated database schema, etc.

If you run into any issues doing so, please open an issue: https://github.com/Azareal/Gosora/issues/new

If you want to manually patch Gosora rather than relying on the above scripts to do it, you'll first want to save your changes with `git stash`, and then, you'll overwrite the files with the new ones with `git pull origin master`, and then, you can re-apply your custom changes with `git stash apply`

After that, you'll need to run `go build ./patcher`.

Once you've done that, you just need to run `patcher.exe` to apply the latest patches to the database, etc.

# Updating a software with a simple installation (Linux)

The update system is currently under development, but you can run `dev-update-linux` to update your instance to the latest commit and to update the associated database schema, etc.

If you run into any issues doing so, please open an issue: https://github.com/Azareal/Gosora/issues/new

If you want to manually patch Gosora rather than relying on the above scripts to do it, you'll first want to save your changes with `git stash`, and then, you'll overwrite the files with the new ones with `git pull origin master`, and then, you'll re-apply your changes with `git stash apply`.

After that, you'll need to run `go build -o Patcher "./patcher"`

Once you've done that, you just need to run `./Patcher` to apply the latest patches to the database, etc.


# Updating a software using systemd (Linux)

You will first want to follow the instructions in the section for updating dependencies.

The update system is currently under development, but you can run `quick-update-linux` in `/home/gosora/src`to update your instance to the latest commit and to update the associated database schema, etc.

If you run into any issues doing so, please open an issue: https://github.com/Azareal/Gosora/issues/new

If you're using a systemd service, then you might want to switch to the `gosora` user with `su gosora` (you may be prompted for the password to the user), you can switch back by typing `exit`.
If this is the first time you've done an update as the `gosora` user, then you might have to configure Git, simply do:

`git config --global user.name "Lalala"`
`git config --global user.email "lalala@example.com"`

Replace that name and email with whatever you like. This name and email only applies to the `gosora` user. If you see a zillion modified files pop-up, then that is due to you changing their permissions, don't worry about it.

If you get an access denied error, then you might need to run `chown -R gosora /home/gosora` and `chgrp -R www-data /home/gosora` to fix the ownership of the files.

If you want to manually patch Gosora rather than relying on the above scripts to do it, you'll first want to save your changes with `git stash`, and then, you'll overwrite the files with the new ones with `git pull origin master`, and then, you'll re-apply your changes with `git stash apply`.

After that, you'll need to run `go build -o Patcher "./patcher"`

Once you've done that, you just need to run `./Patcher` to apply the latest patches to the database, etc.


# Updating Dependencies

Dependencies are third party scripts and programs which Gosora relies on to function. The instructions here do not cover updating MySQL / MariaDB or Go.

You can update them by running the `go get` command.

You'll need to restart the server after you change a template or update Gosora, e.g. with `run.bat` or killing the process and running `./run-linux` or via `./pre-run-linux` followed by `service gosora restart`.
