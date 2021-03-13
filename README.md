# faq

[![Go Report Card](https://goreportcard.com/badge/github.com/jzelinskie/faq?style=flat-square)](https://goreportcard.com/report/github.com/jzelinskie/faq)
[![Build Status](https://github.com/jzelinskie/faq/workflows/Build%20&%20Test/badge.svg)](https://github.com/jzelinskie/faq/actions)
[![Godoc](https://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/jzelinskie/faq)
[![Releases](https://img.shields.io/github/release/jzelinskie/faq/all.svg?style=flat-square)](https://github.com/jzelinskie/faq/releases)
[![LICENSE](https://img.shields.io/github/license/jzelinskie/faq.svg?style=flat-square)](https://github.com/coreos/etcd/blob/master/LICENSE)

faq is a tool intended to be a more flexible [jq], supporting additional formats.
The additional formats are converted into JSON and processed with [libjq].

Supported formats:
- BSON
- Bencode
- JSON
- Property Lists
- TOML
- XML
- YAML

How do you pronounce faq? "Fuck you".

For example usage, read [the examples doc].

[releases]: https://github.com/jzelinskie/faq/releases
[jq]: https://github.com/stedolan/jq
[libjq]: https://github.com/stedolan/jq/wiki/C-API:-libjq
[the examples doc]: /docs/examples.md

## Installation

The `master` branch may be in an *unstable or even broken state* during development.
Please use [releases] instead of the `master` branch in order to get stable binaries.

Behavior such as command-line flags may change causing shell scripts using faq to break after upgrading.
jq programs are stable and should be considered a bug if it does not match jq behavior.

- Statically compiled binaries are available on the [releases] page: just download the binary for your platform, and make it executable.
- A [Homebrew] formula can be installed with `brew install jzelinskie/faq/faq`
- RPMs are available via a [COPR repository]. 
- There's an [AUR PKGBUILD] for Arch Linux that can be installed with your favorite [AUR tooling].

[Homebrew]: https://brew.sh
[COPR repository]: https://copr.fedorainfracloud.org/coprs/ecnahc515/faq
[AUR PKGBUILD]: https://aur.archlinux.org/packages/faq/
[AUR tooling]: https://wiki.archlinux.org/index.php/AUR_helpers

## Development

In order to compile the project, the [latest stable version of Go] and knowledge of a [working Go environment] are required.
A version of [jq] greater than 1.6-rc2 that includes the libjq header files must also be installed on the system.

```sh
git clone git@github.com:jzelinskie/faq.git
cd faq
make all
```

[latest stable version of Go]: https://golang.org/dl
[working Go environment]: https://golang.org/doc/code.html
[jq]: https://stedolan.github.io/jq

## License

faq is made available under the Apache 2.0 license.
See the [LICENSE](LICENSE) file for details.
