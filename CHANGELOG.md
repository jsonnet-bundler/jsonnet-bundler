# Changelog

## 0.5.0 / 2022-06-09

#### Changes:

- **[FEATURE]** Add --quiet option to suppress git progress output (#124)
- **[FEATURE]** Support Bitbucket personal repositories (#156)
- **[FEATURE]** Add --legacy-name flag #158
- **[ENHANCEMENT]** Windows enhancements (#110)
- **[BUGFIX]** Allow dots in a repository path's "user" section (#106)
- **[BUGFIX]** On windows, use `\` instead of `/` (#115)
- **[BUGFIX]** Replace `/` in version by `-` (#146)
- **[BUGFIX]** Correct path resolution to nested local dependencies (#151)

## 0.4.0 / 2020-05-15

You can now `jb update` a single dependency.  
Run `jb update github.com/org/repo` (supports multiple at ones).

#### Changes:

- **[FEATURE]** Update single dependencies (#92)
- **[FEATURE]** Skip dependencies (#99)
- **[ENHANCEMENT]** Add support for subgroups (#91) (#93)
- **[BUGFIX]** Fix local package with relative path (#100) (#103) (#104)
- **[BUGFIX]** Fix unarchiver (#86)

## 0.3.1 / 2020-03-01

#### BREAKING:

The format of `jsonnetfile.json` has changed. While v0.3.0 can
handle the old v0.2 format, v0.2 can't and must not be used with a
`jsonnetfile.json` created by v0.3.0

#### Changes:

- **[FEATURE] Absolute imports (#63)**: Introduces a new style for importing the
  packages installed by `jb`. The `<name>/<file>` style used before caused
  issues, as it was neither unique nor clearly defined what to import.  
  To address this, `jb` will now create a directory structure that allows to use
  import paths similar to Go: `host.tld/user/repo/subdir/file.libsonnet`.  
  The old stlye is still supported, this change is backwards compatible.  
  `jb rewrite` can be used to automatically convert your imports.
- **[FEATURE] `jsonnetfile.json` versions (#85)**: Adds a `verison` key to
  `jsonnetfile.json`, so that `jb` can automatically recognize the schema
  version and handle it properly, instead of panicking.
- **[FEATURE] Generic `git` `https://` (#73)**: Previously the `host.tld/user/repo` slug
  style was only supported for GitHub. All hosts work now.
- **[BUGFIX]** `--jsonnetpkg-home` not working (#80)

## 0.2.0 / 2020-01-08

- **[FEATURE]** Rework installation process adding checksums (#44)
- **[FEATURE]** Add local dependencies as source dependency (#36)
- **[ENHANCEMENT]** Only write jsonnnet files if we made changes (#56)
- **[ENHANCEMENT]** Package install optimizations for git (#38)
- **[ENHANCEMENT]** Add integration tests (#35)
- **[ENHANCEMENT]** Suppress detached head advice (#34)
- **[BUGFIX]** Make sure to fetch git tags (#58)

## 0.1.0 / 2019-04-23

This is the first release for jsonnet-bundler.
