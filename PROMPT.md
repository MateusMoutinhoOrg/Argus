
add the quiet mode option:
changes:

struct GenerationProps:
 - bool quiet  (default to false)

struct Deps:
 - bool quiet (injected by HandleCli) 


adapters/native/native.go -> Print
if quiet is true, dont print


if the user pass the --quiet flag, and quiet is true on user props, then no output should be printed to the console



