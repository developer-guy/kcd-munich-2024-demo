### go

You should run these two commands to verify the modules are working properly:

First, you need to create an apk for the go application with melange:

```shell
dagger -m "../../melange" call build --melange-file melange.yaml --workspace-dir=. --arch=aarch64 directory --path=. export --path=.
```

You will see that `packages` folder, `melange.rsa` and `melange.rsa.pub` are created:

```shell
$ ls -latr
...
melange.rsa
melange.rsa.pub
packages
$ tree -L 5 packages                                                                                                                         1.22.4
packages
└── aarch64
    ├── APKINDEX.json
    ├── APKINDEX.tar.gz
    └── hello-server-0.1.0-r0.apk
```

Now its time to build the OCI image with apko module:

```shell
dagger call -m "../../apko" build --apko-file apko.yaml --source=. --keyring-append=melange.rsa.pub --arch=arm64 --packages-append=packages --imag
e=ghcr.io/developer-guy/hello-server export --path="." --allowParentDirPath
```

You should see `apko.tar` is getting created, then load it into docker using `docker load`:

```shell
$ docker load < apko.tar
Loaded image: ghcr.io/developer-guy/hello-server:latest-arm64
```

Then run it:

```shell
$ docker container run ghcr.io/developer-guy/hello-server:latest-arm64
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

Voila!