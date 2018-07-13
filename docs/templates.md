# Templates

Gosora uses a subset of [Go Templates](https://golang.org/pkg/text/template/) which are run on both the server side and client side with custom transpiler to wring out the most performance. Some more obscure features may not be available (e.g. local variables), but I am adding them in here and there.

The base templates are stored in `/templates/` and you can shadow them by placing modified duplicates in `/templates/overrides/`. The default themes all share the same set of templates present there.

# Non-standard Extensions

We also have a few non-standard extensions only available on certain pages or areas, but these shouldn't be relied on in favour of more general mechanisms.

More to come soon.