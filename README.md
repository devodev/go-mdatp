A CLI as well as a library to interact with the Microsoft Defender ATP REST API.

## Overview
`go-mdatp` provides a client library for the `Microsoft Defender ATP REST API` written in [Go](https://golang.org/). It follows the Microsoft API Reference available [here](https://docs.microsoft.com/en-us/windows/security/threat-protection/microsoft-defender-atp/pull-alerts-using-rest-api).

`go-mdatp` is also a CLI application with everything you need to interact with the API on the command line.

Currently, **`go-mdatp` requires Go version 1.13 or greater**.

#### Supported Architectures
We provide pre-built go-mdatp binaries for Windows, Linux and macOS (Darwin) architectures, in both 386/amd64 flavors.</br>
Please see the release section [here](https://github.com/devodev/go-mdatp/releases).

## Table of Contents

- [Overview](#overview)
  - [Supported Architectures](#supported-architectures)
- [Get Started](#get-started)
  - [Build](#build)

## Get Started
`go-mdatp` uses Go Modules introduced in Go 1.11 for dependency management.

### Build
Build the CLI for a target platform (Go cross-compiling feature), for example linux, by executing:
```
$ mkdir $HOME/src
$ cd $HOME/src
$ git clone https://github.com/devodev/go-mdatp.git
$ cd go-mdatp
$ env GOOS=linux go build -o go_mdatp_linux ./cmd/go-mdatp
```
If you are a Windows user, substitute the $HOME environment variable above with %USERPROFILE%.
