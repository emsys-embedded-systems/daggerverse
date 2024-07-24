package main

import (
	"context"
	"dagger/ci/internal/dagger"
	"fmt"
	"strings"
	"time"
)

type Ci struct{}

func aptGetUpdate(nestedCommands ...string) dagger.WithContainerFunc {
	return func(ctr *dagger.Container) *dagger.Container {
		lists := "/var/lib/apt/lists"
		archives := "/var/cache/apt/archives"
		return ctr.WithMountedTemp(lists).WithMountedTemp(archives).
			// WithEnvVariable(pythonDontWriteBytecode, "1").
			WithExec([]string{"sh", "-c", strings.Join(append([]string{
				"set -eux",
				"apt-get update"},
				nestedCommands...), "; ")}).
			// WithoutEnvVariable(pythonDontWriteBytecode).
			WithoutMount(archives).WithoutMount(lists)
	}
}

// A periodically changing 'echo ...' may trigger a rebuild/update.
func refreshMonthly() string {
	year, month, _ := time.Now().Date()
	return fmt.Sprint("date; echo ", month, year)
}

func aptGetUpgrade() string {
	return "apt-get upgrade -y"
}

func aptGetInstall(packages ...string) string {
	return "DEBIAN_FRONTEND=noninteractive" +
		" apt-get install -y --no-install-recommends " + strings.Join(packages, " ")
}

func (r *Ci) WithPackages(ctr *dagger.Container, packages []string) *dagger.Container {
	return ctr.With(aptGetUpdate(refreshMonthly(),
		aptGetUpgrade(), aptGetInstall(packages...)))
}

func (*Ci) DirOrGit(
	ctx context.Context,
	name, path, commit string,
	// +optional
	token *dagger.Secret,
	// +optional
	dir *dagger.Directory,
	// +optional
	keepGitDir bool,
) (*dagger.Directory, error) {
	if dir != nil {
		return dir, nil
	}
	if token == nil {
		return nil, fmt.Errorf(`must set %q or "git-token"`, name)
	}
	userAndToken, err := token.Plaintext(ctx)
	if err != nil {
		return nil, err
	}
	if !strings.Contains(path, "/") {
		path = "emsys/" + path
	}
	if !strings.HasPrefix(userAndToken, "gitea-bot:") {
		userAndToken = "token:" + userAndToken
	}
	url := "https://" + userAndToken + "@code.emsys.de/" + path + ".git"
	opts := dagger.GitOpts{KeepGitDir: keepGitDir}
	if commit == "" {
		return dag.Git(url, opts).Head().Tree(), nil
	} else {
		return dag.Git(url, opts).Ref(commit).Tree(), nil
	}
}

// emsysRegistryUrl converts "cr.emsys.de/repo/name@sha.." to a "https://cr.emsys.de/.." link.
func emsysRegistryUrl(imageRef string) (url string, found bool) {
	repoSlashNameAtSha, _ := strings.CutPrefix(imageRef, "cr.emsys.de/")
	repo, nameAtSha, _ := strings.Cut(repoSlashNameAtSha, "/")
	nameAndTag, sha, found := strings.Cut(nameAtSha, "@")
	name, _, _ := strings.Cut(nameAndTag, ":")
	if found {
		switch repo { // The "Add tag" button only works with the numeric repo id.
		case "bamboo":
			repo = "12"
		}
		url = "https://cr.emsys.de/harbor/projects/" + repo + "/repositories/" +
			name + "/artifacts-tab/artifacts/" + sha
	}
	return
}

func (*Ci) AppendEmsysRegistryUrlToRef(ref string, quiet bool) string {
	url, found := emsysRegistryUrl(ref)
	if !found || quiet {
		return ref
	}
	return ref + "\n\t" + url
}
