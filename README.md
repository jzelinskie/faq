# faq

faq is a tool intended to be a drop in replacement for "jq", but supports additional formats.
The additional formats are converted into JSON and processed with libjq.

faq is pronounced "fah queue".

Supported formats:
- BSON
- Bencode
- JSON
- TOML
- XML
- YAML

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

## Examples

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

### Viewing the non-binary parts of a torrent file

```sh
faq 'del(.info.pieces)' ubuntu.torrent
```
```json
{
  "announce": "http://torrent.ubuntu.com:6969/announce",
  "announce-list": [
    [
      "http://torrent.ubuntu.com:6969/announce"
    ],
    [
      "http://ipv6.torrent.ubuntu.com:6969/announce"
    ]
  ],
  "comment": "Ubuntu CD releases.ubuntu.com",
  "creation date": 1515735480,
  "info": {
    "length": 1502576640,
    "name": "ubuntu-17.10.1-desktop-amd64.iso",
    "piece length": 524288
  }
}
```

## License

faq is made available under the Apache 2.0 license.
See the [LICENSE](LICENSE) file for details.
