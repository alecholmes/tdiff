# tdiff (Transitive Diff)

**This is an experimental, incomplete prototype.**

`tdiff` finds the transitive dependencies of a Go package and outputs relevant changes since an earlier Git commit.

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


# Notes

If a package is changed after the given SHA and before being added as a dependency, and does not change after this, irrelevant changes will be included.

# Roadmap

- Better README.
- Cleaner code.
- Better usage/help output.
- Nice error messages.
- Check that `git` and `go tool` are installed.
- Flag to include non-Go files in output.
    - TBD: This may just include changed files recursively reachable from the directory of the given package.
- Flag for verbose output (IN PROGRESS).
- Unit tests.
- Investigate: Correctly dealing with stdlib packages.
- Investigate: Is it possible to use a library instead of shelling out to `go list`?
- Investigate: Enumerate potential edge cases.
    - Could internal or test packages cause problems?
- Investigate: Use kingpin instead of flag library? Alternatively, it is nice to not have third party dependencies.


# License

This source code is open source through the GNU Affero General Public License v3.0. See the LICENSE file for full details or [this helpful summary](https://choosealicense.com/licenses/agpl-3.0/).
