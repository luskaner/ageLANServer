## Local development

Copy `go.work.example` to `go.work`

### System requirements

- OS requirements correspond to the server/launcher ones. Cross-compilation works on all out-the-box.
- [Go 1.22](https://go.dev/dl/).
- [Git](https://git-scm.com/downloads).
- [Task](https://taskfile.dev/installation/).
- [GoReleaser](https://goreleaser.com/).

### Debug

It is recommended to use an IDE such as [GoLand](https://www.jetbrains.com/go/) (free for academia)
or [Visual Studio Code](https://code.visualstudio.com/) (free) with
the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.go).

Depending on the module you want to debug, you will need to run the corresponding task **before**:

- server: ```task debug-prepare-server```
    - genCert: ```task debug:prepare-server-genCert```
- launcher: ```task debug:prepare-launcher```
    - config: ```task build-config-admin-agent```
    - config-admin-agent: ```task build-config-admin```
    - agent: ```task build-config-all```

### Build

Run ```task build```.

### Release

1. Install [gpg2](https://docs.releng.linuxfoundation.org/en/latest/gpg.html) if needed.
2. Create a new sign-only GPG key pair (*RSA 4096-bit*) with a passphrase.
3. Copy .env.example to .env and set ```GPG_FINGERPRINT``` to the fingerprint of the key.
4. Finally run ```task release```
