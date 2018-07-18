# Usage

```
Usage:
  faq [flags] [filter string] [files...]

Flags:
  -c, --color-output           colorize the output (default true)
  -h, --help                   help for faq
  -f, --input-format string    input format (default "auto")
  -m, --monochrome-output      monochrome (don't colorize the output)
  -o, --output-format string   output format (default "auto")
  -p, --pretty-output          pretty-printed output (default true)
  -r, --raw-output             output raw strings, not JSON texts
```

# Examples

Note: Example files can be found in the /docs/examples/ directory

## Piping to make something legible

Input
```sh
cat docs/examples/unformatted.json | faq
```

Output
```json
{
  "hello": {
    "world": "whats up"
  },
  "with": "you"
}
```

## Reading a raw string value from a YAML file

Input
```sh
faq -r '.company' docs/examples/sample.yaml
```

Output
```
Awesome Code inc.
```

### Get the name of all of the dependencies of a Go project

```sh
faq -r '.constraint[].name' Gopkg.toml
```

```
github.com/Azure/draft
github.com/BurntSushi/toml
github.com/ashb/jqrepl
github.com/clbanning/mxj
github.com/ghodss/yaml
github.com/globalsign/mgo
github.com/sirupsen/logrus
github.com/spf13/cobra
github.com/zeebo/bencode
golang.org/x/crypto
```

### Viewing the non-binary parts of a torrent file

```sh
curl --silent -L https://cdimage.debian.org/debian-cd/current/amd64/bt-cd/debian-9.4.0-amd64-netinst.iso.torrent | faq -f bencode 'del(.info.pieces)'
```

```json
{
  "announce": "http://bttracker.debian.org:6969/announce",
  "comment": "\"Debian CD from cdimage.debian.org\"",
  "creation date": 1520682848,
  "httpseeds": [
    "https://cdimage.debian.org/cdimage/release/9.4.0//srv/cdbuilder.debian.org/dst/deb-cd/weekly-builds/amd64/iso-cd/debian-9.4.0-amd64-netinst.iso",
    "https://cdimage.debian.org/cdimage/archive/9.4.0//srv/cdbuilder.debian.org/dst/deb-cd/weekly-builds/amd64/iso-cd/debian-9.4.0-amd64-netinst.iso"
  ],
  "info": {
    "length": 305135616,
    "name": "debian-9.4.0-amd64-netinst.iso",
    "piece length": 262144
  }
}
```
