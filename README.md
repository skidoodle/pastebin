# pastebin

a simple and lightweight pastebin service written in go and templ

## Usage

```
$ ./pastebin -help

Usage of ./pastebin:
  -addr string
    	socket address to bind to (default ":3000")
  -db-path string
        path to the database file (default "pastebin.db")
  -max-size int
    	maximum size of a paste in bytes (default 32kB)
  -ttl duration
        time to live for pastes (e.g., 72h, 30m) (default 168h0m0s)
```

## Highlighting

To get syntax highlighting, you must add the file extension at the end of your paste URL: `/<paste_id>.<extension>`

Supported languages can be found [here](https://github.com/alecthomas/chroma/tree/master?tab=readme-ov-file#supported-languages)

### Themes
Themes can be applied by specifying in the URL: `/<paste_id>.<extension>/<theme>`

[List of available themes](https://github.com/alecthomas/chroma/tree/master/styles)
