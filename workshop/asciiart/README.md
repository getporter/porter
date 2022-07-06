# ASCII Art

Make a bundle for the https://github.com/stdupp/goasciiart tool. Use it to
convert cute pictures of gophers into ASCII art when you install the bundle.

Here are some hints so that you can try to solve it in your own way. 
For the full solution, see the [asciiart][asciiart] directory in the workshop materials.

* A good base image for go is `golang:1.11-stretch`.
* You need to run `porter build` after modifying the Dockerfile.tmpl to rebuild
your invocation image to pick up your changes.
* Don't forget to copy your images into your invocation image to /cnab/app/.
* The command to run is `goasciiart -p=gopher.png -w=100`.

[asciiart]: https://getporter.org/src/workshop/asciiart

## Gophers!
[Gopher artwork][gophers] is copyright Ashley McNamara and is licensed under the 
[Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License][cc]. 

[gophers]: https://github.com/ashleymcnamara/gophers
[cc]: http://creativecommons.org/licenses/by-nc-sa/4.0/