# Static site generator

Run from this directory with Go 1.26.1 or later:

```sh
export GITHUB_TOKEN=... # Read-only token; public repository access is sufficient.
go run .
```

The generator reads `../extensions.json` and copies the source assets from
`../static/` into the ignored build directory `output/` inside `generator/`.