# pastebin

a simple and lightweight pastebin service written in go and templ

## Usage

```
$ ./pastebin

Usage: pastebin [-addr <port>] [-max-size <size>]

Options:
  -addr string
    	port to listen on (default ":3000")
  -max-size int
    	maximum size of a paste in bytes (default 32kB)
```

## Highlighting

To get syntax highlighting, you must add the file extension at the end of your paste URL: `/<paste_id>.<extension>`

Supported languages can be found [here](https://github.com/alecthomas/chroma/tree/master?tab=readme-ov-file#supported-languages)

### Themes
Themes can be applied by specifying in the URL: `/<paste_id>.<extension>/<theme>`

[List of available themes](https://github.com/alecthomas/chroma/tree/master/styles)
