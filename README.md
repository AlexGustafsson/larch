<!--<p align="center">
  <img src=".github/banner.png" alt="Banner">
</p>-->
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
### A self-hosted service for managing, archiving, viewing and sharing bookmarks

Note: Larch is currently being actively developed. Until it reaches v1.0.0 breaking changes may occur in minor versions.

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

* Plugin-driven, extensible architecture
* Automatic backups
* Link monitoring
* Archiving of links
* Fully controllable via APIs

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
