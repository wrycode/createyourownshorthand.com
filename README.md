Free instruction manual and software tool for creating your own shorthand system. 

Most information can be found on the website itself, visit it
[here](https://createyourownshorthand.com/). This README contains
development information.


# Development

The site is built using Hugo, so you can clone the repository and run `hugo build` or `hugo serve` to make most changes (e.g. to the survey or survey javascript, the html or images).

The shorthand application itself won't work with `hugo serve` because the server uses the wrong mimetype for the webassembly. Instead use:

```
$ goexec 'http.ListenAndServe(`:8997`, http.FileServer(http.Dir(`./public`)))'
```

You MUST use port 8997 because that is how CORS is configured for the or the IPA dictionary resource.

If you update the Go code, remember to recompile the webassembly:
```
$ GOARCH=wasm GOOS=js go build -o static/main.wasm main.go render.go
```

If you want to use the backend rendering code without the website, edit `options.json` directly and run the code with `go run command.go render.go`.


# TODO
- add known issues
- rewrite readme with more context and full development guide
