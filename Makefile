VERSION ?= 0.1.0
DAGGER = dagger
HOME = $(shell echo $$HOME)

.PHONY: minikube
minikube:
	@# Start the minikube cluster
	minikube start --driver=docker --cpus=4 --memory=8192
	crane copy ghcr.io/developer-guy/hello-server:0.1.0 ttl.sh/dagger-demo/hello-server:0.1.0

.PHONY: apk
apk:
	@# Call Dagger melange module to build the APK
	cd `pwd`/hello-server; $(DAGGER) -m "github.com/tuananh/daggerverse/melange@5e6b42cb28fc18757def43ef0997adf752b329b1" call build \
		--melange-file melange.yaml \
		--workspace-dir=. \
		--arch=aarch64 directory --path=. export --path=.

.PHONY: image
image:
	@# Build the OCI image
	cd `pwd`/hello-server; $(DAGGER) -m "github.com/tuananh/daggerverse/apko@5e6b42cb28fc18757def43ef0997adf752b329b1" call build \
		--apko-file apko.yaml --source=. \
             	--keyring-append=melange.rsa.pub \
             	--arch=arm64 \
             	--packages-append=packages \
             	--tag $(VERSION) \
             	--image=ttl.sh/dagger-demo/hello-server export --path="." --allowParentDirPath

.PHONY: push
push:
	@# Push the image to the local registry
	cd `pwd`/hello-server; $(DAGGER) -m "../ci/" call container-push --source=. --container-as-tarball apko.tar --tag $(VERSION)

.PHONY: deploy
deploy:
	@# Deploy the image to the k3d cluster
	$(DAGGER) -m "ci/" call edit --source=manifests --tag="ttl.sh/dagger-demo/hello-server:$(VERSION)" export --path="manifests"
	$(DAGGER) -m "ci/" call commit-push --source=. --key="$(HOME)/.ssh/id_ed25519" export --path=.git/
