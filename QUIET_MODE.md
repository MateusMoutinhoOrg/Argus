implement the quiet mode for the cli apps created with argus

add the following changes:
- add the property quiet in the struct Deps(pkg/deps/deps.go)
- implement in native/native/native.go in the  unction print, to only print if quiet is false.
- add in the GenerationProps a propety  bool quiet(defaults false)  if is true and the flag --quiet or -q is present it should change deps.r to true.

The hole ideia ,is that if quiet is  true , no print should be displayed on the terminal
