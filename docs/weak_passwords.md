# Weak Passwords

For configuring the list of weak passwords and weak password detection rules, we have `config/weakpass.json` which overwrites the default values defined in `config/weakpass_default.json`

There are two sections: `contains` and `literal`. `contains` scans the password to see if a specified piece of text is in it and `literal` checks if the password matches the specified rule exactly (with some exceptions).

All passwords are converted to lowercase form before either scanner is ran on them to detect common tricks like capitalizing the first letter.

`contains` is slower and may not scale with a large number of rules, but it is more effective at finding certain patterns which a password cracker could exploit to crack someone's password.

`literal` is very inflexible and only matches rules literally. One exception is that it will remove numbers from the end of the password running the rule.
