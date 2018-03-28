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

For example usage, read [the examples doc].

[the examples doc]: /docs/examples.md

## Installation

faq is still under heavy development and has yet to make a binary release.
Please follow the development instructions build your own binary.

## Development

In order to compile the project, the [latest stable version of Go] and knowledge of a [working Go environment] are required.
A modern version of [jq] that includes the libjq header files, must also be installed on the system.
Reproducible builds are handled by using [dep] to vendor dependencies.

```sh
mkdir faq && export GOPATH=$PWD/faq
git clone git@github.com:jzelinskie/faq.git $GOPATH/src/github.com/jzelinskie/faq
cd $GOPATH/src/github.com/jzelinskie/faq
dep ensure
go install github.com/jzelinskie/faq
$GOPATH/bin/faq --help
```

[latest stable version of Go]: https://golang.org/dl
[working Go environment]: https://golang.org/doc/code.html
[jq]: https://stedolan.github.io/jq
[dep]: https://github.com/golang/dep

## License

faq is made available under the Apache 2.0 license.
See the [LICENSE](LICENSE) file for details.
