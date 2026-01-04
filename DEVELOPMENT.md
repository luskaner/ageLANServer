## Local development

Copy `go.work.example` to `go.work`

### System requirements

- OS requirements correspond to the server/launcher ones. Cross-compilation works on all systems out-the-box.
- [Go 1.25](https://go.dev/dl/) or higher, except for Windows 7-8 (and equivalent) which need an unofficial fork
  like [thongtech/go-legacy-win7](https://github.com/thongtech/go-legacy-win7) (recommended)
  or [XTLS/go-win7](https://github.com/XTLS/go-win7).
- [Git](https://git-scm.com/downloads), with the latest supported for Windows 7/8 being v2.46.2.
- [Task](https://taskfile.dev/installation/).
- [GoReleaser](https://goreleaser.com/).

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

Run ```task build```.

### Release

1. Install [gpg2](https://docs.releng.linuxfoundation.org/en/latest/gpg.html) if needed.
2. Create a new sign-only GPG key pair (*RSA 4096-bit*) with a passphrase.
3. Copy .env.example to .env and set ```GPG_FINGERPRINT``` to the fingerprint of the key.
4. Finally run ```task release```

*Note: You will also need a local tag in semantic form like vX.Y.Z*
