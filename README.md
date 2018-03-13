# faq

format agnostic querier, or faq (pronounced fah queue), attempts to be a drop in replacement for jq that supports multiple formats.

faq converts the alternative formats into JSON and uses libjq to process them.

example usage:

```sh
$ faq -f bencode '.announce' ~/Downloads/ubuntu_iso.torrent
"http://torrent.ubuntu.com:6969/announce"
```
