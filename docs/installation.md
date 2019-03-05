# Windows Installation

Run `install.bat`, e.g. double-click on it. You will also have to start-up MySQL, which if you're using Wnmp or friends is just a matter of opening that program and starting the MySQL process via it.

Follow the instructions shown on the screen.

To navigate to the folder the software is in at any time in the future, you can just type `cd` followed by the folder's name, e.g. `cd /home/gosora/src/` and then you can run your commands. cd stands for change directory.


# Linux Simple Installation

Simple installations are usually recommended for trying out the software rather than for deploying it in production as they are less hardened and have fewer service facilities.

This might also be fine, if you're using something else as a reverse-proxy (e.g. Nginx or Apache).

First, we need somewhere for the software to live, if you're familiar with Linux, then you might have some ideas of your own, otherwise we may just go for `~/gosora`.

First, we'll navigate to our home folder by typing: `cd ~`

And then, we'll going to pull a copy of Gosora off the git server with: `git clone https://github.com/Azareal/Gosora`

And now, we're going to rename the downloaded folder from Gosora to gosora because the uppercase letter bugs me with: `mv Gosora gosora`

We can now hop into that folder with the same command we used for getting to the home folder:

`cd gosora`

And now, we'll change the permissions on the installer script, otherwise we'll get an access denied error:

`chmod 755 ./install-linux`

Just run this to run the installer:

`./install-linux`

Follow the instructions shown on the screen.


# Linux Installation with Systemd Service

You will need administrator privileges on the machine (aka root) to add a service.

First, you will need to jump to the place where you want to put the code, we will use `/home/gosora/src/` here, but if you want to use something else, then you'll have to modify the service file with your own path (but *never* in a folder where the files are automatically served by a webserver).

If you place it in `/www/`, `/public_html/` or any similar folder, then there's a chance that your server might be compromised.

The following commands will pull the latest copy of Gosora off the Git repository, will create a user account to run Gosora as, will set it as the owner of the files and will start the installation process.

If you're just casually setting up an installation on your own machine which isn't exposed to the internet just to try out Gosora, then you might not need to setup a seperate account for it or do `chmod 2775 logs`.

Please type the following commands into the console and hit enter:

`cd /home/`

`useradd gosora`

`passwd gosora`

Type in a strong password for the `gosora` user, please oh please... Don't use "password", just... don't, okay? Also, you might want to note this down somewhere.

`mkdir gosora`

`cd gosora`

`git clone https://github.com/Azareal/Gosora`

`mv Gosora src`

`chown -R gosora ../gosora`

`chgrp -R www-data ../gosora`

`cd src`

`chmod 2775 logs`

`chmod 2775 uploads`

`chmod 755 ./install-linux`

`./install-linux`

Follow the instructions shown on the screen.

We will also want to setup a service:

`chmod 755 ./pre-run-linux`

`cp ./gosora_example.service /lib/systemd/system/gosora.service`

`systemctl daemon-reload`


# Additional Configuration

For things like HTTPS, you might also need to [modify your config.json](https://github.com/Azareal/Gosora/blob/master/docs/configuration.md) file after installing Gosora to get it working.

You can get a free private key and certificate pair from Let's Encrypt or Cloudflare.
If you're using Nginx or something else as a reverse-proxy in-front of Gosora, then you will have to consult their documentation for advice on setting HTTPS.


# Advanced Installation

This section explains how to set things up without running the batch or shell files. For Windows, you will likely have to open up cmd.exe (the app called Command Prompt in Win10) to run these commands inside or something similar, while with Linux you would likely use the Terminal or console.

Linux is similar, however you might need to use cd and mv a bit more like in the shell files due to the differences in go build across platforms. Additionally, Linux doesn't require `StackExchange/wmi` or `/x/sys/windows`

You also need to substitute the `gosora.exe` bits for `./Gosora` on Linux. For more info, you might want to take a gander inside the `./run-linux` and `./install-linux` shell files to see how they're implemented.

```bash
git clone https://github.com/Azareal/Gosora

go get

rm -f template_*.go

rm -f gen_*.go

rm -f tmpl_client/template_*.go

rm -f ./Gosora

go generate

go build ./router_gen

router_gen.exe

go build ./cmd/query_gen

query_gen.exe

go build -o gosora.exe

go build "./cmd/install"

install.exe

go get github.com/mailru/easyjson/...

easyjson -pkg common

gosora.exe -build-templates

gosora.exe
```

I'm looking into minimising the number of go gets for the advanced build and to maybe remove the platform and database engine specific dependencies if possible for those who don't need them.

If systemd gives you no permission errors, then make sure you `chown`, `chgrp` and `chmod` the files and folders appropriately.
