# jsonnet-bundler

> NOTE: This project is *alpha* stage. Flags, configuration, behavior and design may change significantly in following releases.

The jsonnet-bundler is a package manager for [Jsonnet](http://jsonnet.org/).


## Install

```
GO111MODULE="on" go get github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb
```
**NOTE**: please use a recent Go version to do this, ideally Go 1.13 or greater.

This will put `jb` in `$(go env GOPATH)/bin`. If you encounter the error
`jb: command not found` after installation then you may need to add that directory to your `$PATH` as shown [in their docs](https://golang.org/doc/code.html#GOPATH).

## Features

- Fetches transitive dependencies
- Can vendor subtrees, as opposed to whole repositories


## Current Limitations

- Always downloads entire dependent repositories, even when updating
- If two dependencies depend on the same package (diamond problem), they must require the same version


## Example Usage

Initialize your project:

```sh
mkdir myproject
cd myproject
jb init
```

The existence of the `jsonnetfile.json` file means your directory is now a
jsonnet-bundler package that can define dependencies.

To depend on another package (another Github repository):
*Note that your dependency need not be initialized with a `jsonnetfile.json`.
If it is not, it is assumed it has no transitive dependencies.*

```sh
jb install https://github.com/anguslees/kustomize-libsonnet
```

Now write `myconfig.jsonnet`, which can import a file from that package.
Remember to use `-J vendor` when running Jsonnet to include the vendor tree.

```jsonnet
local kustomize = import 'kustomize-libsonnet/kustomize.libsonnet';

local my_resource = {
  metadata: {
    name: 'my-resource',
  },
};

kustomize.namePrefix('staging-')(my_resource)
```

To depend on a package that is in a subtree of a Github repo (this package also
happens to bring in a transitive dependency):

```sh
jb install https://github.com/coreos/prometheus-operator/jsonnet/prometheus-operator
```

*Note that if you are copy pasting from the Github website's address bar,
remove the `tree/master` from the path.*

If pushed to Github, your project can now be referenced from other packages in
the same way, with its dependencies fetched automatically.


## All command line flags

[embedmd]:# (_output/help.txt)
```txt
$ jb -h
usage: jb [<flags>] <command> [<args> ...]

A jsonnet package manager

Flags:
  -h, --help     Show context-sensitive help (also try --help-long and
                 --help-man).
      --version  Show application version.
      --jsonnetpkg-home="vendor"  
                 The directory used to cache packages in.

Commands:
  help [<command>...]
    Show help.

  init
    Initialize a new empty jsonnetfile

  install [<uris>...]
    Install all dependencies or install specific ones

  update
    Update all dependencies.


```

## Design

This is an implemention of the design specified in this document: https://docs.google.com/document/d/1czRScSvvOiAJaIjwf3CogOULgQxhY9MkiBKOQI1yR14/edit#heading=h.upn4d5pcxy4c
