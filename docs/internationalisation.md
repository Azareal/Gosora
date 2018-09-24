# Internationalisation

Internationalisation is one of Gosora's top priorities, although not the only one. This means making it possible for administrators to translate the interface to their native language and to otherwise customise to fit their cultures.

This doesn't mean that the software will cover every language on Earth right off the bat, however although, it would be greatly appreciated if people were to help contribute towards making this a reality.

Quite a large portion of the software has been internationalised, although some of the biggest exceptions are a small handful of complex phrases which are currently hard to localise, the error messages and a few phrases in the Control Panel.

These exceptions are going to be resolved one by one as I go along until everything is covered, although it should be noted that some languages, particularly right-to-left ones might be hard to localise without a custom theme, but we'll look into seeing what we can do about those.

# Customising Phrases

You can add a custom language to Gosora by adding a new file to the `/langs/` folder with the name of the language followed by the extension `.json`, e.g. `spanish.json` or `espanol.json`, if that is what you prefer, although the first might be more obvious to a larger audience.

You can also customise the phrases by doing the same thing and naming the file `english_custom.json` or whatever you wish. The contents of the file should basically follow the same format as in `english.json` and it should be noted that new phrases may be added to and removed from that file from time to time.

You can then set the default language for your site by going into `/config/config.json` and changing `"Language": "english"` to `"Language": "spanish"` or whatever the name of the language is. This value takes the value of the Name field in the language file and not the file name, although I would advise using unique names there and perhaps the name of the file for consistency.
