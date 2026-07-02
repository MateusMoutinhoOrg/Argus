
add the quiet mode:
changes:

struct GenerationProps:
 - bool quiet  (default to false)


 changes when quiet is true:
 no menssages, both by the UserCallback or by the engine should be plotted
 