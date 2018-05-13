# jsonnet-bundler

> NOTE: This project is *alpha* stage. Flags, configuration, behavior and design may change significantly in following releases.

The jsonnet-bundler is a package manager for [jsonnet](http://jsonnet.org/).

## Install

```
go get github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb
```

## Usage

All command line flags:

[embedmd]:# (_output/help.txt)
```txt
$ jb -h
usage: jb [<flags>] <command> [<args> ...]

A jsonnet package manager

Flags:
  -h, --help  Show context-sensitive help (also try --help-long and --help-man).
      --jsonnetpkg-home="vendor"  
              The directory used to cache packages in.

Commands:
  help [<command>...]
    Show help.

  init
    Initialize a new empty jsonnetfile

  install [<packages>...]
    Install all dependencies or install specific ones


```
