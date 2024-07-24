# CI

Dagger module with reusable dagger code for Emsys CI plans.

The following module functions will be generated:

```go
// Installs Ubuntu/Debian packages.
func (r *Ci) WithPackages(ctr *Container, packages []string) *Container

type CiDirOrGitOpts struct {
	Token *Secret
	Dir *Directory
	KeepGitDir bool
}
// Selects a directory with sources.
func (r *Ci) DirOrGit(name string, path string, commit string, opts ...CiDirOrGitOpts) *Directory

// Append a line with a https://cr.emsys.de/... link for a published image-ref.
func (r *Ci) AppendEmsysRegistryURLToRef(ctx context.Context, ref string, quiet bool) (string, error)
```
