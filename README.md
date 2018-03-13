# faq

format agnostic querier, or faq (pronounced fah queue), is a jq clone supporting multiple formats.

example usage:

```sh
$ faq -f bencode '.announce' ~/Downloads/ubuntu_iso.torrent
"http://torrent.ubuntu.com:6969/announce"
```
