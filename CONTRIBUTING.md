# Contributing

First and foremost, if you want to add a contribution, you'll have to open a pull request and to sign the CLA (contributor level agreement).

It's mainly there to deal with any legal issues which may come our way and to switch licenses without having to track down every contributor who has ever contributed.

Some things we could do is commercial licensing for companies which are not authorised to use open source licenses or moving to a more permissive license, although I'm not too experianced in these matters, if anyone has any ideas, then feel free to put them forward.

Try to prefix commits which introduce a lot of bugs or otherwise has a large impact on the usability of Gosora with UNSTABLE.

If something seems to be strange, then feel free to bring up an alternative for it, although I'd rather not get hung up on the little details, if it's something which is purely a matter of opinion.

# Coding Standards

All code must be unit tested where ever possible with the exception of JavaScript which is untestable with our current technologies, tread with caution there.

Use tabs not spaces for indentation.

# Golang

Use the standard linter and listen to what it tells you to do.

The route assignments in main.go are *legacy code*, add new routes to `router_gen/routes.go` instead.

Try to use the single responsibility principle where ever possible, with the exception for if doing so will cause a large performance drop. In other words, don't give your interfaces / structs too many responsibilities, keep them simple.

Avoid hand-rolling queries. Use the builders, a ready built statement or a datastore structure instead. Preferably a datastore.

Commits which require the patcher / update script to be run should be prefixed with "Database Changes: "

More coming up.

# JavaScript

Use semicolons at the end of statements. If you don't, you might wind up breaking a minifier or two.

Always use strict mode.

Don't worry about ES5, we're targetting modern browsers. If we decide to backport code to older browsers, then we'll transpile the files.

Please don't use await. It incurs too much of a cognitive overhead as to where and when you can use it. We can't use it everywhere quite yet, which means that we really should be using it nowhere.

Please don't abuse `const` just to shave off a few nanoseconds. Even in the Go server where I care about performance the most, I don't use const everywhere, only in about five spots in thirty thousand lines and I don't use it for performance at all there.

To keep consistency with Go code, variables must be camelCase.

# JSON

To keep consistency with Go code, map keys must be camelCase.

# Phrases

Try to keep the name of the phrase close to the actual phrase in english to make it easier for localisers to reason about which phrase is which.
