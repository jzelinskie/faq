# Examples

## Basic

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
faq -r '.constraint[].name' Gopkg.toml -o json
```

The output format when using raw with TOML must be in JSON, because valid TOML requires a top-level object.

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

## Advanced

### Viewing the non-binary parts of a torrent file

```sh
curl -s https://torrent.fedoraproject.org/torrents/Fedora-Workstation-Live-x86_64-28.torrent | faq -f bencode -o yaml 'del(.info.pieces)'
```

```yaml
announce: http://torrent.fedoraproject.org:6969/announce
creation date: 1525097038
info:
  files:
  - length: 1215
    path:
    - Fedora-Workstation-28-1.1-x86_64-CHECKSUM
  - length: 1787822080
    path:
    - Fedora-Workstation-Live-x86_64-28-1.1.iso
  name: Fedora-Workstation-Live-x86_64-28
  piece length: 262144
```

### Passing extra arguments as variables

```sh
faq -n -f json -o json --args '1234' --jsonargs '{"jsoninput": {"moreadvanced": true}}' --kwargs 'fizz=test1' --kwargs 'buzz=test2' --jsonkwargs 'fizzbuzz={"jsonwargs": "areuseful"}' '$ARGS, $ARGS.positional[1].jsoninput.moreadvanced, $fizz, $buzz, $fizzbuzz'
{
  "named": {
    "buzz": "test2",
    "fizz": "test1",
    "fizzbuzz": {
      "jsonwargs": "areuseful"
    }
  },
  "positional": [
    "1234",
    {
      "jsoninput": {
        "moreadvanced": true
      }
    }
  ]
}
true
"test1"
"test2"
{
  "jsonwargs": "areuseful"
}
```
