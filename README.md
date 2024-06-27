# kcd-munich-2024-demo

No More YAML Soup: Taking Control with Dagger's Pipeline-as-Code Philosophy

> https://www.kcdmunich.de/schedule/

## Goal

The goal of this repository is to demonstrate how to use Dagger in a Kubernetes environment to manage the deployment of applications.

For this demonstration, we are going to use a simple Go application that exposes a REST API. The application is going to be deployed in a Kubernetes cluster using Dagger with a pipeline-as-code approach. 

For the sake of simplicity, we are going to use a local Kubernetes cluster managed by `k3d` and `Flux` to manage the GitOps workflow.

Talk is cheap - show me the code!

### Prerequisites

* [flux CLI](https://fluxcd.io/flux/installation/#install-the-flux-cli)
* Kubernetes (_we are using [minikube](https://minikube.sigs.k8s.io/docs/start/) in this demonstration but any Kubernetes distribution would be fine_)
* [dagger](https://docs.dagger.io/quickstart/cli/#install-the-dagger-cli)
* [crane](https://github.com/google/go-containerregistry/blob/main/cmd/crane/README.md#installation)

### Setup

First things first, we are going to create a new Kubernetes cluster using `minikube`:

```bash
make minikube
```

Verify the cluster is up and running:

```bash
kubectl cluster-info
```

If everything is working as expected, we can proceed to the next step.

After that, we are going to install `flux` in the cluster:

```bash
flux bootstrap github \
  --owner=developer-guy \
  --repository=kcd-munich-2024-demo \
  --branch=master \
  --path=./clusters/dagger-in-action \
  --personal
```

> **Note:** The `--personal` flag is used to create a personal access token for the GitHub repository, this command is going to ask you to provide your PAT (Personal Access Token), to fix this you can use `gh auth token |export GITHUB_TOKEN=$(cat /dev/stdin)` trick to set the token, btw, gh is the GitHub CLI where you can install it from [here](https://cli.github.com).

> **BONUS:** There is an alternative way to bootstrap Flux with Git-less approach using OCI Registry,  [@stefanprodan](https://x.com/stefanprodan) who is the maintainer of the Flux started an RFC for this feature, you can find more details [here](https://github.com/fluxcd/flux2/pull/4749)

After the `flux` is installed, you will see a bunch of files in the `clusters/dagger-in-action` directory created, these files are the Kubernetes manifests that are going to be deployed in the cluster by Flux. So, we are going to deploy the Dagger in the cluster by creating the manifests in that folder., now its time to deploy the Dagger in the cluster, as we follow the GitOps approach, we are going to use the `flux` to deploy the Dagger in the cluster.

Dagger provides a Helm chart to deploy the Dagger in the cluster, you can find more information how to set up a Dagger in Kubernetes environment [here](https://docs.dagger.io/integrations/kubernetes/).

Now, lets create a Helm repository for the Dagger Helm chart:

```bash
flux create source helm dagger-repo \
  --url=oci://registry.dagger.io \
  --export > ./clusters/dagger-in-action/dagger-oci-source.yaml
```

After that, we are going to create a HelmRelease to deploy the Dagger in the cluster:

```bash
 flux create helmrelease dagger \
  --source=HelmRepository/dagger-repo.flux-system \
  --chart=dagger-helm \
  --target-namespace=dagger \
  --create-target-namespace \
  --export > ./clusters/dagger-in-action/dagger-helm-release.yaml
```

Then run reconcile command to apply the changes immediately without waiting the default interval:

```bash
$ flux reconcile source git flux-system
► annotating GitRepository flux-system in flux-system namespace
✔ GitRepository annotated
◎ waiting for GitRepository reconciliation
...
```

After a few seconds, you should see the Dagger is deployed in the cluster:

```bash
$ kubectl get pods -n dagger
NAME                                     READY   STATUS    RESTARTS   AGE
dagger-dagger-dagger-helm-engine-5bmrz   1/1     Running   0          100s
```

Now, we have the Dagger deployed in the cluster, we can proceed to the next step.

Let's connect to the Dagger engine locally:

```bash
DAGGER_ENGINE_POD_NAME="$(kubectl get pod \
    --selector=name=dagger-dagger-dagger-helm-engine --namespace=dagger \
    --output=jsonpath='{.items[0].metadata.name}')"
export DAGGER_ENGINE_POD_NAME

_EXPERIMENTAL_DAGGER_RUNNER_HOST="kube-pod://$DAGGER_ENGINE_POD_NAME?namespace=dagger"
export _EXPERIMENTAL_DAGGER_RUNNER_HOST
```

Let's ensure that the Dagger engine connection is working:

```bash
$ echo $_EXPERIMENTAL_DAGGER_RUNNER_HOST
kube-pod://dagger-dagger-dagger-helm-engine-5bmrz?namespace=dagger
```

Then, call the simple hello module by [@solomonstre](https://x.com/solomonstre):

```bash
dagger -m github.com/shykes/daggerverse/hello@v0.2.0 call hello
```

If you see the output of the above command as `hello, world!`, then the connection is working as expected and we can proceed to the next step.

Let's deploy the application:

> **Note:** We already built the `0.1.0` version of the application for the demonstration purposes.

```shell
flux create kustomization hello-server \
  --source=GitRepository/flux-system \
  --path="./kustomize" \
  --prune=true \
  --interval=10m \
  --timeout=1m \
  --target-namespace=default \
  --namespace=flux-system --export > ./clusters/dagger-in-action/hello-server-kustomization.yaml
```

Then it will be creating a deployment for the application:

```shell
$ kubectl port-forward pod/$(kubectl get pods -l app=hello-server -o jsonpath='{.items[0].metadata.name}') 8080
Forwarding from 127.0.0.1:8080 -> 8080
Forwarding from [::1]:8080 -> 8080
```

Then, you can test the application:

```shell
$ http :8080
HTTP/1.1 200 OK
Content-Length: 18
Content-Type: text/plain; charset=utf-8
Date: Sat, 29 Jun 2024 08:29:18 GMT

Hello World 0.1.0!
```

Noice!

### Deploy the Go application

Now, we are going to deploy the Go application in the cluster using Dagger modules `apko` and `melange` where you can find the details of these modules in [@tuananh_org](https://x.com/tuananh_org)'s [daggerverse](https://github.com/tuananh/daggerverse).

> [Daggerverse](https://daggerverse.dev) is a place where you can discover and share modules full of Dagger functions encapsulating the community's devops knowledge.

But let me show you the code of the module to basically show you what is happening under the hood:

<details>

<summary>Melange module</summary>

```go
type Melange struct{}

func (m *Melange) Build(
	ctx context.Context,
	melangeFile *File,
	workspaceDir *Directory,
// +default="amd64"
	arch string,
// +default="latest"
	imageTag string,
) *Directory {
	// generate public/private key pair
	cli := dag.Pipeline("melange-build")

	ctr := cli.Container().From(fmt.Sprintf("cgr.dev/chainguard/melange:%s", imageTag)).
		WithWorkdir("/workspace").
		WithExec([]string{
			"keygen"})

	f, _ := melangeFile.Name(ctx)

	c := cli.Container().
		From(fmt.Sprintf("cgr.dev/chainguard/melange:%s", imageTag)).
		WithMountedDirectory("/workspace", workspaceDir).
		WithDirectory("/workspace", ctr.Directory("/workspace")).
		WithWorkdir("/workspace").
		WithExec([]string{
			"build", fmt.Sprintf("%s", f), "--arch", arch, "--signing-key=melange.rsa"},
			ContainerWithExecOpts{
				ExperimentalPrivilegedNesting: true,
				InsecureRootCapabilities:      true,
			})

	pk := c.File(filepath.Join("/workspace", "melange.rsa.pub"))

	return c.Directory("/workspace").WithFile(".", pk)
}
```

</details>

> If you would like to learn how to develop your own Dagger modules, please check the [Dagger documentation](https://docs.dagger.io/quickstart/daggerize).

For the the who don't know what is `apko` and `melange`, these are the newest tools by Chainguard that are used to create an OCI images. `apko` is a tool to create an OCI image from a apks you created with `melange` from scracth.

I highly recommend to check the `apko` and `melange` tools, they are really cool tools to create an OCI images, [here](https://edu.chainguard.dev/open-source/build-tools/), also, [@adrianmouat](https://x.com/adrianmouat) who is a DevRel from Chainguard gave a presentation about `Building Container Images the Modern Way` where he talked about the tools that you can use to create OCI images today, you can find the video [here](https://www.youtube.com/watch?v=nZLz0o4duRs).


First, we need to create an apk for the go application with melange:

```shell
dagger -m "github.com/tuananh/daggerverse/melange@5e6b42cb28fc18757def43ef0997adf752b329b1" call build --melange-file melange.yaml --workspace-dir=. --arch=aarch64 directory --path=. export --path=.
```

You will see that `packages` folder, `melange.rsa` and `melange.rsa.pub` are created:

```shell
$ ls -latr
...
melange.rsa
melange.rsa.pub
packages
$ tree -L5 packages
packages
└── aarch64
    ├── APKINDEX.json
    ├── APKINDEX.tar.gz
    └── hello-server-0.1.0-r0.apk
```

Now its time to build the OCI image with apko module:

```shell
dagger call -m "github.com/tuananh/daggerverse/apko@5e6b42cb28fc18757def43ef0997adf752b329b1" build --apko-file apko.yaml --source=. --keyring-append=melange.rsa.pub --arch=arm64 --packages-append=packages --image=ghcr.io/developer-guy/hello-server --tag v2 export --path="." --allowParentDirPath
```

You should see `apko.tar` is getting created, then let's push it to the registry:

```shell
$ dagger -m "../ci/" call container-push --source=. --container-as-tarball apko.tar
```
> **Note:** The `REGISTRY_PASSWORD` is the password of the registry, you can set it as an environment variable ie. `pbpaste|export REGISTRY_PASSWORD=$(cat /dev/stdin)`.

Once we pushed the image to the registry, let's test it before we can deploy it in the cluster:

```shell
$ crane ls ghcr.io/developer-guy/hello-server
0.1.0

$ docker container run --rm -p 8080:8080 ghcr.io/developer-guy/hello-server:0.1.0
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:   export GIN_MODE=release
 - using code:  gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /                         --> main.main.func1 (3 handlers)
[GIN-debug] [WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.
Please check https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies for details.
[GIN-debug] Environment variable PORT is undefined. Using port :8080 by default
[GIN-debug] Listening and serving HTTP on :8080
```

In a second terminal window, let's test the application:

```shell
$ http :8080
HTTP/1.1 200 OK
Content-Length: 18
Content-Type: text/plain; charset=utf-8
Date: Sat, 29 Jun 2024 08:14:45 GMT

Hello World 0.1.0!
```

Noice! The application is working as expected, now we can deploy it in the cluster.

To deploy the application in the cluster, we are going to use our own Dagger module, let's see what we have as functions in our module:

```shell
$ dagger -m "ci/" functions
Name             Description
commit-push      CommitPush local changes to the Git repository using the SSH Key.
container-push   ContainerPush pushes the container tarball to the Docker daemon using the Docker CLI.
edit             Edit the kustomization file in the source directory with the given image tag
```

We are going to use the `edit` function to edit the `kustomization.yaml` file with the image tag:

```shell
dagger-m "ci/" call edit --source=manifests --tag="ghcr.io/developer-guy/hello-server:$(VERSION)" export --path="manifests"
```

This command will edit the `kustomization.yaml` file with the image tag, then we can deploy the application in the cluster by commiting and pushing these changes:

```shell
dagger -m "ci/" call commit-push --source=. --key=/Users/batuhanapaydin/.ssh/id_ed25519 export --path=.git/
```

Then, you should see the changes are getting applied in the cluster right after you trigger re-conciliation:

```shell
$ flux reconcile source git flux-system
```

After a few seconds, you should see the new version of the application is deployed in the cluster, do the same test as we did before:

```shell
$ kube port-forward pod/$(kubectl get pods -l app=hello-server -o jsonpath='{.items[0].metadata.name}') 8080
Forwarding from 127.0.0.1:8080 -> 8080
Forwarding from [::1]:8080 -> 8080
```

Then, you can test the application:

```shell
$ http :8080
HTTP/1.1 200 OK
Content-Length: 18
Content-Type: text/plain; charset=utf-8
Date: Sat, 29 Jun 2024 08:29:18 GMT

Hello World 0.2.0!
```

Yay!