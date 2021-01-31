<p align="center">
  <img src=".github/logo.png" alt="Logo">
</p>
<p align="center">
  <a href="https://github.com/AlexGustafsson/larch/blob/master/go.mod">
    <img src="https://shields.io/github/go-mod/go-version/AlexGustafsson/larch" alt="Go Version" />
  </a>
  <a href="https://github.com/AlexGustafsson/larch/releases">
    <img src="https://flat.badgen.net/github/release/AlexGustafsson/larch" alt="Latest Release" />
  </a>
  <br>
  <strong><a href="#quickstart">Quick Start</a> | <a href="#contribute">Contribute</a> </strong>
</p>

# Larch
### A self-hosted service and toolset for managing, archiving, viewing and sharing bookmarks

Note: Larch is currently being actively developed. Until it reaches v1.0.0 breaking changes may occur in minor versions.

Larch is a new service and all-around tool for managing, archiving, viewing and sharing bookmarks. It builds on one novel idea - Larch will one day become obsolete, but its archives must not.

That is, Larch is designed from the ground up knowing that it may one day cease to work. Whether this is due to radically new CPU architectures, deprecation of the binary format or any other reason, Larch will ensure that you may keep your archives for a long time to come.

Until then, however, Larch comes with some great features centered around a series of tools and APIs.

As a tool, Larch enables you to easily create archives of websites (bookmarks) and convert them between several formats such as WARC (archive.org) and WebArchive (Safari) right from your CLI.

As a service and API, Larch enables you to create and manage archives, bookmarks and share them via a lightweight server. It is built with complete control and extensibility in mind, centered around its API and a pluggable architecture.

## Quickstart
<a name="quickstart"></a>

Upcoming.

## Table of contents

[Quickstart](#quickstart)<br/>
[Features](#features)<br />
[Installation](#installation)<br />
[Usage](#usage)<br />
[Contributing](#contributing)

<a id="features"></a>
## Features

* Plugin-driven, extensible architecture (upcoming)
* Automatic backups (upcoming)
* Link monitoring (upcoming)
* Archiving of websites
* Fully controllable via APIs (upcoming)
* Supports [WARC](https://github.com/internetarchive/heritrix3/wiki/WARC%20%28Web%20ARChive%29) archives (ISO 28500:2017)
* Supports [WebArchive](https://en.wikipedia.org/wiki/Webarchive)
* Supports TOR (upcoming)
* Supports IPFS (upcoming)
* Supports Encrypted Archives (upcoming)

<a id="installation"></a>
## Installation

### Using Homebrew

Upcoming.

```sh
brew install alexgustafsson/tap/larch
```

### Downloading a pre-built release

Download the latest release from [here](https://github.com/AlexGustafsson/larch/releases).

### Build from source

Clone the repository.

```sh
git clone https://github.com/AlexGustafsson/larch.git && cd larch
```

Optionally check out a specific version.

```sh
git checkout v0.1.0
```

Build the application.

```sh
make build
```

## Usage
<a name="usage"></a>

_Note: This project is still actively being developed. The documentation is an ongoing progress._

```
Usage: larch [global options] command [command options] [arguments]

A service for managing, archiving, viewing and sharing bookmarks

Version: v0.1.0, build . Built Fri Jan 22 21:08:46 CET 2021 using go version go1.15.6 darwin/amd64

Options:
  --verbose   Enable verbose logging (default: false)
  --help, -h  show help (default: false)

Commands:
  version  Show the application's version
  help     Shows a list of commands or help for one command
```

## Documentation

### WARC

WARC 1.0 is implemented according to the ISO 28500:2017 draft available here: [http://bibnum.bnf.fr/WARC/WARC_ISO_28500_version1_latestdraft.pdf](http://bibnum.bnf.fr/WARC/WARC_ISO_28500_version1_latestdraft.pdf).

Resources:
* http://fileformats.archiveteam.org/wiki/WARC
* https://iipc.github.io/warc-specifications/specifications/warc-format/warc-1.1/

## Contributing
<a name="contributing"></a>

Any help with the project is more than welcome. The project is still in its infancy and not recommended for production.

### Development

```sh
# Clone the repository
https://github.com/AlexGustafsson/larch.git && cd larch

# Show available commands
make help

# Build the project for the native target
make build
```

_Note: due to a bug (https://gcc.gnu.org/bugzilla/show_bug.cgi?id=93082, https://bugs.llvm.org/show_bug.cgi?id=44406, https://openradar.appspot.com/radar?id=4952611266494464), clang is required when building for macOS. GCC cannot be used. Build the server like so: `CC=clang make server`._

### License

The Larch logo was created by Amanda Svensson and is licensed under [Creative Commons Attribution-NonCommercial-NoDerivs 3.0 Unported (CC BY-NC-ND 3.0)](https://creativecommons.org/licenses/by-nc-nd/3.0/).
