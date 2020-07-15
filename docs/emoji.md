# Emoji

Emojis are a work in progress. We plan to implement UIs to input and pick them easily in the future. We also plan to make them more visually appealing and consistent.

Right now, you can input the emoji directly with an emoji keyboard, like those present on mobile devices or Windows 10 (WIN KEY + . OR WIN KEY + ;).

You can also input them if you know the shortcodes which are defined in the configuration files `config/emoji_default.json` or `config/emoji.json`. `emoji_default.json` should be left untouched as it gets updated with Gosora. Any additions should be made via `emoji.json`.

The file is a work in progress but it contains two sections: `no_defaults` and `emojis`.

`no_defaults` wipes clear anything defined by `emoji_default.json` to start over with a clean slate. You can set this to true to enable it or leave it absent to disable it.

`emojis` lets you create a list of emoji shortcodes which are substituted by the parser for Unicode emojis. You should only specify shortcodes which start and end with `:` here. We plan to add a section for literals like `:P` in the future.