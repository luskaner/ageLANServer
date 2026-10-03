# Especificación: diálogos gráficos (zenity) en el módulo `launcher`

Estado: propuesta cerrada para implementación.
Alcance: **exclusivamente** el módulo `launcher` (`github.com/luskaner/ageLANServer/launcher`).
Ningún otro módulo del workspace cambia, salvo por el `go.work.sum` que se regenera.

---

## 1. Objetivo

Añadir diálogos gráficos nativos al `launcher` usando la librería
[`github.com/ncruces/zenity`](https://github.com/ncruces/zenity), **conservando el
comportamiento actual de consola como fallback**. El modo por defecto es
`auto`, por lo que el cambio es transparente para el usuario que ya usa el
launcher hoy.

## 2. Justificación de la elección de integración

Se usa la **librería Go** `github.com/ncruces/zenity`, no el binario `zenity`
del CLI ni un binario nuevo empaquetado:

| Opción | Valorada | Decisión |
|---|---|---|
| Librería Go `github.com/ncruces/zenity` | `CGO_ENABLED=0`; en Windows usa Win32 nativo sin dependencias ni manifiesto; en macOS usa `osascript` preinstalado; en Linux delega en `qarma`/`zenity`/`matedialog` del `PATH` y expone `IsAvailable()`. Ya es un módulo Go → cero binarios nuevos en el `.goreleaser`, cero trabajo en `tools/scripts/internal/goreleaser/config.go`. | **Elegida** |
| Empaquetar `zenity.exe` / binario universal | Añade ~1.5 MB por plataforma, una matriz de build nueva y un artefacto más que mantener en `codeql.yml` y `goreleaser`. | Descartada |
| Shell out al binario `zenity` del `PATH` | En Windows/macOS el usuario tendría que instalarlo aparte (`brew`, `scoop`, `go install`); contradice que el launcher sea autosuficiente. | Descartada |

Requisitos de toolchain verificados: `zenity` declara `go 1.25.0`; el launcher
declara `go 1.27.0` (`launcher/go.mod:3`) y el workspace `go 1.27.1`
(`go.work:1`). Sin conflicto. `CGO_ENABLED=0` (`DEVELOPMENT.md:57`) es
compatible: `zenity` no usa cgo.

Dependencias que entra el `go.mod` del launcher al hacer `go get`:
`github.com/dchest/jsmin`, `github.com/ncruces/go-strftime`,
`golang.org/x/image`. Los *tool* deps de `zenity` (`josephspurrier/goversioninfo`,
`randall77/makefat`, `akavel/rsrc`) son exclusivos de `cmd/zenity` y su script de
build, así que el launcher no los usa: quedan hashes sobrantes en `go.sum` y ahí
se quedan, porque `go mod tidy` no puede ejecutarse en este workspace (ver el
aviso en §7.7). **No editar `go.mod`/`go.sum` a mano.**

## 3. Alcance funcional exacto

El `launcher` tiene exactamente **dos puntos interactivos** hoy, ambos sobre
`stdin`. Son el único alcance funcional de esta especificación.

| # | Punto interactivo | Ubicación actual | Diálogo zenity |
|---|---|---|---|
| I1 | Selección del servidor descubierto en LAN | `launcher/internal/cmdUtils/server.go:115` (`selectServerIndex`) | `zenity.List` |
| I2 | Confirmación antes de arrancar servidor propio | `launcher/internal/cmd/root.go:702-705` | `zenity.Question` |

Todo lo demás del launcher es `logger.Println`/`Printf` no interactivo y **no
cambia**.

### 3.1 Lo que queda fuera de alcance (decisiones explícitas)

- **No** se convierte ningún `logger.Println` en diálogo (incluido "Setting
  up...", "Starting 'server'...", errores fatales). Los errores se siguen
  reportando por consola + fichero de log. Motivo: cambiaría ~40 rutas de
  error por todo el módulo y el `exitCode` es el contrato estable con
  `start.sh`/`start.bat`.
- **No** se añade un flag `--dialog` por juego ni un modo "always GUI" que
  bloquee el flujo cuando no hay servidor que descubrir.
- **No** se pasa `zenity.Context(...)` a los diálogos. `Ctrl+C` mantiene el
  comportamiento actual: el handler de `SIGINT`/`SIGTERM` de
  `cmd/root.go:501-513` fuerza el `teardown` y hace `os.Exit`.
- **No** se toca `common/resources/start.sh` / `start.bat`. Observación
  aparte: ambos hacen `read -r dummy` / `pause` cuando el launcher devuelve un
  código distinto de 0, lo que **cuelga indefinidamente** si no hay consola
  interactiva (p. ej. lanzado desde Steam en una sesión de servicio). Es un
  problema preexistente de `common`, fuera de este módulo; si se quiere
  resolver, la corrección natural es un diálogo `--error` de zenity en el
  launcher o una guarda de TTY en `common/resources`.

## 4. Diseño

### 4.1 Nuevo paquete `launcher/internal/dialog`

Responsabilidad única: abstraer los dos prompts para que puedan servirse por
consola o por zenity. No conoce el juego, el servidor ni el flujo de `runRoot`.

```
launcher/internal/dialog/
├── dialog.go        # interfaz, singleton de sesión, resolución de modo
├── console.go       # implementación de consola (fallback, = comportamiento actual)
├── zenity.go        # implementación zenity
├── dialog_test.go
├── console_test.go
└── zenity_test.go
```

**Restricción de dependencias (obligatoria).** `dialog` **no** puede importar
`launcher/internal/cmdUtils` ni `launcher/internal/cmdUtils/logger` (el segundo
ya no podría importar `dialog`, y `cmdUtils` necesita importar `dialog`). Por
eso la salida de la consola no usa `logger.Println` directamente sino un sink
inyectable.

#### `dialog.go`

```go
// Package dialog abstracts the launcher's interactive prompts so they can be
// answered either in a graphical window or in the console.
package dialog

import (
	"io"
	"sync"
	"sync/atomic"
)

// Mode values accepted by Config.Dialog and --dialog.
const (
	ModeAuto  = "auto"
	ModeTrue  = "true"
	ModeFalse = "false"
)

// Dialog is the set of interactive prompts the launcher needs.
type Dialog interface {
	// Name returns the backend identifier recorded in the log file.
	Name() string

	// SelectServer asks which of the discovered 'server's to use. servers
	// holds the candidates already sorted by latency, each one carrying both
	// the full Description and the compact Label. It returns the 0-based index
	// into servers.
	// ok is false when the user declined to pick one, in which case the
	// caller must fall back to its own default (start its own 'server').
	// stdin is only read by the console implementation.
	SelectServer(servers []ServerCandidate, stdin io.Reader) (index int, ok bool)

	// ListCandidates shows the candidate list without asking anything. The
	// console backend prints it, exactly as SelectServer does before its
	// prompt, so paths that answer on their own still leave the console user
	// with a trace of what was considered. Backends that render the list in
	// their own window leave it as a no-op.
	ListCandidates(servers []ServerCandidate)

	// ConfirmStartServer asks whether to go ahead and start the 'server'.
	// It returns false only when the user actively declined.
	ConfirmStartServer(text string, stdin io.Reader) bool
}

// Por qué existe ListCandidates: el camino de auto-select
// (Server.SingleAutoSelect con un único servidor descubierto) responde sin
// preguntar nada, así que SelectServer no llega a llamarse. Antes de este
// cambio esa ruta SÍ imprimía el bloque numerado "Found the following
// 'server's:" + "1. <desc>" y sólo después "Auto-selecting the only found
// 'server'.", porque el bucle de selectServerIndex pintaba la lista antes de
// comprobar el auto-select. Sin este método, mover el atajo por encima del
// prompt habría borrado esas dos líneas de la salida de consola (y del fichero
// de log) en un camino que el usuario ve todos los días.
//
// ListCandidates deja esa decisión donde corresponde: el backend. El de consola
// imprime exactamente el mismo bloque que ya imprime SelectServer antes del
// prompt —extraído a un printCandidates compartido, sin duplicar el formato—; el
// gráfico es un no-op documentado, porque su ventana pinta la lista por su
// cuenta. Así se evita meter singleAutoSelect dentro de la interfaz: la política
// de "cuándo no se pregunta" sigue en un único sitio, selectDiscoveredServer
// (§4.2), y ListCandidates sólo resuelve "cómo se muestra una lista cuando no
// hay pregunta".

// Output are the sinks the console prompts write to. They mirror
// cmdUtils/logger so prompt output keeps reaching both the log file and stdout.
type Output struct {
	Println func(...any)
	Printf  func(string, ...any)
}

// Resolution is the outcome of resolving the configured mode to a backend.
type Resolution struct {
	Dialog Dialog
	// Name is "zenity" or "console".
	Name string
	// Reason is non-empty when the configured mode could not be honoured and
	// the console fallback was used instead. It is meant to be logged.
	Reason string
}

var (
	active atomic.Pointer[Dialog]

	outputMu sync.RWMutex
	output   = Output{Println: func(a ...any) { fmt.Println(a...) },
		Printf: func(f string, a ...any) { fmt.Printf(f, a...) }}
)

// New resolves mode ("auto", "true" or "false") to a dialog backend. It never
// returns a nil Dialog.
func New(mode string) Resolution { /* ... */ }

// Set installs the dialog for the rest of the session. Set is safe to call
// from tests; RunRoot resets it after teardown.
func Set(d Dialog) { active.Store(&d) }

// Active returns the installed dialog, or a console dialog when Set has not
// been called, so the prompts keep working on early call paths and in tests.
func Active() Dialog { /* ... */ }

// Reset removes the installed dialog.
func Reset() { active.Store(nil) }

// SetOutput installs the sinks used by the console dialog.
func SetOutput(o Output) { /* ... */ }
```

`New(mode)` cumple esta tabla, evaluando `available := zenity.IsAvailable()`
**una sola vez** mediante `sync.OnceValue`:

| `mode` | `IsAvailable()` | `Name` | `Reason` |
|---|---|---|---|
| `false` | irrelevante | `console` | `""` |
| `true` | `true` | `zenity` | `""` |
| `true` | `false` | `console` | `"Graphical dialogs are not available in this system, using the console instead."` |
| `auto` | `true` | `zenity` | `""` |
| `auto` | `false` | `console` | `""` |

Regla dura: **`mode = true` nunca aborta el launcher**. Es un requisito de
compatibilidad, no un detalle: la librería sólo devuelve `false` en Linux/BSD
cuando no hay `qarma`/`zenity`/`matedialog` en el `PATH`, y el usuario que
escribió `"true"` explícitamente no debe verse bloqueado por ello. `auto` cae
en silencio porque es el modo por defecto.

`Name` y `Reason` se devuelven en lugar de imprimirse dentro de `dialog` para
que el mensaje salga por `cmdUtils/logger` (fichero de log + stdout) sin que
`dialog` tenga que importar `cmdUtils/logger`.

#### `console.go` — el fallback

Contiene, sin cambios de comportamiento, la lógica que hoy vive en
`cmdUtils/server.go:115-140`, más la lectura de `Enter` que hoy vive en
`cmd/root.go:704`:

```go
type consoleDialog struct{}

func (consoleDialog) Name() string { return "console" }

func (consoleDialog) SelectServer(servers []ServerCandidate, reader io.Reader) (int, bool) {
	out := sinks()
	for {
		printCandidates(servers)
		out.Printf("Enter the number of the 'server' (1-%d): ", len(servers))
		var option int
		if _, err := fmt.Fscan(reader, &option); err != nil {
			// Stdin exhausted or broken: we can never get a valid answer,
			// so retrying would spin forever printing the list.
			out.Println("Could not read selection from input.")
			return 0, false
		}
		if option < 1 || option > len(servers) {
			out.Println("Invalid option. Please enter a number from the list.")
			continue
		}
		return option - 1, true
	}
}

// ListCandidates imprime la lista exactamente como SelectServer la imprime
// antes de su prompt.
func (consoleDialog) ListCandidates(servers []ServerCandidate) { printCandidates(servers) }

func (consoleDialog) ConfirmStartServer(text string, reader io.Reader) bool {
	sinks().Println(text + " Press enter to continue...")
	_, _ = bufio.NewReader(reader).ReadBytes('\n')
	return true
}

// printCandidates escribe la lista numerada por los sinks de consola. La
// consola tiene sitio para la línea completa, así que imprime Description.
func printCandidates(servers []ServerCandidate) {
	out := sinks()
	out.Println("Found the following 'server's:")
	for i := range servers {
		out.Printf("%d. %s\n", i+1, servers[i].Description)
	}
}
```

El accessor de sinks se llama `sinks()`, **no** `out()`: `out` es el nombre de
la variable local que lo recibe (`out := sinks()`), y un `out()` a nivel de
paquete quedaría sombreado dentro de `SelectServer`, dando a leer dos cosas
distintas con el mismo nombre. Sus tres usos están en `console.go`, `zenity.go` y
`dialog_test.go`.

Puntos de equivalencia exacta con el código actual que **deben conservarse**:

- El bucle principal queda idéntico, incluido el mensaje literal
  `"Enter the number of the 'server' (1-%d): "` que verifica
  `select_server_test.go:80`.
- El `for` reintenta ante opción fuera de rango
  (`TestSelectServerIndexInvalidThenValidRetries`).
- El error de lectura **no** reintenta; devuelve `ok=false` (regresión
  `TestSelectServerIndexEOFReturnsNegative`).
- `ConfirmStartServer` devuelve `true` incluso con `EOF`: hoy
  `_, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')` ignora el error, y eso
  significa "continuar". Un `EOF` no debe abortar el arranque.

#### `zenity.go`

Funciones inyectables para test, siguiendo la convención `xxxFn` ya usada en
`cmd/root.go:71-133`:

```go
var (
	zenityListFn     = zenity.List
	zenityQuestionFn = zenity.Question
)

// availableOnce sondea el sistema una sola vez por proceso: la respuesta no
// puede cambiar mientras corre el launcher, e IsAvailable() hace un lookup en
// el PATH en Linux.
var availableOnce = sync.OnceValue(zenity.IsAvailable)

type zenityDialog struct{}

func (zenityDialog) Name() string { return "zenity" }

// ListCandidates es un no-op: la ventana gráfica pinta la lista por su cuenta.
func (zenityDialog) ListCandidates([]ServerCandidate) {}
```

Sólo hay dos indirecciones, `zenityListFn` y `zenityQuestionFn`, y **las dos las
overridea al menos un test** (§7.3). No existe un `zenityAvailableFn`: la
disponibilidad se sustituye reasignando el `availableOnce` entero
(`availableOnce = sync.OnceValue(func() bool { return tc.available })`, §7.1), que
además es lo único que puede saltarse el `sync.Once` ya ejecutado. Mantener un
tercer seam que ningún test usaría sería configurabilidad muerta.

Con esto, `availableOnce` es a la vez la caché de producción y el seam de test:
`var` reasignable, no un `func` en línea.

`SelectServer` mapea las descripciones a ítems **numerados**, y no devuelve el
índice directamente, sino el ítem elegido; el índice se recupera parseando el
prefijo. Esto es forzado por la API (que devuelve `string`) y es robusto ante
descripciones duplicadas:

```go
items := make([]string, len(servers))
for i, s := range servers {
	items[i] = fmt.Sprintf("%d. %s", i+1, s.label())
}
selected, err := zenityListFn(
	// <=30 chars: el control de texto mide 241 px fijos en Windows.
	"Select a 'server':",
	items,
	zenity.Title("Select 'server'"),
	zenity.OKLabel("Connect"),
	// Cancelar significa "usar ningún servidor descubierto", que es
	// exactamente lo que el caller ya hace arrancando el suyo. La etiqueta es
	// corta porque el botón mide 75 px fijos en Windows.
	zenity.CancelLabel("Start own"),
	// Solo tienen efecto en Unix; se ignoran en Windows y macOS.
	zenity.Width(600),
	zenity.Height(400),
)
```

El `label()` usado arriba es el resumen compacto del candidato
(`Description` es la línea completa):

```go
func (c ServerCandidate) label() string {
	if c.Label == "" {
		return c.Description
	}
	return c.Label
}
```

### 4.3.1 Por qué las etiquetas de la GUI son cortas

`zenity v0.10.15` no dimensiona la ventana de `List` en Windows: los tamaños
están escritos a fuego en `list_windows.go`.

| Control   | Anchura | Observaciones                       |
| --------- | ------- | ----------------------------------- |
| Texto     | 241 px  | cabecera del diálogo                |
| Lista     | 241 px  | `WS_VSCROLL`, **sin** scroll lateral |
| Botones   | 75 px   | uno por botón                      |

`zenity.Width()`/`zenity.Height()` no sirven para arreglarlo: `util.go` los
traduce a argumentos `--width`/`--height` que sólo entiende el ejecutable de
Unix, y en Windows se descartan. Con la `description` completa
(`192.168.1.10, 192.168.1.11 (a.local, b.local) - 5 ms (v1.11.0)`) el final se
salía de los 241 px: se perdían la latencia y la versión, que es justo lo que
sirve para elegir.

De ahí las tres decisiones de copy:

| Elemento  | Valor                        | Motivo                          |
| --------- | ---------------------------- | ------------------------------- |
| Items     | `Label` (`IP - N ms (ver.)`) | cabe en 241 px                  |
| Cabecera  | `Select a 'server':`         | cabe en 241 px                  |
| Cancelar  | `Start own`                  | cabe en 75 px                   |

Ninguna de las tres pierde información frente a la versión anterior: lo que se
sacrifica (IPs alternativas y hostnames) sigue visible en la consola, y la
cancelación sigue significando exactamente lo mismo ("ningún servidor
descubierto", o sea arrancar el propio).

Helper de recuperación:

```go
// indexFromItem reverses the "%d. %s" numbering applied to the list items.
func indexFromItem(item string) (int, bool) {
	i := strings.Index(item, ". ")
	if i <= 0 {
		return 0, false
	}
	n, err := strconv.Atoi(item[:i])
	if err != nil || n < 1 {
		return 0, false
	}
	return n - 1, true
}
```

`ConfirmStartServer`:

```go
err := zenityQuestionFn(text,
	zenity.Title("Start 'server'"),
	zenity.OKLabel("Start"),
	zenity.CancelLabel("Cancel"),
	zenity.QuestionIcon,
	zenity.Width(500), // Solo Unix.
)
```

**Tabla de degradación.** Es la parte crítica del diseño: cualquier fallo del
backend gráfico en tiempo de ejecución cae de vuelta a la consola, sin dejar al
usuario sin prompt.

| Resultado de zenity | `SelectServer` | `ConfirmStartServer` |
|---|---|---|
| `nil` + índice parseable | `(índice, true)` | `true` |
| `nil` + índice no parseable | log + **delega en consola** | — |
| `zenity.ErrCanceled` | `(0, false)` → el caller arranca servidor propio | `false` → aborta con `ErrServerStartCanceled` |
| `zenity.ErrUnsupported` u otro error | log + **delega en consola** | log + **delega en consola** |

La asimetría `ErrCanceled` entre I1 e I2 es intencionada y no un descuido:
- En I1, "cancelar" significa literalmente "ninguno de estos", y el flujo ya
  tiene ese camino (`selectServerIndex` ya devolvía `-1` en EOF y
  `cmd/root.go:649-654` ya arrancaba servidor propio). Se hace explícito en el
  `CancelLabel`.
- En I2 no existe camino de cancelación; hay que crearlo (ver §5).

`zenity.ErrExtraButton` no puede producirse porque no se pasa
`zenity.ExtraButton(...)`; si appearance algún día, cae en "otro error" y degrada
a consola.

### 4.2 Reubicación de `selectServerIndex`

La comprobación `singleAutoSelect && len == 1` se **extrae** de
`selectServerIndex` y sube a un helper testeable, de modo que ambas
implementaciones compartan exactamente la misma semántica y no haya
duplicación del mensaje `"Auto-selecting the only found 'server'."`:

En `cmdUtils/server.go`:

```go
// selectDiscoveredServer resolves which of the processed servers to use.
// It returns the 0-based index into procServers and false when the user
// declined to pick one, in which case the caller falls back to starting its
// own server.
func selectDiscoveredServer(procServers []*processedServer, singleAutoSelect bool, stdin io.Reader) (int, bool) {
	candidates := make([]dialog.ServerCandidate, len(procServers))
	for i, procServer := range procServers {
		candidates[i] = dialog.ServerCandidate{
			Description: procServer.description,
			Label:       procServer.label,
		}
	}
	if singleAutoSelect && len(procServers) == 1 {
		// Auto-selecting still lists the candidate first: that is what the
		// console has always done before this shortcut, and the backend that
		// would have rendered the list is not rendering anything now.
		dialog.Active().ListCandidates(candidates)
		logger.Println("Auto-selecting the only found 'server'.")
		return 0, true
	}
	return dialog.Active().SelectServer(candidates, stdin)
}
```

Los candidatos se construyen **antes** del atajo a propósito: es lo que
permite que la rama de auto-select liste el candidato (ver `ListCandidates`,
§4.1) y lo que hace que la lista se construya una sola vez para las dos ramas.

`processedServer` lleva las dos formas porque cada backend necesita una:

```go
type processedServer struct {
	server.MesuredIpAddress
	id          uuid.UUID
	description string
	// label es la versión compacta de description, sin las IPs alternativas ni
	// los hostnames, para los diálogos con poco ancho.
	label string
}
```

`description` se imprime en consola (caben todas las IPs y los hostnames);
`label` es sólo `IP - N ms (versión)`, en el orden que decide la elección.

Y `DiscoverServersAndSelectBestIpAddr` (`server.go:94-110`) pasa a:

```go
if procServers := processedServers(gameTitle, servers); len(procServers) > 0 {
	idx, ok := selectDiscoveredServer(procServers, singleAutoSelect, os.Stdin)
	if i := usableServerIndex(idx, ok, len(procServers)); i >= 0 {
		selectedServer := procServers[i]
		ip = selectedServer.Ip
		id = selectedServer.id
	}
}
```

con la guarda de rango extraída a un helper testeable:

```go
// usableServerIndex turns the dialog answer into a safe index into
// procServers. It returns -1 when the user declined, or when a backend
// reports an index that is out of range, so the caller starts its own
// 'server' instead of indexing out of bounds.
func usableServerIndex(idx int, ok bool, procCount int) int {
	if !ok || idx < 0 || idx >= procCount {
		return -1
	}
	return idx
}
```

**Por qué un helper y no `ok && idx >= 0` en línea.** `idx` viene de una
interfaz, no de un `switch` cerrado sobre el mismo fichero: con
`Dialog='auto'`, el backend puede ser `zenity`, que parsea el índice del texto
del ítem elegido, y el parseo devuelve `ok=false` en cuanto no encaja, pero un
backend futuro podría devolver `ok=true` con un índice fuera de rango. El
código anterior confiaba en que eso no pasaba porque el único productor posible
era el `switch` local. La versión con `usableServerIndex` deja el invariante
explícito y comprobable: `procServers` sólo se indexa con algo que
`0 <= i < len(procServers)`, y el caso degenerado cae en el mismo camino
"arranco mi propio servidor" que un `ErrCanceled`. Cubierto por
`TestUsableServerIndexRejectsOutOfRangeIndex`.

`func selectServerIndex` **se elimina** de `server.go` (pasa a `dialog/console.go`).
La firma de `DiscoverServersAndSelectBestIpAddr` **no cambia**, por lo que el
seam de DI `discoverServersFn` (`cmd/root.go:111`) y
`runRoot_test.go` quedan intactos.

La guarda de rango se aplica siempre, también cuando el diálogo es el de
consola y ya-ha-devuelto un índice válido: es el precio de que el índicecruce
una frontera de interfaz.

### 4.3 Cambios en `cmd/root.go`

1. **Import**: `"github.com/luskaner/ageLANServer/launcher/internal/dialog"`.

2. **Seam de DI**: añadir al bloque de `var` de `cmd/root.go:82-133`:
   ```go
   dialogNewFn = dialog.New
   ```
   Imprescindible: `runRoot` reinstala el diálogo con `dialog.Set`, así que un
   test no puede inyectar un `Dialog` falso sólo con `dialog.Set`. Overrideando
   `dialogNewFn` el test controla el backend completo, exactamente igual que
   `discoverServersFn` o `configStartServerFn`.

3. **Flag**, insertado tras el flag `gameConfig` (`root.go:139`), por ser el
   primero con alcance `[Config]`:
   ```go
   fs.StringP("dialog", "d", autoValue, `Whether to ask the interactive questions (which 'server' to use, and whether to start one) in a graphical window instead of in the console, "auto" uses graphical dialogs if they are available in the system, "false" always asks in the console. It always falls back to the console if graphical dialogs are unavailable.`)
   ```
   La abreviatura `-d` está libre (las usadas son `t c b e m p a o n g s z r l i`).

4. **Defaults** en `initConfig` (`root.go:779-805`), antes de `"Config.CanAddHost"`:
   ```go
   "Config.Dialog": autoValue,
   ```

5. **Binding** flag→clave (`root.go:809-829`):
   ```go
   "dialog": "Config.Dialog",
   ```
   La variable de entorno `ageLANServerLauncher_Config__Dialog` **no** queda
   cubierta de forma automática. Verificado empíricamente contra el binario: ni
   `ageLANServerLauncher_Config__Dialog`, ni
   `ageLANServer_Launcher_Config__Dialog`, ni
   `AGELANSERVER_LAUNCHER_CONFIG_DIALOG` cambian el backend, y tampoco lo hacen
   las formas equivalentes de opciones preexistentes (`Config.CanAddHost`,
   `Server.Start`). La causa está en `common/config.go:126`: `koanfEnvProvider`
   construye el prefijo en mayúsculas (`AGELANSERVER_LAUNCHER_`) y el proveedor
   `env` compara el prefijo de forma sensible a mayúsculas, mientras que
   `TransformFunc` conserva las mayúsculas del nombre de la variable, así que la
   clave resultant (`CONFIG.DIALOG`) nunca casa con la clave de koanf
   (`Config.Dialog`). Arreglarlo exige tocar `common/`, fuera del alcance de esta
   especificación; el flag `-d` y la clave de TOML funcionan correctamente.

6. **Validador**, junto a los demás (`root.go:872-924`):
   ```go
   func validateDialogValue(dialogMode string) (exitCode int) {
       if !autoTrueFalseValues.Contains(dialogMode) {
           logger.Printf("Invalid value for dialog (auto/true/false): %s\n", dialogMode)
           return internal.ErrInvalidDialog
       }
       return common.ErrSuccess
   }
   ```
   Reutiliza el conjunto `autoTrueFalseValues` (`root.go:65`), igual que
   `validateServerStartValue`.

7. **Resolución del backend**, un único bloque insertado en `runRoot`
   inmediatamente **después** del `defer` de teardown (`root.go:266-274`) y
   antes de `writeFileLogFn(gameId, "start")` (`root.go:275`):
   ```go
   if ec := validateDialogValue(cfg.Config.Dialog); ec != common.ErrSuccess {
       atomicExitCode.Store(int32(ec))
       return
   }
   dialogResolution := dialogNewFn(cfg.Config.Dialog)
   dialog.SetOutput(dialog.Output{Println: logger.Println, Printf: logger.Printf})
   dialog.Set(dialogResolution.Dialog)
   defer dialog.Reset()
   logger.Printf("Dialog backend: %s.\n", dialogResolution.Name)
   if dialogResolution.Reason != "" {
       logger.Println(dialogResolution.Reason)
   }
   ```

   Dos decisiones de colocación, ambas deliberadas:

   - **Va después** del `defer` de teardown, no antes. El `pid lock` se toma en
     `root.go:211` y sólo se suelta dentro de `teardown`; un `return` anterior
     al `defer` lo dejaría tomado (hoy eso ya ocurre en `root.go:218-225`, un
     bug preexistente que este cambio evita extender).
   - **`defer dialog.Reset()`** se registra **después** del `defer` de teardown,
     de modo que por LIFO se ejecuta **antes** que él. Es seguro: `teardown`
     llama a `config.Revert()` y `logger.WriteFileLog()`, ninguno de los cuales
     usa diálogos.

   Validar y resolver aquí (juntos, una sola vez) evita el estado intermedio
   raro en el que se instaló un diálogo de consola y luego se aborta por un
   valor inválido.

8. **Confirmación de arranque**, reemplaza `root.go:697-706`:
   ```go
   if cfg.Server.Start == autoValue && !cfg.Server.StartWithoutConfirmation {
       str := "No 'server's were found, proceeding to"
       if runBattleServerManager {
           str += " start a battle server (if needed) and then"
       }
       str += " start the 'server'."
       if !dialog.Active().ConfirmStartServer(str, os.Stdin) {
           logger.Println("Canceled starting the 'server'.")
           atomicExitCode.Store(int32(internal.ErrServerStartCanceled))
           return
       }
   }
   ```

   El `if !cfg.Server.StartWithoutConfirmation` se sube al mismo nivel que el
   `if cfg.Server.Start == autoValue`; la semántica es idéntica y evita anidar.
   `bufio` deja de usarse en `root.go` → quitar el import `bufio` (`root.go:4`)
   si no queda otra referencia.

   Cancelar devuelve por el `defer` de `root.go:266` con
   `atomicExitCode != 0`, luego `teardown(false)` ejecuta
   `config.Revert()` (`root.go:259-261`). Es el comportamiento correcto: el
   `setupCommand` ya se ejecutó y el estado del sistema ya se tocó.

### 4.4 Cambios en el resto de ficheros

| Fichero | Cambio |
|---|---|
| `launcher/internal/config.go` | Añadir `Dialog string` al struct `Config` (junto a `Log`, ~línea 20). **Sin tag `koanf`**: el resto de escalares del struct (`CanBroadcastBattleServer`, `SetupCommand`, ...) tampoco lo llevan, y `Config.Dialog` se empareja por el nombre del campo igual que ellos. Un tag explícito aquí sería la única excepción de estilo del struct. |
| `launcher/internal/errors.go` | **Añadir al final** del bloque `iota`, después de `ErrFlushCache`: `ErrInvalidDialog` y `ErrServerStartCanceled`. Añadir al final es obligatorio: `ErrInvalidCanTrustCertificate = iota + launcherCommon.ErrLast` (`errors.go:8`) numeraría hacia arriba y **rompería todos los códigos de salida existentes**. |
| `launcher/internal/cmdUtils/server.go` | Añadir import de `dialog`; eliminar `selectServerIndex`; añadir `selectDiscoveredServer`; reescribir el cuerpo de `DiscoverServersAndSelectBestIpAddr` (§4.2). |
| `launcher/resources/config.toml` | Añadir la clave `Dialog` documentada en `[Config]`, después de `Log = false` (línea 11). |
| `launcher/README.md` | Añadir una viñeta en `## Features` → subsección nueva `## Dialogs`, y mentionar el flag `-d` en `## Command Line`. |
| `launcher/go.mod`, `launcher/go.sum`, `go.work.sum` | Sólo por `go get`. Ni `go mod tidy` ni `go work sync`: rompen el grafo del workspace (ver el aviso en §7.7). Sin ediciones manuales. |

`launcher/resources/config.game.toml` **no** se toca: es la plantilla por juego y
sólo replica `[Config.SetupCommand]`/`[Config.RevertCommand]`, que no son
presentables.

`launcher/internal/cmdUtils/logger/log.go` **no** se toca. La sección de
diagnóstico "backend de diálogo" queda cubierta porque `logger.Printf("Dialog
backend: %s.\n", ...)` se escribe en el fichero de log a través de
`commonLogger`. Añadir una sección nueva obligaría a que `cmdUtils/logger`
importase `dialog`, lo que es justo la dependencia prohibida en §4.1.

`internal/game/battleServerBroadcast/*` no se toca: no participa.

## 5. Nuevo código de salida

| Constante | Valor | Disparador |
|---|---|---|
| `internal.ErrInvalidDialog` | siguiente a `ErrFlushCache` + 1 | `Config.Dialog` ∉ {`auto`, `true`, `false`}. |
| `internal.ErrServerStartCanceled` | siguiente + 2 | El usuario cierra/cancela el diálogo `ConfirmStartServer`. |

`ErrServerStartCanceled` es el único hueco de comportamiento real que exige
esta especificación: el prompt actual sólo tiene "continuar", así que "no
continuar" necesita un código de salida. Cancelar **no** es "continuar
ignorando la respuesta" ni "arrancar el servidor igualmente"; abortar con un
código distinto de 0 es lo único que activa el `config.Revert()`.

## 6. Configuración resultante

```toml
[Config]
# Whether to ask the interactive questions (which 'server' to use, and whether to start
# one) in a graphical window instead of in the console.
# 'auto': use graphical dialogs when they are available in the system, otherwise use the console.
# 'true': same as 'auto', but a warning is printed when falling back to the console.
# 'false': always ask in the console.
# Note: graphical dialogs are always available in Windows and macOS. In Linux-like systems
# they require 'qarma', 'zenity' or 'matedialog' to be installed.
Dialog = 'auto'
```

Casos que el usuario puede provocar y qué debe ocurrir:

| Configuración | Resultado esperado |
|---|---|
| `Dialog = 'auto'` en Windows/macOS | Diálogo gráfico siempre. |
| `Dialog = 'auto'` en Linux sin zenity | Consola, sin ruido en el log más allá de `"Dialog backend: console."`. |
| `Dialog = 'true'` en Linux sin zenity | Consola + `"Graphical dialogs are not available in this system, using the console instead."`. El launcher **no** aborta. |
| `Dialog = 'false'` en Windows | Consola. Idéntico a la versión actual. |
| `Dialog = 'true'` + `Server.SingleAutoSelect = true` con 1 servidor | Sin prompt en ningún backend; se auto-selecciona. |
| `Dialog = 'true'` + `Server.StartWithoutConfirmation = true` | Sin prompt I2 en ningún backend. |
| `Dialog = 'xxx'` | `ErrInvalidDialog`, sin tocar el sistema. |

## 7. Plan de pruebas

### 7.1 `launcher/internal/dialog/dialog_test.go` (nuevo)

| Test | Qué fija |
|---|---|
| `TestNewResolvesMode` | La tabla completa de §4.1 (`mode` × disponible → `Name`/`Reason`). Se sobrescribe `availableOnce = sync.OnceValue(func() bool { return tc.available })` por caso; hay que poder reasignar `availableOnce`, así que se declara como `var`, no como `func` en línea. |
| `TestActiveDefaultsToConsole` | `Active()` sin `Set` devuelve `console`. |
| `TestSetAndReset` | `Set(consoleDialog{})` → `Name()=="console"`; `Reset()` → `Active().Name()=="console"`. |
| `TestIndexFromItem` | `"1. a"`→`0`, `"3. x"`→`2`, `"x"`→`false`, `". a"`→`false`, `"0. a"`→`false`, `"a. b"`→`false`. |

### 7.2 `launcher/internal/dialog/console_test.go` (nuevo)

Porta literal de los tests existentes, más los nuevos:

- `TestConsoleSelectServerValidInput` ← `TestSelectServerIndexValidInput`.
- `TestConsoleSelectServerInvalidThenValidRetries` ← `TestSelectServerIndexInvalidThenValidRetries`.
- `TestConsoleSelectServerEOFReturnsNotOK` ← `TestSelectServerIndexEOFReturnsNegative`
  (adapta la aserción: ahora es `ok == false`, ya no el centinela `-1`).
- `TestConsoleSelectServerPrintsServerList` ← `TestSelectServerIndexPrintsServerList`,
  comprobando las mismas cadenas, incluida
  `"Enter the number of the 'server' (1-3): "`, con la copia local de
  `captureStdout` de este paquete (ver la nota sobre `captureStdout` en §7.6).
- `TestConsoleConfirmStartServerEOFReturnsTrue` — regresión que fija el
  comportamiento heredado de ignorar el error de `ReadBytes`.
- `TestConsoleConfirmStartServerPrintsPrompt`.
- `TestConsoleListCandidatesPrintsTheSameList` — `ListCandidates` imprime
  exactamente el bloque que `SelectServer` imprime antes de su prompt, y nada
  más (el llamador añade lo que venga después).

### 7.3 `launcher/internal/dialog/zenity_test.go` (nuevo)

Con `zenityListFn` / `zenityQuestionFn` sustituidas por dobles que capturan
`text`, `items` y devuelven el errorcodificado:

- `TestZenitySelectServerReturnsIndexedItem` — `items` == `["1. a", "2. b"]`,
  devuelve `"2. b"` → `(1, true)`.
- `TestZenitySelectServerUsesCompactLabelNotFullDescription` — regresión del
  truncado de los 241 px: el item es `Label`, y además se afirma que **no**
  contiene los hostnames de `Description`.
- `TestZenitySelectServerFallsBackToDescriptionWithoutLabel` — sin `Label` se
  imprime `Description`, nunca una fila vacía.
- `TestZenitySelectServerCancelReturnsNotOK` — `ErrCanceled` → `(0, false)`;
  además se pasa un `stdin` con `"1\n"` para comprobar que **no** se consulta
  (si se delegase, devolvería `ok=true`).
- `TestZenitySelectServerUnsupportedFallsBackToConsole` — `ErrUnsupported` +
  `stdin` `"1\n"` → `(0, true)`.
- `TestZenitySelectServerUnparsableItemFallsBackToConsole`.
- `TestZenitySelectServerOtherErrorFallsBackToConsole`.
- `TestZenityConfirmOK`, `TestZenityConfirmCancelReturnsFalse`,
  `TestZenityConfirmOtherErrorFallsBackToConsole`.
- `TestZenityListCandidatesIsSilent` — `ListCandidates` no imprime nada: la
  ventana gráfica pinta la lista por su cuenta y no debe duplicarse en consola.

Nota: las opciones `zenity.Option` no son introspeccionables (aplican a un
`options` no exportado), así que los asserts se limitan a `text`, `items` y al
resultado. Es una limitación conocida y aceptada; no se prueba `Title`/`OKLabel`.

### 7.4 `launcher/internal/cmdUtils/select_server_test.go` (reescrito)

El fichero pasa a `server_dialog_test.go` (o se conserva el nombre y se
reemplaza su contenido). `selectServerIndex` ya no existe; los tests se
reescriben contra `selectDiscoveredServer`, que sí es testeable porque no toca
red:

- `TestSelectDiscoveredServerAutoSelectSingle` — `singleAutoSelect=true`, 1
  servidor, `stdin` vacío → `(0, true)`. Cubre el antiguo
  `TestSelectServerIndexAutoSelectSingle`, ahora en su nivel real. Llama a
  `dialog.Reset()` + `t.Cleanup(dialog.Reset)` porque esta rama depende del
  backend global (`ListCandidates`): sin aislar, el resultado dependería del
  orden de ejecución de los tests del paquete.
- `TestSelectDiscoveredServerAutoSelectIgnoredWhenMultiple` — 3 servidores,
  `singleAutoSelect=true` → **sí** pregunta.
- `TestSelectDiscoveredServerAutoSelectNeedsNoPrompt` — el `Dialog` falso
  registra si se le llamó a `SelectServer`: el atajo no pregunta.
- `TestSelectDiscoveredServerAutoSelectListsCandidate` — el atajo llama a
  `ListCandidates` con las descripciones (ver §4.1).
- `TestSelectDiscoveredServerAutoSelectConsoleOutputUnchanged` — compara la
  salida del atajo **byte a byte** contra
  `"Found the following 'server's:\n1. solo\nAuto-selecting the only found 'server'.\n"`.
- `TestUsableServerIndexRejectsOutOfRangeIndex` — tabla del guard de rango
  (`0`, `n-1`, `n`, `99`, `-1`, y `ok=false`).
- `TestSelectDiscoveredServerDelegatesDescriptionsInOrder` — se instala un
  `Dialog` falso con `dialog.Set` y se comprueba que recibe las descripciones
  en el orden de latencia y que su índice se respeta.
- `TestSelectDiscoveredServerPassesBothDescriptionAndLabel` — cada candidato
  llega con las dos formas, y además se afirma que `Label` es más corta que
  `Description` (si no, se recortaría igual).
- `TestSelectDiscoveredServerCancelledReturnsNotOK` — el diálogo falso devuelve
  `ok=false` → `(0, false)`.
- `TestSelectDiscoveredServerFallsBackToConsoleByDefault` — sin `dialog.Set`,
  `stdin` `"2\n"` → `(1, true)`.

Todos los que instalan un `Dialog` falso usan el helper `installFakeDialog`, que
hace `dialog.Reset()` antes y `dialog.Reset` en el `t.Cleanup`.

`testProcessedServers` y `serverMesuredIpAddress` se reutilizan sin cambios.

### 7.5 `launcher/internal/cmd/runRoot_test.go` (añadidos)

Todos requieren `dialog.Reset()` en el `t.Cleanup` que ya usa `applyOverrides`,
porque `runRoot` deja un `dialog` instalado en el proceso global:

- `TestRunRootInvalidDialogValue` — `Config.Dialog = "xxx"` →
  `exitCode == internal.ErrInvalidDialog`, y `configStartServerFn` no se invoca.
- `TestRunRootUsesConsoleWhenDialogsUnavailable` — `dialogNewFn` devuelve
  `Resolution{Dialog: fake, Name: "console", Reason: "..."}`; el `Reason`
  aparece en la salida capturada.
- `TestRunRootServerStartCanceled` — `dialogNewFn` devuelve un `Dialog` cuyo
  `ConfirmStartServer` devuelve `false`; asserts: `exitCode ==
  internal.ErrServerStartCanceled` y `configStartServerFn` **no** se invoca.
- `TestRunRootServerStartConfirmed` — `ConfirmStartServer` devuelve `true` →
  `configStartServerFn` se invoca con los valores esperados.
- `TestRunRootServerStartWithoutConfirmationSkipsDialog` — con
  `StartWithoutConfirmation=true`, `ConfirmStartServer` **no** se invoca
  (doble que devuelva `false`, para que un fallo del test se manifieste como
  código de salida inesperado, no como un falso positivo).

### 7.6 Nota sobre `captureStdout`: tres copias, y por qué se quedan

`captureStdout` (redirigir `os.Stdout` por un `os.Pipe`, drenarlo en goroutine y
restaurar) existe hoy en tres paquetes de test distintos —
`dialog/console_test.go`, `cmdUtils/select_server_test.go` y
`cmd/runRoot_test.go`— porque antes de este trabajo vivía en uno solo
(`cmdUtils`) y al partir la lógica de prompts entre paquetes se replicó. Se
decide **dejarlo triplicado** en vez de consolidarlo, por dos razones:

- Consolidarlo exigiría un `internal/testutil` no-test, es decir un paquete que
  entra en `go build ./...` y `go vet ./...` e importa `testing` sólo para
  evitar 25 líneas de boilerplate que nunca se empaqueta. Es un coste
  arquitectónico permanente en un módulo que se distribuye como binario a cambio
  de ~50 líneas netas ahorradas. Go no permite paquetes importables sólo desde
  test, así que no existe la variante "sólo-test".
- Es boilerplate estable de `os.Pipe`, idéntico por construcción (copy-paste), y
  cada copia muta un global distinto: `os.Stdout` en `dialog` y en `cmdUtils`
  capturan la salida de `logger`, mientras que el `captureStdout` de `dialog`
  además cubre los sinks por defecto del propio paquete. Que divergan no rompe
  nada: lo detectan los propios tests que asertan sobre cadenas exactas
  (`TestConsoleSelectServerPrintsServerList`,
  `TestSelectDiscoveredServerAutoSelectConsoleOutputUnchanged`,
  `TestRunRoot*`).

Lo que no es aceptable es una **cuarta** copia: cualquier helper nuevo que
necesite capturar stdout debe ir al paquete que lo usa, no a un denominador
común que no existe.

### 7.7 Verificación

```powershell
# Dependencias (desde F:\ageLANServer\launcher, por el go.work)
go get github.com/ncruces/zenity@v0.10.15

# OJO: go mod tidy y go work sync NO son seguros en este workspace.
#   - `go mod tidy` intenta resolver los paquetes internos de `common` y
#     `launcher-common` desde el proxy y falla con
#     "module github.com/luskaner/ageLANServer/common@latest ... does not
#     contain package .../common/uuid".
#   - `go mod tidy -e` no falla, pero es peor: añade requirements de
#     `github.com/luskaner/ageLANServer/common` y `.../launcher-common` al
#     go.mod del launcher.
#   - `go work sync` propaga esos cambios a los 17 módulos del workspace.
# Se conserva por tanto el grafo que produce `go get` (zenity directa;
# dchest/jsmin y x/image indirectas) y se dejan los tool-deps de zenity
# (goversioninfo, makefat, rsrc) sin podar: son hashes sobrantes en go.sum,
# inofensivos, que `go mod tidy` eliminaría. Sin ediciones manuales de
# go.mod/go.sum más allá del `go get`.

# Tests e Hygiene (desde F:\ageLANServer)
$env:CGO_ENABLED = "1"
go test ./launcher/... -race      # varias regresiones lo exigen
go vet ./launcher/...
gofmt -l ./launcher/internal/dialog ./launcher/internal/cmdUtils

# Matriz de compilación que exige .github/workflows/codeql.yml:43
# OJO: en PowerShell `go build -o $null ./launcher` NO vale. $null se expande a
# cadena vacía, así que `-o` se queda sin valor, `./launcher` se consume como
# ruta de salida y no queda ningún paquete que compilar:
#     no Go files in F:\ageLANServer
$env:CGO_ENABLED = "0"
$tmp = Join-Path $env:TEMP "launcher-build-matrix"
New-Item -ItemType Directory -Force -Path $tmp | Out-Null
foreach ($os in @("windows","linux","darwin")) {
  foreach ($arch in @("amd64","arm64")) {
    $env:GOOS = $os; $env:GOARCH = $arch
    go build -o (Join-Path $tmp "launcher-$os-$arch") ./launcher
    if ($LASTEXITCODE -ne 0) { Write-Output "FAIL $os/$arch" }
    else { Write-Output "OK   $os/$arch" }
  }
}
```

Los tests de la §7.3 no deben abrir diálogos reales: todos los puntos de
entrada de `zenity` en `zenity.go` pasan por `zenityListFn` /
`zenityQuestionFn`, que los tests sustituyen.

## 8. Orden de implementación

1. `dialog/dialog.go` + `dialog/console.go`, con `console_test.go` en verde y
   **sin tocar nada más**. En este punto el launcher no cambia de
   comportamiento.
2. `cmdUtils/server.go`: `selectDiscoveredServer`, `dialog.Active()`,
   eliminar `selectServerIndex`, actualizar `select_server_test.go`.
3. `cmd/root.go`: `dialogNewFn`, flag, default, binding, validador, bloque de
   resolución, confirmación de arranque.
4. `internal/config.go`, `internal/errors.go`, `resources/config.toml`,
   `README.md`.
5. `go get` únicamente. **No** `go mod tidy` ni `go work sync`: rompen el
   grafo del workspace (ver el aviso en §7.7).
6. `dialog/zenity.go` + `zenity_test.go`, `config_defaults_test.go` extendido
   con `"Config.Dialog"`, y los tests de `runRoot_test.go` de la §7.5.
7. Verificación completa de la §7.7, incluida la matriz de compilación.

Los pasos 1-4 son un incremento funcional por sí solo (equivalente a `Dialog =
'false'` en todas partes); el 6 es lo que activa los diálogos.

## 9. Riesgos

| Riesgo | Mitigación |
|---|---|
| `golang.org/x/image` engorda la build de Windows/macOS. | Aceptado: el enlazador descarta lo no usado; los binarios reales sólo cargan los diálogos de lista y mensaje. **Delta medido: +3,05 % en windows/amd64, ver §10.1.** |
| En Linux, `zenity` no instalado ⇒ `IsAvailable()==false`. | `Dialog='auto'` (por defecto) cae a consola en silencio. `Dialog='true'` cae a consola con aviso. Ningún modo bloquea el arranque. |
| `zenity.List` devuelve `string`; no se puede devolver un índice. | Los ítems se numeran `"%d. %s"` y el índice se parsea del prefijo, con test dedicado. |
| Descripciones duplicadas entre servidores. | Resuelto por la numeración del ítem; cada ítem es único por posición. |
| Regresión en la salida de los prompts: dejarían de estar en el fichero de log. | `dialog.SetOutput` recibe `logger.Println`/`logger.Printf`; `TestConsoleSelectServerPrintsServerList` sigue verificando las cadenas exactas. |
| `IsAvailable()` hace búsquedas en el PATH en cada prompt. | Se evalúa una sola vez con `sync.OnceValue`. |
| Numeración de `internal/errors.go` rota por añadir códigos en medio. | Se añade **siempre** al final del bloque `iota`; verificado en §4.4. |
| Un diálogo gráfico bloqueado se cuelga sin salida. | `Ctrl+C` sigue por el handler de señales de `root.go:501-513` (fuerza `teardown` + `osExit`), igual que hoy. |

## 10. Criterios de aceptación

- [ ] `go test ./launcher/... -race` en verde; `gofmt -l` limpio en los
      ficheros tocados por este trabajo.
- [ ] Compila en `windows`/`linux`/`darwin` × `amd64`/`arm64` con
  `CGO_ENABLED=0`.
- [ ] Con `Dialog = 'false'`, el flujo observable (stdout + fichero de log +
  códigos de salida) es **idéntico** al previo al cambio, **más** una única
  línea añadida al inicio: `"Dialog backend: console."`. Ésa es la diferencia
  real y verificable; el resto del flujo no debe registrar ni una línea más.
- [ ] En Windows y macOS, con `Dialog = 'auto'`, I1 e I2 aparecen como diálogos
      nativos y el launcher continúa correctamente en ambos casos.
- [ ] Cancelar I1 arranca servidor propio; cancelar I2 aborta con
      `ErrServerStartCanceled` y el sistema queda revertido.
- [ ] Un error de `zenity` (binary ausente, display no disponible,
      `ErrUnsupported`) degrada al prompt de consola sin perder el flujo.
- [ ] `launcher -h` muestra `-d, --dialog`; un valor inválido desde TOML o
      desde el CLI produce `ErrInvalidDialog`. No desde env: la capa de variables
      de entorno está rota para claves mixtas (ver §4.3.5).

### 10.1 Delta de binario medido

Riesgo de §9 medido sobre `windows/amd64`, `CGO_ENABLED=0`, sin flags de
ldflags (build limpio de `go build -o <out> ./launcher`):

| Build | Bytes |
|---|---|
| Base (`HEAD`, sin este trabajo) | 16.082.432 |
| Con los diálogos zenity | 16.572.416 |
| **Delta** | **+489.984 bytes** |

**+489.984 bytes = +478,5 KiB ≈ +490 KB, +3,05 %.**

Cómo reproducirlo, sin ensuciar el árbol de trabajo:

```powershell
git worktree add --detach $tmp HEAD
$env:CGO_ENABLED = "0"; $env:GOOS = "windows"; $env:GOARCH = "amd64"
go -C $tmp build -o base.exe ./launcher
go build -o new.exe ./launcher
(Get-Item new.exe).Length - (Get-Item base.exe).Length
git worktree remove --force $tmp; git worktree prune
```

El 3,05 % es el número que confirma que el riesgo de §9 ("`golang.org/x/image`
engorda la build") está acotado y es aceptable: no se empaqueta ningún binario
extra, no hay `cgo`, y el enlazador ya descartaba lo que no se usa antes de este
cambio.
