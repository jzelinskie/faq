# Examples

## Usage

```
Usage:
  faq [flags] [filter string] [files...]

Flags:
  -a, --ascii-output        force output to be ascii instead of UTF-8
  -C, --color-output        colorize the output (default true)
  -c, --compact             compact instead of pretty-printed output
  -f, --format string       input format (default "auto")
  -h, --help                help for faq
  -M, --monochrome-output   monochrome (don't colorize the output)
  -r, --raw                 output raw strings, not JSON texts
  -S, --sort-keys           sort keys of objects on output
  -t, --tab                 use tabs for indentation
```

## Command-line fu


### Piping to make something legible

```sh
echo '{"hello":{"world":"whats up"},"with":"you"}' | faq
```

```json
{
  "hello": {
    "world": "whats up"
  },
  "with": "you"
}

```

### Reading a raw string value from a YAML file

```sh
faq -r '.apiVersion' etcdcluster.yaml
```
```
etcd.database.coreos.com/v1beta2
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
