# tdiff (Transitive Diff)

**This is an experimental, incomplete prototype.**

`tdiff` finds the transitive dependencies of a Go package and outputs relevant changes since an earlier Git commit.

## Uses

`tdiff` must be called from a directory underneath `$GOPATH`.

"Related" means recursively reachable from the package given to `tdiff`.

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


# Notes

If a package is changed after the given SHA and before being added as a dependency, and does not change after this, irrelevant changes will be included.


# License

This source code is open source through the GNU Affero General Public License v3.0. See the LICENSE file for full details or [this helpful summary](https://choosealicense.com/licenses/agpl-3.0/).
