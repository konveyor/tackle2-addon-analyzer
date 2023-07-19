Running the addon locally is relatively straight forward:

Set environment variables:
- The `HUB_BASE_URL` must be set when not running the hub locally.
- The `TASK` **must** be set.
- The `TOKEN` must be set when auth is enabled in the hub.

The addon runs analyzer (CLI) commands:
- /usr/bin/konveyor-analyzer
- /usr/bin/konveyor-analyzer-dep

These commands have dependencies that can be complicated to install and
most developers won't want to install them in their development environment.
The recommended method is to run the commands in docker using the _quay.io/konveyor/analyzer-lsp_
image. A wrapper for these commands is provided in `hack/` which can be linked into `/usr/bin`.

The `reset` script should be invoked before each run:
- (re)creates /tmp/addon.
- copies settings.json to /tmp/addon/opt.

Example:
```
sudo ln -s $(pwd)/hack/windup-shim /usr/bin
sudo ln -s $(pwd)/hack/konveyor-analyzer /usr/bin
sudo ln -s $(pwd)/hack/konveyor-analyzer-dep /usr/bin
```

Example: with Hub running in minikube.
```
$ minikube ip
192.168.49.2
$ export HUB_BASE_URL=http://192.168.49.2/hub
$ export TASK=1
$ reset
$ make run
```
