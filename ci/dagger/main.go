// A generated module for Ci functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
	"fmt"
	"time"
)

type Ci struct{}

// ContainerPush pushes the container tarball to the Docker daemon using the Docker CLI.
func (m *Ci) ContainerPush(
	// method call context
	ctx context.Context,
	// source directory with the container tarball
	source *Directory,
	// container tarball to push to the Docker daemon
	containerAsTarball *File,
	// Registry token to authenticate with the Docker registry
	// token *Secret,
	// +default="latest"
	tag string) (*Container, error) {
	cli := dag.Pipeline("container-push")

	//plaintext, _ := token.Plaintext(ctx)

	image := fmt.Sprintf("ttl.sh/dagger-demo/hello-server:%s", tag)

	return cli.Container().
		From("ghcr.io/regclient/regctl:latest").
		WithFile("/tmp/container.tar", containerAsTarball).
		//WithExec([]string{"registry", "login", "ghcr.io", "-u=developer-guy", fmt.Sprintf("-p=%s", plaintext)}).
		WithExec([]string{"image", "import", image, "/tmp/container.tar"}).
		WithExec([]string{"image", "copy", image, image}).
		Sync(ctx)
}

// Edit the kustomization file in the source directory with the given image tag
func (m *Ci) Edit(
	// method call context
	ctx context.Context,
	// source directory with the kustomization file
	source *Directory,
	// image tag to set in the kustomization file
	tag string) *Directory {
	return dag.Kustomize().
		Edit(source).
		Set().Image(tag).
		Directory()
}

// CommitPush local changes to the Git repository using the SSH Key.
func (m *Ci) CommitPush(
	// method call context
	ctx context.Context,

	// local dir with the Git repository and the changes
	source *Directory,

	// Git branch to push to.
	// +optional
	// +default="master"
	prBranch string,

	// SSH key with access credentials for the Git repository
	key *File,
) *Directory {

	cli := dag.Pipeline("commit-push")

	c := cli.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "update"}).
		WithExec([]string{"apk", "add", "git", "openssh"}).
		WithDirectory("/work", source).
		WithWorkdir("/work").
		WithFile("/tmp/.ssh/id", key, ContainerWithFileOpts{Permissions: 0400}).
		WithEnvVariable("GIT_SSH_COMMAND", "ssh -i /tmp/.ssh/id -o StrictHostKeyChecking=no").
		WithEnvVariable("CACHE_BUSTER", time.Now().String()).
		WithExec([]string{"git", "config", "--global", "user.name", "dagger-bot"}).
		WithExec([]string{"git", "config", "--global", "user.email", "developerguyn@gmail.com"})

	return c.WithExec([]string{"git", "status"}).
		WithExec([]string{"git", "add", "."}).
		WithExec([]string{"git", "commit", "-m", "autocommit"}).
		WithExec([]string{"git", "push"}).
		WithExec([]string{"git", "pull", "--rebase", "origin", "master"}).
		Directory(".git")
}
