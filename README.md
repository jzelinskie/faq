# faq

[![Go Report Card](https://goreportcard.com/badge/github.com/jzelinskie/faq?style=flat-square)](https://goreportcard.com/report/github.com/jzelinskie/faq)
[![Build Status Travis](https://img.shields.io/travis/jzelinskie/faq.svg?style=flat-square&&branch=master)](https://travis-ci.org/jzelinskie/faq)
[![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/jzelinskie/faq)
[![Releases](https://img.shields.io/github/release/jzelinskie/faq/all.svg?style=flat-square)](https://github.com/jzelinskie/faq/releases)
[![LICENSE](https://img.shields.io/github/license/jzelinskie/faq.svg?style=flat-square)](https://github.com/coreos/etcd/blob/master/LICENSE)

**Note**: The `master` branch may be in an *unstable or even broken state* during development. Please use [releases](https://github.com/jzelinskie/faq/releases) instead of the `master` branch in order to get stable binaries.

faq is a tool intended to be a more flexible "jq", supporting additional formats.
The additional formats are converted into JSON and processed with libjq.

Supported formats:
- BSON
- Bencode
- JSON
- TOML
- XML
- YAML

How do you pronounce "faq"? Fuck you.

For example usage, read [the examples doc].

[the examples doc]: /docs/examples.md

## Installation

faq is still under heavy development and has only unstable binary releases.
Behavior such as command-line flags may change causing shell scripts using faq to break after upgrading.
jq programs are stable and should be considered a bug if it does not match jq behavior.

### Linux

```sh
curl -Lo /usr/local/bin/faq https://github.com/jzelinskie/faq/releases/download/0.0.5/faq-linux-amd64
chmod +x /usr/local/bin/faq
```

### macOS

```sh
brew tap jzelinskie/faq
brew install faq
```

## Development

In order to compile the project, the [latest stable version of Go] and knowledge of a [working Go environment] are required.
A version of [jq] greater than 1.6-rc2 that includes the libjq header files must also be installed on the system.

```sh
git clone git@github.com:jzelinskie/faq.git
cd faq
make
```

[latest stable version of Go]: https://golang.org/dl
[working Go environment]: https://golang.org/doc/code.html
[jq]: https://stedolan.github.io/jq

## License

faq is made available under the Apache 2.0 license.
See the [LICENSE](LICENSE) file for details.
