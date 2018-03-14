# faq

format agnostic querier, or faq (pronounced fah queue), attempts to be a drop in replacement for jq that supports multiple formats.

faq converts the alternative formats into JSON and uses libjq to process them.

example usage:

```sh
$ echo '{"hello": "world"}' | faq
{
  "hello": "world"
}

$ faq -f bencode 'del(.info.pieces)' ubuntu.torrent
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
