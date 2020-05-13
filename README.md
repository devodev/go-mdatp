# go-mdatp

A CLI as well as a library to interact with the Microsoft Defender ATP REST API.

## Overview

`go-mdatp` provides a client library for the `Microsoft Defender ATP REST API` written in [Go](https://golang.org/). It follows the Microsoft API Reference available [here](https://docs.microsoft.com/en-us/windows/security/threat-protection/microsoft-defender-atp/pull-alerts-using-rest-api).

`go-mdatp` is also a CLI application with everything you need to interact with the API on the command line.

Currently, **`go-mdatp` requires Go version 1.13 or greater**.

### Supported Architectures

We provide pre-built go-mdatp binaries for Windows, Linux and macOS (Darwin) architectures, in both 386/amd64 flavors.</br>
Please see the release section [here](https://github.com/devodev/go-mdatp/releases).

## Table of Contents

- [Overview](#overview)
  - [Supported Architectures](#supported-architectures)
- [Get Started](#get-started)
  - [Build](#build)
- [CLI](#cli)
  - [Usage](#usage)
  - [Configuration File](#configuration-file)

## Get Started

`go-mdatp` uses Go Modules introduced in Go 1.11 for dependency management.

### Build

Build the CLI for a target platform (Go cross-compiling feature), for example linux, by executing:

```bash
$ mkdir $HOME/src
$ cd $HOME/src
$ git clone https://github.com/devodev/go-mdatp.git
$ cd go-mdatp
$ env GOOS=linux go build -o go_mdatp_linux ./cmd/go-mdatp
..
```

If you are a Windows user, substitute the $HOME environment variable above with %USERPROFILE%.

## CLI

### Usage

> Auto-generated documentation for each command can be found [here](./docs/go-mdatp.md).

```bash
Interact with the Microsoft Defender ATP REST API.

Usage:
  go-mdatp [command]

Available Commands:
  alert       Alert resource type commands.
  gendoc      Generate markdown documentation for the go-mdatp CLI.
  help        Help about any command

Flags:
  -h, --help      help for go-mdatp
  -v, --version   version for go-mdatp

Use "go-mdatp [command] --help" for more information about a command.
```

### Configuration file

Commands that need to interact with the API require credentials to be provided using a YAML configuration file.</br>
The following locations are looked into if the --config flag is not provided:

```bash
$CWD/.go-office365.yaml
```

The following is the current schema used.
> Credentials can be found in `Azure Active Directory`, under: `Installed apps`.</br>

```yaml
---
Global:
  Identifier: some-id
Credentials:
  ClientID: 00000000-0000-0000-0000-000000000000
  ClientSecret: 00000000000000000000000000000000
  TenantID: 00000000-0000-0000-0000-000000000000
```
