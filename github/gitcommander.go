package github

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-github/github"
	"github.com/jenkins-x/godog-jx/utils"
)

var stdoutPrefix = utils.Color("\x1b[35m") + "        "
var stderrPrefix = utils.Color("\x1b[34m") + "        "

type GitCommander struct {
	Dir      string
	UseHttps bool
}

func CreateGitCommander() *GitCommander {
	dir := os.Getenv("WORK_DIR")
	if len(dir) == 0 {
		dir = "work"
	}
	return &GitCommander{
		Dir: dir,
	}
}

// Clone performs a clone in the directory for the repository
func (g *GitCommander) Clone(repo *github.Repository) (string, error) {
	cloneURL, err := g.GetCloneURL(repo)
	if err != nil {
		return "", err
	}
	return g.CloneFromURL(repo, cloneURL)
}

// CloneFromURL performs a clone in the directory for the repository using the given clone URL
func (g *GitCommander) CloneFromURL(repo *github.Repository, cloneURL string) (string, error) {
	owner := repo.Owner
	if owner == nil {
		return "", fmt.Errorf("No owner for the repository %v", repo)
	}
	ownerName := owner.Login
	if ownerName == nil {
		return "", fmt.Errorf("No owner login for the repository %v", repo)
	}
	repoName := repo.Name
	if repoName == nil {
		return "", fmt.Errorf("No repo name for the repository %v", repo)
	}
	outDir := filepath.Join(g.Dir, *ownerName, *repoName)

	runDir := filepath.Join(g.Dir, *ownerName)
	err := os.MkdirAll(runDir, 0770)

	if err != nil {
		return outDir, err
	}

	err = runCommand(runDir, "git", "clone", cloneURL)
	return outDir, err
}

// GetCloneURL returns the URL used to clone this repository
func (g *GitCommander) GetCloneURL(repo *github.Repository) (string, error) {
	return GetCloneURL(repo, g.UseHttps)
}

func (commander *GitCommander) GetLastCommitSha(dir string) (string, error) {
	text, err := commandAsString(dir, "git", "rev-parse", "HEAD")
	if err != nil {
		return text, err
	}
	return strings.TrimSpace(text), nil
}

// DeleteWorkDir removes all files inside the work dir so that the test can start clean
func (g *GitCommander) DeleteWorkDir() error {
	if _, err := os.Stat(g.Dir); err == nil {
		return RemoveDirContents(g.Dir)
	}
	return nil
}

func RemoveDirContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func (commander *GitCommander) ResetMasterFromUpstream(dir string, upstreamCloneURL string) error {
	err := runCommand(dir, "git", "remote", "add", "upstream", upstreamCloneURL)
	if err != nil {
		return err
	}
	err = runCommand(dir, "git", "fetch", "upstream")
	if err != nil {
		return err
	}
	err = runCommand(dir, "git", "checkout", "master")
	if err != nil {
		return err
	}
	err = runCommand(dir, "git", "reset", "--hard", "upstream/master")
	if err != nil {
		return err
	}
	err = runCommand(dir, "git", "push", "origin", "master", "--force")
	if err == nil {
		utils.LogInfof("reset the git repository at %s to the upstream master\n", dir)
	}
	return err
}

// runCommand runs the given command in the directory
func runCommand(dir string, prog string, args ...string) error {
	cmd := exec.Command(prog, args...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = utils.NewPrefixWriter(os.Stdout, stdoutPrefix)
	cmd.Stderr = utils.NewPrefixWriter(os.Stderr, stderrPrefix)
	if err := cmd.Run(); err != nil {
		text := prog + " " + strings.Join(args, " ")
		return fmt.Errorf("Failed to run command %s in dir %s due to error %v", text, dir, err)
	}
	return nil
}

// commandAsString runs the given command in the directory and returns the output of the command
func commandAsString(dir string, prog string, args ...string) (string, error) {
	var outb bytes.Buffer
	cmd := exec.Command(prog, args...)
	cmd.Dir = dir
	cmd.Stdout = &outb
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		text := prog + " " + strings.Join(args, " ")
		return "", fmt.Errorf("Failed to run command %s in dir %s due to error %v", text, dir, err)
	}
	return outb.String(), nil
}
