### System requirements

- OS requirements correspond to the server/launcher ones. Cross-compilation works on all systems out-the-box.
- Go 1.26:
    * Officially from https://go.dev/dl/ if not running Windows or is 10 and higher (or equivalent).
    * Unnoficially from [thongtech/go-legacy-win7](https://github.com/thongtech/go-legacy-win7) if running Windows
      7-8.X (or equivalent). Regardless if you install the official version, you need to install this one too for
      release.
- [Git](https://git-scm.com/downloads), with the latest supported for Windows 7-8 being v2.46.2.
- [Task](https://taskfile.dev/installation/).
- [GoReleaser](https://goreleaser.com/).

### Setup

Copy `.env.example` to `.env` and set:

* ```GPG_FINGERPRINT``` to the fingerprint of the key. Required only for `task release`.
* ```GOROOT_LEGACY``` to the legacy go installation path. Required for `task build` and
  `task release`.

### Debug

It is recommended to use an IDE such as [GoLand](https://www.jetbrains.com/go/) (free for academia)
or [Visual Studio Code](https://code.visualstudio.com/) (free) with
the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.go).

Depending on the module you want to debug, you will need to run the corresponding task **before**:

- server: ```task debug:prepare-server```
    - genCert: ```task debug:prepare-server-genCert```
- launcher: ```task debug:prepare-launcher```
    - config: ```task debug:build-config-admin-agent```
    - config-admin-agent: ```task debug:build-config-admin```
    - agent: ```task debug:build-config-all```
- battle-server-manager: ```task debug:prepare-battle-server-manager```

### Build

1. Make sure you have CGO disabled with ```go env -w CGO_ENABLED=0```
2. Run ```task build```.

### Release

1. Install [gpg2](https://docs.releng.linuxfoundation.org/en/latest/gpg.html) if needed.
2. Create a new sign-only GPG key pair (*RSA 4096-bit*) with a passphrase.
3. Make sure you have CGO disabled with ```go env -w CGO_ENABLED=0```
4. Finally run ```task release```

*Note: You will also need a local tag in semantic form like vX.Y.Z*
