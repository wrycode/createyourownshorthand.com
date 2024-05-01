#!/usr/bin/env bash
GOARCH=wasm GOOS=js go build -o static/main.wasm main.go render.go
hugo

# to serve content locally, you can use:
# goexec 'http.ListenAndServe(`:8997`, http.FileServer(http.Dir(`./public`)))'
