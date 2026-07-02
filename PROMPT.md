
add the quiet mode option:
changes:

struct GenerationProps:
 - bool quiet  (default to false)


if the user pass the --quiet flag, and quiet is true on user props, then no output should be printed to the console

