# tdiff (Transitive Diff)

**This is still under development.**

For applications that live in shared Git repositories, it can be difficult to determine which
changes between two Git commits are relevant to the application and which are not.

`tdiff` aims to isolate only the relevant changes for an application and report those changes in various formats.

Relative changes are determined by building a graph of package dependencies and ignoring packages not
reachable from a given package.

To illustrate, consider the following set of 6 Go packages:

```
myorg/app_one
    imports myorg/lib/collections
    imports thirdparty/http_util

myorg/app_two
    imports myorg/format/text

myorg/lib/collections
    imports myorg/format/text
    imports net/http

myorg/format/text

thirdparty/http_util
    imports net/http

net/http
```

There relevant set of packages for each:

* `myorg/app_one`: [`myorg/app_one`, `myorg/lib/collections`, `thirdparty/http_util`, `myorg/format/text`, `net/http`]
* `myorg/app_two`: [`myorg/app_two`, `myorg/format/text`]
* `myorg/lib/collections`: [`myorg/lib/collections`, `myorg/format/text`, `net/http`]
* `myorg/format/text`: [`myorg/format/text`]
* `thirdparty/http_util`: [`thirdparty/http_util`, `net/http`]
* `net/http`: [`net/http`]

Given that, when determining changes for `myorg/app_two`, only changes in the `myorg/app_two` and `myorg/format/text`
packages are considered.

## Installing

```
go get -u github.com/alecholmes/tdiff
```

## Uses

"Related" means the package passed to `tdiff` and all recursively reachable packages.

### List related packages changed since the given SHA

```
tdiff -package your/app/list_utils -sha OLDER_GIT_SHA -packages
```

### List related files changed since the given SHA

```
tdiff -package your/app/list_utils -sha OLDER_GIT_SHA -files
```

### List Git commits since the given SHA that contain changes to related files

```
tdiff -package your/app/list_utils -sha OLDER_GIT_SHA -commits
```

### Get JSON output for all of the above, and more

```
tdiff -package your/app/list_utils -sha OLDER_GIT_SHA -json
```

The JSON output includes separate sections for packages, files, and commits.

Additionally, each changed package also includes a path indicating how it is reachable from the given root path.
A path of `["A", "B", "C"]` indicates "A imports B, and B imports C".

### HTML Summary

```
# This will print the name of the summary HTML temp file:
tdiff -package your/app/list_utils -sha OLDER_GIT_SHA -html

# Open the summary HTML file:
open $(tdiff -package your/app/list_utils -sha OLDER_GIT_SHA -html)
```

### Including Non-Go Files

`tdiff` can consider and include all files, not just Go source files. This is useful for picking up changes
to config files, for example.

To include all files, add the `-artifacts` flag, e.g. `tdiff -package your/app/list_utils -sha OLDER_GIT_SHA -packages -artifacts`.


# Notes

If a package is changed after the given SHA and before being added as a dependency, and does not change after this, irrelevant changes will be included.


# Roadmap

- Better usage/help output.
- Nice error messages.
- Check that `git` and `go tool` are installed.
- Flag for verbose output (IN PROGRESS).
- Better unit test coverage, especially for the `app` package.
- Investigate: Enumerate potential edge cases.
    - Could internal or test packages cause problems?
- Investigate: Use kingpin instead of flag library?


# License

This source code is open source through the GNU Lesser General Public License v3.0.
See the [LICENSE](LICENSE) file for full details or [this helpful summary](https://choosealicense.com/licenses/lgpl-3.0/).
