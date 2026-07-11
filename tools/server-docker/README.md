## Docker

### Dockerfile

* [Server](server/Dockerfile).
* [Gen-Cert](genCert/Dockerfile).
* [Battle-Server-Manager](battle-server-manager/Dockerfile).

### Compose

* [compose-server](compose-server.yml): Server. Usable for all except AoM: RT and AoE IV: AE.
  * Environment variables:
    - **`GAME_ID`** (Mandatory): Either `age1`, `age2`, `age3`, `age4` or `athens`.
    - `GENCERT_ARGS` (Optional): Additional args to pass to `genCert` binary.
    - `SERVER_ARGS` (Optional): Additional args to pass to `server` binary.
* [compose-full](compose-full.yml): Server + Battle Server. Required for AoE: 2 on macOS Native, AoM: RT and AoE IV: AE.
  * Environment variables (+ `compose-server` ones):
    - **`BATTLE_SERVER_EXE`** (Mandatory): Path to `BattleServer.exe`.
    - `BS_MANAGER_ARGS` (Optional):  Additional args to pass to `battle-server-manager` binary.

*Note: There are issues on non-Linux OSes to receiver server announcements due to Docker running on a VM.*
