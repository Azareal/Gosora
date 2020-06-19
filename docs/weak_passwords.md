# Weak Passwords

For configuring the list of weak passwords and weak password detection rules, we have `config/weakpass.json` which overwrites the default values defined in `config/weakpass_default.json`

There are two sections: `contains` and `literal`. `contains` scans the password to see if a specified piece of text is in it and `literal` checks if the password matches the specified rule exactly (with some exceptions).

`contains` is slower and may not scale with a large number of rules, but it is more effective at finding certain patterns which a password cracker could exploit to crack someone's password.

`literal` is very inflexible and only matches rules literally. With two exceptions, the password fed to it is in lowercase form, so common variants like capitalizing the first letter will be detected. Sticking a number at the end of the common literal will also be detected.
