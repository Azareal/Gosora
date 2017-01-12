# Gosora

A super fast forum software written in Go.

The initial code-base was forked from one of my side projects, but has now gone far beyond that.

Azareal's Discord Chat: https://discord.gg/eyYvtTf

If you like this software, please give it a star and give us some feedback :)

If you dislike it, please give us some feedback on how to make it better! We're always looking for feedback. We love hearing your opinions. If there's something missing or something doesn't look quite right, don't worry! We plan to change many things before the alpha!


# Features
Basic Forum Functionality. All of the little things you would expect of any forum software. E.g. Moderation, Custom Themes, Avatars, and so on.

Custom Pages. Under development. The Control Panel portion is incomplete, but you can create them by hand today. They're basically html/templates templates in the /pages/ folder.

Emojis. Allow your users to express themselves without resorting to serving tons upon tons of image files.

In-memory static file, forum and group caches. We're pondering over extending this solution over to topics, users, etc.

A profile system including profile comments and moderation tools for the profile owner.

A template engine which compiles templates down into machine code. Over ten times faster than html/templates. Compatible with templates written for html/templates, you don't need to learn any new templating language.

A plugin system. More on this to come.

A responsive design. Looks great on mobile phones, tablets, laptops, desktops and more!


# Dependencies

Go 1.7. You will need to install this. Pick the .msi, if you want everything sorted out for you rather than having to go around updating the environment settings. https://golang.org/doc/install

MySQL Database. You will need to setup a MySQL Database somewhere. A MariaDB Database works equally well and is much faster than MySQL.
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

Add a modern alert system.

Add per-forum permissions to finish up the foundations of the permissions system.

Add a *better* plugin system. E.g. Allow for plugins written in Javascript and ones written in Go. Also, we need to add many, many, many more plugin hooks.

I will need to ponder over implementing an even faster router. We don't need one immediately, although it would be nice if we could get one in the near future. It really depends. Ideally, it would be one which can easily integrate with the current structure without much work, although I'm not beyond making some alterations to faciliate it, assuming that we don't get too tightly bound to that specific router.

Allow themes to define their own templates and to override core templates with their own.

Add a friend system.

Improve profile customisability.

Implement all the common BBCode tags in plugin_bbcode

Implement all the common Markdown codes in plugin_markdown

Add more administration features.

Add more features for improving user engagement. E.g. A like system. I have a few of these in mind, but I've been pre-occupied with implementing other features.

Add a widget system.

Add support for multi-factor authentication.

Add support for secondary emails for users.
