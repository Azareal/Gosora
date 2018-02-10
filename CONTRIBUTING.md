We're not accepting contributions right now, although you're welcome to poke me about things. I'd like to put a process together at some point.

# Coding Standards

All code must be unit tested where ever possible with the exception of JavaScript which is untestable with our current technologies, tread with caution there.

# Golang

Use the standard linter and listen to what it tells you to do.

The route assignments in main.go are *legacy code*, add new routes to `router_gen/routes.go` instead.

Try to use the single responsibility principle where ever possible, with the exception for if doing so will cause a large performance drop. In other words, don't give your interfaces / structs too many responsibilities, keep them simple.

Avoid hand-rolling queries. Use the builders, a ready built statement or a datastore structure instead. Preferably a datastore.

More coming up.

# JavaScript

Use semicolons at the end of statements. If you don't, you might wind up breaking a minifier or two.

Always use strict mode.

Don't worry about ES5, we're targetting modern browsers. If we decide to backport code to older browsers, then we'll transpile the files.

To keep consistency with Go code, variables must be camelCase.

# JSON

To keep consistency with Go code, map keys must be camelCase.