# Especificación: salida de terminal (`lipgloss` v2, glifos y ASCII art) en `launcher` y `launcher-config*`
Estado: propuesta cerrada para implementación.
Alcance: **exclusivamente** los módulos `launcher`, `launcher-common`, `launcher-config` y
`launcher-config-admin`. **Quedan fuera** `launcher-agent`, `launcher-config-admin-agent`,
`server`, `battle-server-manager`, `battle-server-broadcast`, `server-genCert`, `config-helper`,
`tools/*` y el módulo `common` (compartido).

---

## 1. Objetivo

Hacer la salida de consola de `launcher`, `config` y `config-admin` más usable y bonita
usando `lipgloss`, y añadir un lenguaje de glifos (emoji + símbolos Unicode) y de arte ASCII,
**sin perder nunca la información que hoy se imprime**, y garantizando explícitamente que
una terminal antigua —`cmd.exe` de Windows 7, con página de códigos 437 y sin soporte de
VT— siga mostrando, como mínimo, exactamente el estado actual del output.

Tres requisitos, en orden de prioridad:

| # | Requisito | Cómo se cumple |
|---|---|---|
| R1 | **Paridad en terminal antigua**: `cmd.exe` Win7 muestra al menos el output actual | Un único concepto de *tier de glifo* (§6.1). En `TierASCII` no se emite **ni un solo byte** fuera de ASCII ni de ANSI, y el texto de cada mensaje conserva todos sus tokens de información (§7) |
| R2 | Output más bonito y usable en terminales modernas | `lipgloss` v2 para estilos/anchos/tablas, glifos por tier, banner, alineación clave-valor, resumen final |
| R3 | Cero cambios en los logs de fichero y en los módulos fuera de alcance | Estilado **solo** en el sink de consola; los sinks de fichero y `common/logger` no se tocan (§6.5) |

Restricción no negociable heredada del repo (ver `launcher/ZENITY.md` §7.7): **`go mod tidy` y
`go work sync` no son seguros en este workspace**. Hay que usar `go get` desde el directorio del
módulo, y `go.mod`/`go.sum` no se editan a mano.

---

## 2. Estado actual (análisis)

### 2.1 No hay estilado en absoluto

| Búsqueda | Resultado |
|---|---|
| Imports de `lipgloss` / `termenv` en `.go` del repo | **0** |
| `lipgloss` en el grafo de algún binario distribuido | **0**. La única aparición es `tools/scripts/go.mod:22` como `// indirect` de `goreleaser/v2` |
| Secuencias ANSI en código Go | **0** |
| Emoji / símbolos Unicode / box-drawing en salida de usuario | **0** |
| Arte ASCII / banner / figlet | **0** |

Los emoji que existen en el repo están solo en `README.md` (documentación) y en `t.Logf` de
tests (`server/internal/routes/router/httptest_complex_flows_test.go`, 94 `t.Logf("✓ …→ …")`).
Ningún test afirma sobre emoji ni símbolos.

### 2.2 El lenguaje visual actual son tres cosas y nada más

1. **Prefijo de log** `|NOMBRE| ` — `common/logger/logger.go:76-78`:
   ```go
   func Prefix(name string) {
       logger.SetPrefix("|" + strings.ToUpper(name) + "| ")
   }
   ```
   Solo se ve cuando **no** es interactivo, porque `Initialize` activa `log.Lmsgprefix` únicamente
   en ese caso (`common/logger/logger.go:65-74`). En una terminal interactiva el usuario ve
   líneas peladas, sin prefijo y sin timestamp. Ese es el output "bonito" que hay hoy.
2. **Sangría por tabuladores** `\t` y `\t\t` como único mecanismo de jerarquía, en ~40 puntos:
   `launcher-config/internal/userData/backup.go:34,57,73,88,101,109`,
   `launcher-common/configRevert.go:166,168`,
   `battle-server-manager/internal/cmd/start.go:168,171,174` (fuera de alcance).
3. **Separadores `========== SECCIÓN ==========`** —
   `launcher/internal/cmdUtils/logger/log.go:362-369`, y solo al **log de fichero**.

Marcadores de estado actuales, en texto plano:

| Marcador | Dónde | Ejemplo |
|---|---|---|
| `"OK."` | `battle-server-manager/internal/cmd/start.go:135` (fuera de alcance) | `OK.` |
| `"SKIP"` / `"OK"` | `tools/server-replay/internal/logEntry/http/http.go:52,86` (fuera de alcance) | `OK` |
| `"Successfully …"` / `"Failed …"` | `launcher-config/internal/cmd/setUp.go:25,29,36,40,48,52,58,62` | `Successfully added user certificate` |
| `"…"`, `"..."` (progresivo) | `launcher/internal/cmd/root.go:475,567,620` | `Setting up...` |

### 2.3 Detección de terminal: insuficiente y mal reutilizable

Existe exactamente un mecanismo, `common/terminal.go` (34 líneas), que responde a "¿es TTY?":

```go
func Interactive() bool {
	return terminal.IsTerminal(int(terminal.StdinFd())) &&
		terminal.IsTerminal(int(terminal.StdoutFd()))
}
```

- Su **único** consumidor es `common/logger/logger.go:70`, para decidir si poner timestamps.
- No hay ninguna detección de: `NO_COLOR`, `CLICOLOR`, `CLICOLOR_FORCE`, `TERM`, `TERM=dumb`,
  `COLORTERM`, `WT_SESSION`, `ConEmuANSI`, `ANSICON`, `TTY_FORCE`, `ENABLE_VIRTUAL_TERMINAL_PROCESSING`,
  página de códigos de salida (`GetConsoleOutputCP`), ni ancho de la consola.
- `golang.org/x/sys/windows` se usa en 20+ sitios del repo pero **nunca** para modo de consola
  o code page. El patrón `windows.NewLazySystemDLL(...)` ya está establecido en
  `common/executor/exec/ShellExecuteExW_windows.go:14-15,20-21`, así que se puede reutilizar.

### 2.4 Topología de los sinks (esto condiciona todo el diseño)

| Binario | `commonLogger.Initialize(w)` | Sink de fichero | Sink de stdout |
|---|---|---|---|
| `launcher` | `os.Stdout` → nunca; usa `Initialize(nil)` (`launcher/main.go:15`) | buffer `Buf` → volcado a `logs/.../launcher.txt` sólo con `--log` | `fmt.Printf/Println` **crudo**, en `launcher/internal/cmdUtils/logger/log.go:143-151` |
| `config` (`launcher-config`) | `os.Stdout` (`launcher-config/main.go:16`) | `internal.Logger.Buffer(...)` — **ruta independiente** | es el `log.Logger` sobre `os.Stdout` |
| `config-admin` (`launcher-config-admin`) | `os.Stdout` (`launcher-config-admin/main.go:17`) | `internal.Logger.Buffer(...)` — ruta independiente | es el `log.Logger` sobre `os.Stdout` |

Los wrappers de `launcher` escriben **a los dos**:
```go
// launcher/internal/cmdUtils/logger/log.go:143
func Printf(format string, a ...any) {
	commonLogger.PrefixPrintf("main", format, a...)  // -> Buf -> log de fichero
	fmt.Printf(format, a...)                          // -> stdout crudo
}
```

Consecuencias directas:

- Para `launcher`, **el estilado debe aplicarse sólo a la mitad `fmt.Printf`**. Si se estila antes
  de `PrefixPrintf`, los códigos ANSI y los emoji se cuelan en `launcher.txt` y rompen el
  `grep`-abilidad del log. Además `common/logger` recibe el mismo texto que hoy, así que el log
  no cambia ni un byte.
- Para `config` y `config-admin`, `commonLogger` **es** el printer de stdout. Los ficheros de log
  van por `internal.Logger.Buffer`, que recibe cadenas construidas aparte → se quedan planos sin
  tocar nada.

### 2.5 Trampa latente: envolver el writer rompe los flags del logger

`common/logger.Initialize` decide los flags comparando **punteros**:

```go
// common/logger/logger.go:70
if writer != os.Stdout || !common.Interactive() {
	flags = log.Lmicroseconds | log.Ltime | log.LUTC | log.Lmsgprefix
}
```

Si alguien "simplifica"envolviendo el writer para que `colorprofile` haga el *downsampling*
—`commonLogger.Initialize(colorprofile.NewWriter(os.Stdout, os.Environ()))`—, la comparación
`writer != os.Stdout` pasa a ser verdadera y aparecen timestamps + prefijo `|MAIN|` en una
terminal interactiva. Eso es una regresión de legibilidad, no una mejora. **Por eso el diseño de
§6 no envuelve el writer**: se estila la cadena y se deja `Initialize` intacto.

### 2.6 Tests que se romperían con un rediseño ingenuo

| Test | Afirmación | Nota |
|---|---|---|
| `launcher/internal/dialog/console_test.go:152-159` | **igualdad byte a byte**: `"Found the following 'server's:\n1. x\n2. xx\n3. xxx\n"` | se rompe con cualquier cambio de formato |
| `launcher/internal/cmdUtils/select_server_test.go:217-228` | **igualdad byte a byte** del mismo texto | ídem |
| `launcher/internal/dialog/console_test.go:106-117,128,139` | `strings.Contains` de la lista y del prompt | toleran indentación, no toleran glifos extra |
| `launcher/internal/cmd/runRoot_test.go:1733,1736` | `Contains("Dialog backend: console.")` | tolera marcador antes/después |
| `common/logger/logger_extra_test.go:25-30,45-50` | `Contains("|TEST|")`, `Contains("|MYMOD|")` | no debe verse afectado |

**Y aquí está la propiedad clave del diseño**: en `go test`, `os.Stdout` es un pipe, luego no es
TTY → la detección devuelve `TierASCII` → **los tests golden existentes siguen pasando sin
tocarlos**, siempre que en `TierASCII` el texto base sea idéntico y solo se añada el marcador
ASCII como prefijo determinista. Eso es una garantía del diseño, no una casualidad.

### 2.7 Puntos de salida por fichero (módulos en alcance, sin tests)

| Fichero | Llamadas |
|---|---|
| `launcher/internal/cmd/root.go` | 74 |
| `launcher-config/internal/cmd/setUp.go` | 54 |
| `launcher-config/internal/cmd/revert.go` | 41 |
| `launcher/internal/cmdUtils/logger/log.go` | 31 |
| `launcher-config/internal/admin/admin.go` | 30 |
| `launcher/internal/cmdUtils/root.go` | 26 |
| `launcher-config-admin/internal/cmd/setUp.go` | 21 |
| `launcher/internal/cmdUtils/{cert,game,hosts}.go` | 46 |
| `launcher-config/internal/{cacert.go,userData/backup.go,userData/*Backup.go}` | 25 |
| `launcher-common/configRevert.go` | 9 |
| `launcher-config-admin/internal/{cmd/flushCache.go,cmd/revert.go,hosts/*.go}` | 25 |

---

## 3. Hechos de terminal que condicionan el diseño (con fuente)

### 3.1 `cmd.exe` de Windows 7

| Hecho | Consecuencia de diseño |
|---|---|
| **No existe VT.** `ENABLE_VIRTUAL_TERMINAL_PROCESSING` (0x0004) fue introducido en Windows 10 1511 (build 10586). En Win7 `SetConsoleMode` con ese bit falla o no hace nada | Hay que **verificar** el bit, no asumirlo: `SetConsoleMode` y luego **re-leer** con `GetConsoleMode` y comprobar que el bit quedó activo |
| **Code page de salida por defecto = 437** (OEM). Go no lo cambia; escribe bytes UTF-8 | Cualquier byte UTF-8 sale como `?` o mojibake. Hay que detectar `GetConsoleOutputCP()` |
| **Fuentes raster** (Terminal, Lucida Console) sin cobertura emoji | Unicode ≠ Emoji. Aunque el code page sea 65001, `✅` se ve como caja. Son dos ejes distintos |
| `cmd.exe` no define `TERM` ni `TERM_PROGRAM` | No se puede decidir por `TERM` en Windows |
| Consoal de 80×25 por defecto | El banner no puede pasar de ~7 líneas, y el output de la ejecución tiene que caber en lo que queda |

### 3.2 Lo que `colorprofile` (dependencia de `lipgloss`) ya resuelve — y lo que no

`lipgloss` v2 delega el *color profile* en `github.com/charmbracelet/colorprofile`. `colorprofile.Detect(w, env)`
ya cubre (fuente: `colorprofile/env.go`, `colorprofile/env_windows.go`):

- `NO_COLOR`, `CLICOLOR`, `CLICOLOR_FORCE`, `TTY_FORCE`.
- `TERM=dumb` → `NoTTY`.
- `COLORTERM=truecolor`, `tmux`/`screen`, `terminfo` (`Tc`/`RGB`).
- **En Windows**, vía `windowsColorProfile`:
  - `ConEmuANSI=ON` → `TrueColor`.
  - **`major < 10 || build < 10586` → `NoTTY`** (salvo `ANSICON` → `ANSI`/`ANSI256`).
    👉 Esto **ya cubre Win7 `cmd.exe`** de forma correcta.
  - `build < 14931` → `ANSI256`; si no → `TrueColor`.
  - `WT_SESSION` presente → `TrueColor`.

Cuatro huecos que `colorprofile` **no** cubre y que este diseño tiene que tapar:

| Hueco | Detalle | Arreglo propuesto |
|---|---|---|
| **H1 — VT no verificado** | En Windows 10/11 con `conhost.exe` plano (sin `WT_SESSION`, sin `ConEmuANSI`, sin `TERM`), `windowsColorProfile` devuelve `TrueColor` **sin comprobar `ENABLE_VIRTUAL_TERMINAL_PROCESSING`**. Go no la habilita solo. Resultado: secuencias ANSI impresas literalmente | `ui/vt_windows.go`: verificación propia con `GetConsoleMode`/`SetConsoleMode` + re-lectura. Si no se puede verificar → sin color |
| **H2 — `NO_COLOR` vacío no desactiva color** | `envNoColor` usa `strconv.ParseBool(env.get("NO_COLOR"))`. `NO_COLOR=` (vacío) o `NO_COLOR=0` → `ParseBool` falla → **no desactiva**. La spec de <https://no-color.org> dice que la mera presencia desactiva | Detección propia por **presencia** de la variable, antes de `colorprofile` |
| **H3 — code page / emoji** | `colorprofile` no sabe nada de glifos | Eje de tier propio (§6.1) |
| **H4 — ancho de consola** | No lo expone; hace falta para banner, wrap y alineación | `GetConsoleScreenBufferInfo` en Windows, `TIOCGWINSZ`/ioctl en Unix vía `charmbracelet/x/term` |

### 3.3 API real de `lipgloss` v2 (verificada en el código)

Módulo: **`charm.land/lipgloss/v2`** (no `github.com/charmbracelet/lipgloss`). Última versión
observable en el grafo del repo: `v2.0.6`.

| Punto | Comportamiento verificado | Implicación |
|---|---|---|
| `Style.Render(str)` | **Puro**: emite exactamente el ANSI que le pediste. **No** hace *downsampling* | No se puede confiar en `Render` para degradar; hay que usar `lipgloss.Complete(profile)(ansi, ansi256, truecolor)` |
| `lipgloss.Complete(profile)` | Devuelve la función que elige color según perfil; con `NoTTY` devuelve `colorprofile.NoColor` | Es el mecanismo correcto, y `NoTTY` ⇒ texto plano: **cinturón y tirantes** junto a nuestro gating propio |
| `lipgloss.Writer` / `Print*` / `Fprint*` / `Sprint*` | `var Writer = colorprofile.NewWriter(os.Stdout, os.Environ())`; `Print*`/`Sprint*` usan `Writer.Profile` | Son **globales atados a `os.Stdout`**. No sirven para el sink de fichero ni para tests. No hacen falta: no vamos a escribir por el writer de `colorprofile` (§2.5) |
| `Style.TabWidth` | Por defecto convierte `\t` a **4 espacios** al renderizar; `TabWidth(lipgloss.NoTabConversion)` lo desactiva | Obligatorio `TabWidth(lipgloss.NoTabConversion)` en los estilos base, o los `\t` de §2.2 se convierten en espacios y cambia la sangría en Win7 |
| `lipgloss.Border` / `NormalBorder` / `RoundedBorder` / `ThickBorder` / `DoubleBorder` / `ASCIIBorder` / `MarkdownBorder` | existen | `ASCIIBorder()` para el tier ASCII, `RoundedBorder()` para tiers superiores |
| `lipgloss.Blend1D(n, c1, c2)` | degradado de `n` pasos | para el banner |
| `lipgloss.Width/Height/Size`, `JoinHorizontal`, `Wrap`, `Place`, `Layer`/`Compose` | disponibles | ancho, ajuste de línea |
| `Style.Inline(true)`, `MaxWidth`, `MaxHeight` | disponibles | forzar una sola línea |
| `lipgloss.HasDarkBackground(in, out)`, `LightDark(bool)` | disponibles | **no usar**: consulta el terminal (OCRD / CSI 11) y en Win7 vintage puede bloquear o ensuciar la consola. Los colores se eligen sin consultar el fondo |

---

## 4. Justificación de las decisiones de integración

| Decisión | Alternativas descartadas | Motivo |
|---|---|---|
| **A. `lipgloss` v2 sin Bubble Tea** | `bubbletea` + `bubbles` | El output es un **log lineal**; un ELM con alt-screen es desproporcionado y colisiona con el contrato de cleanup (§8). Se evalúa en §8 con detalle |
| **B. Núcleo en `launcher-common/ui`** | duplicar en `launcher`, `launcher-config`, `launcher-config-admin`; o meterlo en `common` | Los 4 módulos en alcance ya importan `launcher-common` (`launcher-config-admin/main.go:10`, `launcher-config/internal/cmd/setUp.go:15`, `launcher/internal/cmd/root.go:43`). `common` está compartido con `server` y `battle-server-manager`, que quedan fuera de alcance: tocarlo los contaminaría |
| **C. El estilado ocurre en el *string*, no en el *writer*** | envolver `os.Stdout` con `colorprofile.Writer` | Rompe la comparación de punteros de `common/logger.Initialize` (§2.5) |
| **D. `Tier` explícito propio, no confiar en `colorprofile` para glifos** | confiar en el perfil de color | Perfil de color y cobertura de glifos son ortogonales (§6.1) |
| **E. Bandas: `Banner`, `Rule`, `KV`, `Step`, `Ok/Fail/Warn/Info`** | `fang` | `fang` es un wrapper de **cobra** (`fang.Execute(ctx, *cobra.Command)`). El repo usa `pflag` con dispatchers propios en `common/cmd/flagSet.go`, compartido con módulos fuera de alcance |
| **F. Emoji solo en marcadores de línea, nunca en tablas** | emoji en todas partes | Los emoji son de **doble ancho** (`runewidth` = 2). Romperían la alineación de cualquier tabla o lista |
| **G. Sin `SetConsoleOutputCP` implícito** | forzar 65001 | Es un cambio **global de la consola compartida**, no del proceso; el efecto sería sorprendente para el usuario y persistente. Se documenta `chcp 65001` en `common/resources/start.bat` en su lugar |
| **H. Sin reescritura de texto en el tier ASCII por defecto** | reescribir mensajes ("Successfully added user certificate" → "Added user certificate") | La garantía de paridad se vuelve mecánica y auditable si el tier ASCII conserva las frases actuales |

---

## 5. Alcance funcional exacto

### 5.1 Dentro de alcance

1. Nuevo paquete `launcher-common/ui` con: detección de capacidades, catálogo de glifos, tema
   `lipgloss`, helpers de impresión y utilidades de ancho/sanitizado.
2. Estilado del sink de **consola** de `launcher`, `config` y `config-admin`.
3. Banner de arranque, regla separadora, alineación clave-valor, marcadores por tipo.
4. Banner + resumen en `dialog/console.go` (el prompt de selección de servidor).
5. Flag `--output` donde se puede registrar **localmente** (§5.3) + variable de entorno.
6. `chcp 65001` en `common/resources/start.bat` (§6.9).
7. Tests de paridad, de catálogo de glifos y de detección.

### 5.2 Fuera de alcance (decisiones explícitas)

| Fuera | Motivo |
|---|---|
| `launcher-agent`, `launcher-config-admin-agent` | Requisito del usuario. Además son procesos sin consola propia en el flujo normal (`agent` corre como watcher; `config-admin-agent` detrás de un named pipe), así que estilizarlos no aportaría nada visible |
| `server`, `battle-server-manager`, `battle-server-broadcast`, `server-genCert`, `config-helper`, `tools/*` | Fuera de alcance |
| `common/**` | Compartido con los anteriores. Se usan sus puntos de extensión (`Initialize(io.Writer)`, `Interactive()`), no se modifica |
| Los **logs de fichero** (`logs/.../*.txt`) | Deben seguir siendo ASCII plano y grepeables. `ANSIColorScheme`/glifos nunca se escriben ahí |
| `WriteFileLog` (`launcher/internal/cmdUtils/logger/log.go:81-130`) y los separadores `==========` | Son salida de diagnóstico de fichero, no de consola |
| Spinners, barras de progreso, redraw con `\r`, alt-screen | Ver §8. Requiere Bubble Tea, y el veredicto es no (de momento) |
| Reescritura de los textos de los mensajes | Solo con el test de paridad de §7.2 y revisión uno a uno. Fase 3 |
| Cambios en la lógica de `dialog` (selección, confirmación) | Solo presentación |

### 5.3 El flag `--output` y por qué no va en `common`

`--help`/`--version` se registran en `common/cmd/flagSet.go:82-86` (`addDefaultFlags`), función
compartida por `server`, `battle-server-manager`, `config` y `config-admin`. Añadir `--output`
ahí tocaría `common` y contaminaría los binarios fuera de alcance.

Solución sin tocar `common`:

- **Variable de entorno** `AGE_LANSERVER_OUTPUT=auto|color|ascii` — funciona hoy en los 4 módulos,
  precedencia más alta salvo override explícito.
- **Flag local**: `launcher` usa `SingleFlagSet` y expone `s.Fs()`
  (`common/cmd/flagSet.go:108-110`), así que `launcher/internal/cmd/root.go:140` puede registrar
  `--output` directamente. Para `config`/`config-admin` (`RootFlagSet`) **no** hay
  `Fs()` expuesto → se resuelve solo por entorno en la v1. Dejar anotado como deuda (§12).

---

## 6. Diseño

### 6.1 El modelo de capacidad: tres ejes, no un interruptor

El error de diseño más fácil de cometer es tratar "bonito" como un booleano. No lo es. Hay
**tres ejes ortogonales** y hay combinaciones reales que solo existen si se separan:

- Win10 con `conhost` + code page 437 → **color sí, Unicode no**.
- Linux en `LANG=C` → **color sí, Unicode no** (y emoji tampoco).
- Terminal moderno → color + Unicode + emoji.
- `cmd.exe` Win7 → nada.

```go
package ui

// GlyphTier indica el glifo más complejo que la terminal puede mostrar sin
// degradarse a caracteres de reemplazo.
type GlyphTier uint8

const (
	// TierASCII: sólo caracteres ASCII imprimibles (0x20-0x7E).
	// Único tier garantizado en cmd.exe de Windows 7.
	TierASCII GlyphTier = iota
	// TierUnicode: añade box-drawing, flechas, geométricos y dingbats BMP.
	// Requiere code page UTF-8 en Windows y fuente con esos bloques.
	TierUnicode
	// TierEmoji: añade emoji de doble ancho.
	// Requiere lo anterior + fuente con emoji o terminal conocido-bueno.
	TierEmoji
)

// Capability es el resultado inmutable de la detección.
type Capability struct {
	// Profile decide el downsampling de color (NoTTY, ASCII, ANSI, ANSI256, TrueColor).
	Profile colorprofile.Profile
	// Tier decide el glifo máximo.
	Tier GlyphTier
	// Columns es el ancho de la consola, o 0 si no se pudo determinar.
	Columns int
}

// ColorOk indica que se pueden emitir secuencias SGR con seguridad.
func (c Capability) ColorOk() bool { return c.Profile >= colorprofile.ANSI }

// ColumnLimit devuelve el ancho a ajustar, o 0 si no se debe ajustar.
func (c Capability) ColumnLimit() int {
	if c.Columns <= 0 {
		return 0
	}
	return c.Columns
}
```

`Profile` y `Tier` se calculan por separado a propósito: `Profile` lo decide casi todo
`colorprofile` (y lo hace bien, incluido Win7 → `NoTTY`), `Tier` es 100% nuestro porque
`colorprofile` no tiene concepto de glifo.

### 6.2 Detección

```go
// probe agrupa las dependencias del sistema para que los tests no toquen la consola real.
type probe struct {
	isTerminal  func() bool
	vtUsable    func() bool   // windows: ENABLE_VIRTUAL_TERMINAL_PROCESSING verificado
	outputCode  func() uint32 // windows: GetConsoleOutputCP
	columns     func() int    // windows: GetConsoleScreenBufferInfo
	localeUTF8  func() bool   // unix: LANG/LC_ALL/LC_CTYPE
	osVersion   func() (major, build uint32)
}

// Detect es la única función que produce una Capability.
func Detect(w io.Writer, env []string, p probe) Capability {
	force, forced := forcedTier(env)
	if forced && force == ForceASCII {
		return Capability{Profile: colorprofile.NoTTY, Tier: TierASCII}
	}
	// colorprofile se ocupa del color (NO_COLOR/CLICOLOR/TERM/dumb/terminfo/Windows).
	profile := colorprofile.Detect(w, env)
	if noColorPresent(env) { // H2: presencia, no ParseBool
		profile = colorprofile.NoTTY
	}
	if p.isTerminal() && profile >= colorprofile.ANSI && !p.vtUsable() {
		// H1: colorprofile cree que hay color pero la consola no lo soporta.
		profile = colorprofile.NoTTY
	}
	...
}
```

Orden de decisión (fase 1: `Tier`):

1. `AGE_LANSERVER_OUTPUT=ascii` → `TierASCII`, `Profile = NoTTY`. Absoluto.
2. `AGE_LANSERVER_OUTPUT=color` → `Tier` según el resto de reglas, pero `Profile` nunca `NoTTY`
   (aun así: si la VT no es usable, se **degrada a `TierASCII`**, no se fuerza; documentado).
3. No TTY en stdout (redirección, pipe, `go test`) → `TierASCII`, `Profile = NoTTY`.
   **Esta regla es la que salva los tests golden de §2.6.**
4. `TERM=dumb` → `TierASCII`.
5. `NO_COLOR` presente → `Profile = NoTTY` (pero el tier no baja: sin color aún se pueden usar
   signos Unicode, igual que `NO_COLOR` permite texto en negrita).
6. **Unicode** (`Tier >= TierUnicode`) si:
   - Unix: el locale es UTF-8 (`LANG`/`LC_ALL`/`LC_CTYPE` contienen `utf`/`UTF-8`), o `TERM`
     está en la lista de terminales conocidos-buenos (`xterm*`, `alacritty`, `kitty`,
     `wezterm`, `foot`, `ghostty`, `rio`, `vte`, `konsole`, `alacritty`, `st`), o
     `COLORTERM`/`WT_SESSION` presentes. Si el locale es `C`/`POSIX`/ausente → `TierASCII`.
   - Windows: `osVersion().major >= 10 && build >= 10586` **y** `outputCode() == 65001`
     **y** `vtUsable()`. Los tres, no uno: así un `chcp 65001` en Win7 **no** sube el tier
     (buena noticia), y un `cmd` de Win10 sin `chcp` tampoco.
7. **Emoji** (`Tier == TierEmoji`) si `TierUnicode` **y** una de estas:
   - Unix: `TERM` en la lista de terminales con emojigarantizado (kitty, alacritty, wezterm,
     foot, ghostty, rio, konsole, xterm>=2020) **o** locale UTF-8 **y** ancho >= 60.
   - Windows: `WT_SESSION` presente (Windows Terminal), o `ConEmuANSI=ON`.
     Un `conhost` plano con 65001 se queda en `TierUnicode` porque la fuente raster del
     sistema no tiene emoji (esto es exactamente el caso de Win10 en un equipo sin Windows
     Terminal, que es donde `✅` se vería como caja).

`vtUsable` en Windows (`ui/vt_windows.go`):

```go
var (
	modkernel32                     = windows.NewLazySystemDLL("kernel32.dll")
	procGetConsoleOutputCP          = modkernel32.NewProc("GetConsoleOutputCP")
)

func vtUsable() bool {
	// Emuladores que traen su propio intérprete VT y son de fiar.
	if os.Getenv("WT_SESSION") != "" || os.Getenv("ConEmuANSI") == "ON" {
		return true
	}
	if major, _, build := windows.RtlGetNtVersionNumbers(); major < 10 || build < 10586 {
		return false // H1: antes de 10586 el bit no existe; no se intenta siquiera
	}
	h, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil || h == 0 {
		return false
	}
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return false
	}
	const enableVT = windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	if mode&enableVT != 0 {
		return true
	}
	// Probamos a activarlo y **verificamos** con una segunda lectura: en una consola
	// que no soporta el bit, SetConsoleMode puede no fallar y simplemente no hacer nada.
	if err := windows.SetConsoleMode(h, mode|enableVT); err != nil {
		return false
	}
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return false
	}
	return mode&enableVT != 0
}
```

Nota de efectos secundarios: sólo se hace `OR` del bit VT sobre el modo de la consola de
**salida**, nunca se borra ningún otro bit, y sólo si no estaba ya activo. Es lo mismo que hace
`cmd.exe` en Win10. Para quien prefiera no tocar nada: `AGE_LANSERVER_OUTPUT=ascii` (o
`TERM=dumb`, o redirigir) lo evita.

`outputCode` y `columns`:

```go
func outputCode() uint32 {
	r, _, _ := procGetConsoleOutputCP.Call()
	return uint32(r)
}

func consoleColumns(fallback int) int {
	h, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil || h == 0 {
		return fallback
	}
	var info consoleScreenBufferInfo
	if err := getConsoleScreenBufferInfo(h, &info); err != nil {
		return fallback
	}
	if w := int(info.Size.X); w > 0 {
		return w
	}
	return fallback
}
```

Ancho: si no es TTY o falla la ioctl → `Columns = 0` → **no se ajusta nada**, que es lo
deseable cuando la salida va a un fichero o a un pipe (no meter saltos de línea artificiales
en los logs).

### 6.3 Árbol del paquete `launcher-common/ui`

```
launcher-common/ui/
    capability.go    Capability, GlyphTier, Detect, probe
    capability_unix.go   stub de build para !windows
    vt_windows.go    vtUsable, outputCode, consoleColumns
    vt_other.go      stubs (false, 0, 0)
    glyph.go         Glyph, catálogo, G(name)
    theme.go         estilos lipgloss resueltos desde Capability
    print.go         Println/Printf/Step/Detail/Ok/Fail/Warn/Info/Section/Rule/KV/Banner
    width.go         envoltura de línea a ColumnLimit, sangría
    sanitize.go      StripANSI (tests + sink de fichero)
    *_test.go
```

Restricción de dependencias (igual que en `ZENITY.md` §4.1): el paquete **no** importa
`launcher`, ni `common/logger`, ni nada de `launcher/internal`. Solo `charm.land/lipgloss/v2`,
`github.com/charmbracelet/colorprofile`, `github.com/charmbracelet/x/ansi` y `golang.org/x/sys`.

### 6.4 Catálogo de glifos

```go
type Glyph struct {
	// ASCII es un token puramente ASCII y sin caracteres de control.
	// Puede tener más de una celda (p.ej. "[ ok ]"); por eso NUNCA se usa
	// dentro de tablas o columnas alineadas.
	ASCII string
	// Unicode es un carácter BMP de una celda presente en CP437/CP850 o en las
	// fuentes de consola habituales.
	Unicode string
	// Emoji es un emoji de doble ancho. Sólo en marcadores de línea.
	Emoji string
}
```

| Nombre | ASCII | Unicode | Emoji | Uso |
|---|---|---|---|---|
| `Success` | `[ ok ]` | `✔` | `✅` | operación completada |
| `Failure` | `[fail]` | `✖` | `❌` | operación fallida |
| `Warn` | `[warn]` | `⚠` | `⚠️` | advertencia (no bloqueante) |
| `Info` | `[info]` | `›` | `ℹ️` | paso en curso / nota |
| `Question` | `[ ? ]` | `?` | `❓` | prompt interactivo |
| `Bullet` | `-` | `•` | *(nunca)* | elemento de lista |
| `Arrow` | `->` | `→` | *(nunca)* | procedencia, "de X a Y" |
| `Branch` | `+-` | `├─` | *(nunca)* | árbol / sangría |
| `Last` | ``` `- ``` | `└─` | *(nunca)* | último elemento de árbol |
| `Rule` | `-` | `─` | *(nunca)* | separador horizontal |
| `CornerTL` | `+` | `┌` | *(nunca)* | marco del banner |
| `CornerTR` | `+` | `┐` | *(nunca)* | " |
| `CornerBL` | `+` | `└` | *(nunca)* | " |
| `CornerBR` | `+` | `┘` | *(nunca)* | " |
| `Rocket` | `*` | *(nunca)* | `🚀` | cabecera del launcher |

Reglas duras del catálogo:

1. **ASCII siempre puro ASCII**: test que rechaza cualquier byte `>= 0x7F` o `< 0x20`
   (salvo el espacio) en el campo `ASCII`.
2. **Unicode de una celda**: test con `runewidth.StringWidth(g.Unicode) == 1`. Por eso se
   descartan `⚠️` (con VS16), `☕`-con-VS16 y cualquier ZWJ sequence en la columna Unicode.
3. **Emoji fuera de tablas**: los marcadores de línea no se alinean en columna, así que su
   ancho variable no rompe nada. Dentro de tablas, listas alineadas, rutas y mensajes de error
   solo se usan `Bullet`, `Arrow`, `Branch`, `Last`.
4. **El nombre del glifo es lo único que aparece en el código.** Prohibido escribir `✔` o `✅`
   o `[ ok ]` literal en un mensaje. Un test de `grep` sobre los 4 módulos lo hace cumplir.

```go
// G devuelve el glifo efectivo para el tier actual.
func G(name Name) string {
	g := glyphs[name]
	switch Current().Tier {
	case TierEmoji:
		if g.Emoji != "" {
			return g.Emoji
		}
		fallthrough
	case TierUnicode:
		if g.Unicode != "" {
			return g.Unicode
		}
	}
	return g.ASCII
}
```

### 6.5 Tema `lipgloss`

```go
package ui

// styles son estilos ya resueltos para la Capability actual. En TierASCII /
// Profile <= ASCII todos son el estilo cero, así que Render es la identidad:
// la degradación es estructural, no heurística.
var (
	stTitle   lipgloss.Style
	stVersion lipgloss.Style
	stOK      lipgloss.Style
	stFail    lipgloss.Style
	stWarn    lipgloss.Style
	stInfo    lipgloss.Style
	stDim     lipgloss.Style
	stKey     lipgloss.Style
	stRule    lipgloss.Style
	stBanner  lipgloss.Style
)

// base evita que lipgloss convierta los \t del código existente (lipgloss los
// pasa a 4 espacios al renderizar). NoTabConversion deja los bytes intactos.
func base() lipgloss.Style {
	return lipgloss.NewStyle().TabWidth(lipgloss.NoTabConversion)
}

func initTheme(c Capability) {
	if c.Profile <= colorprofile.ASCII {
		stTitle, stVersion, stOK, stFail, stWarn = base(), base(), base(), base(), base()
		stInfo, stDim, stKey, stRule, stBanner = base(), base(), base(), base(), base()
		return
	}
	// Complete traduce un color "ideal" al perfil real de la terminal.
	// Con NoTTY devuelve colorprofile.NoColor, o sea texto plano.
	complete := lipgloss.Complete(c.Profile)
	accent := complete(lipgloss.Color("14"), lipgloss.Color("39"), lipgloss.Color("#7D56F4"))
	ok := complete(lipgloss.Color("10"), lipgloss.Color("42"), lipgloss.Color("#04B575"))
	warn := complete(lipgloss.Color("11"), lipgloss.Color("214"), lipgloss.Color("#FFB454"))
	fail := complete(lipgloss.Color("9"), lipgloss.Color("196"), lipgloss.Color("#EB4268"))
	mute := complete(lipgloss.Color("8"), lipgloss.Color("243"), lipgloss.Color("#6C6C6C"))

	stTitle = base().Foreground(accent).Bold(true)
	...
	stDim = base().Foreground(mute)
	stRule = base().Foreground(mute)
}
```

Puntos deliberados:

- **No** se usa `lipgloss.HasDarkBackground` (§3.3): hace una consulta OCR al terminal y en
  consoals antiguas es un riesgo de bloqueo o de ruido. Los colores se eligen sin preguntar.
- Los estilos se resuelven **una vez** en `Initialize`, no por línea: el coste por línea es
  un `Render` de lipgloss, que es puro y barato, pero la decisión se toma una vez.
- Se **no** usa `colorprofile.Writer` en absoluto (§2.5). `colorprofile` se usa solo como
  `Profile` + `Complete`.

### 6.6 API de impresión y relación con los sinks existentes

```go
package ui

// Initialize fija el destino de consola y calcula la Capability. Se llama una
// vez desde main(), antes de commonLogger.Initialize.
func Initialize(w io.Writer, env []string)

// Current devuelve la Capability resuelta (útil para tests y para el diálogo).
func Current() Capability

// SetOutput cambia el destino (tests). Devuelve la función de restauración,
// siguiendo la convención del repo (common.SetTerminal, dialog.SetOutput).
func SetOutput(w io.Writer) (restore func())

func Printf(format string, a ...any)
func Println(a ...any)

// Step es una acción en curso. Detail es una línea subordinada a la anterior.
func Step(format string, a ...any)
func Detail(format string, a ...any)

// Marcadores. Todos aceptan el MISMO texto que hoy; sólo añaden prefijo.
func Ok(format string, a ...any)
func Fail(format string, a ...any)
func Warn(format string, a ...any)
func Info(format string, a ...any)

// Section imprime un encabezado de fase.
func Section(title string)

// Rule imprime un separador del ancho de la consola.
func Rule()

// KV alinea un par clave/valor en la columna de claves.
func KV(indent int, key string, value string)

// Banner imprime la cabecera del programa. version puede ser "".
func Banner(program string, version string)
```

Y el punto de integración por módulo, que es donde está la mínima magia:

**`launcher`** — `launcher/internal/cmdUtils/logger/log.go:143-151`. El log de fichero recibe el
texto **sin cambios**; solo se estila la mitad de stdout:

```go
func Printf(format string, a ...any) {
	commonLogger.PrefixPrintf("main", format, a...) // log de fichero: intacto
	ui.Printf(format, a...)                          // consola: estilada
}

func Println(a ...any) {
	commonLogger.PrefixPrintln("main", a...) // log de fichero: intacto
	ui.Println(a...)                         // consola: estilada
}
```

Se añaden, en el mismo fichero, envolturas de una línea que delegan en `ui`:

```go
func Ok(format string, a ...any)    { commonLogger.PrefixPrintf("main", format+"\n", a...); ui.Ok(format, a...) }
func Fail(format string, a ...any)  { commonLogger.PrefixPrintf("main", format+"\n", a...); ui.Fail(format, a...) }
func Warn(format string, a ...any)  { commonLogger.PrefixPrintf("main", format+"\n", a...); ui.Warn(format, a...) }
func Info(format string, a ...any)  { commonLogger.PrefixPrintf("main", format+"\n", a...); ui.Info(format, a...) }
func Detail(format string, a ...any) { commonLogger.PrefixPrintf("main", format+"\n", a...); ui.Detail(format, a...) }
```

(Para no duplicar, por detrás hay un helper privado `dup(format, a...)` que escribe en
`commonLogger` el texto plano y devuelve el `format` con `\n`, para que el fichero y la consola
reciban exactamente el mismo texto.)

**`config` y `config-admin`** — aquí `commonLogger` *es* el printer de stdout
(`launcher-config/main.go:16`, `launcher-config-admin/main.go:17`) y `Initialize` **no se toca**.
El cambio es mecánico en los puntos de salida:

```go
// antes
commonLogger.Println("Successfully added user certificate")
// después
commonLogger.Println(commonUi.Ok("Successfully added user certificate"))

// antes
commonLogger.Println("Failed to back up metadata")
// después
commonLogger.Println(commonUi.Fail("Failed to back up metadata"))
```

Consecuencias correctas por construcción:
- `Initialize(os.Stdout)` sigue recibiendo `os.Stdout` → los flags no cambian → no aparece
  `|MAIN|` ni timestamp en una terminal interactiva.
- Los ficheros de log de estos dos binarios se escriben por `internal.Logger.Buffer(...)` con
  cadenas propias; no pasan por aquí y quedan ASCII planos.

`commonLogger.Initialize(nil)` en `launcher/main.go:15` y `common.ChdirToExe()` se quedan igual.

**`launcher-common`** — `launcher-common/configRevert.go` usa `commonLogger` para mensajes que
en el flujo normal van al **log** (el `config` hijo los redirige a un buffer:
`launcher/internal/cmd/root.go:569-572` → `RunRevert(..., out, ...)` →
`configRevert.go:219-221` pone `options.Stdout/Stderr = out`). Se deja **sin estilar**, porque
su destino real es un fichero de log. Los dos mensajes de `Reverting configuration…` y
`\t'config-admin-agent' process is still executing…` los emite el proceso hijo, no el launcher,
así que estilizarlos aquí no aportaría nada visible.

→ **Decisión: `launcher-common/configRevert.go` no se toca.** Se anota en §12.

### 6.7 Sangría y ancho

Regla: **la sangría la pone `ui`, no el mensaje**. Los `\t` y `\t\t` de §2.2 se sustituyen por
llamadas con nivel explícito. Ejemplo, `launcher-config/internal/userData/backup.go:34`:

```go
// antes
commonLogger.Printf("\tSwitching %s <-> %s\n", currentPath, backupPath)
// después
commonLogger.Println(commonUi.Detail("Switching %s <-> %s", currentPath, backupPath))
```

`ui.Detail` aplica sangría de nivel 2 y, si `ColumnLimit() > 0`, envuelve con
`lipgloss.Wrap(text, limit-indent, " ")` y vuelve a anteponer la sangría en las líneas de
continuación. Con `Columns == 0` (pipe, fichero) **no envuelve**, para no meter saltos de línea
artificiales en la salida redirectionada.

Si un mensaje se pasa tal cual a `commonLogger` (sin `ui`), los `\t` se conservan: por eso
`base()` lleva `TabWidth(lipgloss.NoTabConversion)` (§6.5), aunque el caso sólo aparezca si
alguien estila una cadena con tabuladores.

### 6.8 Rediseño del output

#### 6.8.1 `launcher` — referencia de diseño

Hoy, en una terminal interactiva (`|MAIN|` oculto, líneas peladas):

```
Using main config file: C:\Program Files\AgeLANServer\config.toml
Using game config file: C:\Program Files\AgeLANServer\config.aos4.toml
Dialog backend: console.
Game aos4.
Running as administrator, this is not recommended for security reasons. It will request isolated admin privileges if/when needed.
Looking for the game...
Game found on C:\Program Files (x86)\Steam\steam.exe.
'config-admin-agent' from a previous run is still active, stopping it...
'config-admin-agent' stopped.
Cleaning up (if needed)...
Setting up...
Found 'server' executable path: C:\Program Files\AgeLANServer\server\server.exe
Successfully added host mappings
Successfully trusted certificate
```

**Propuesta `TierASCII`** (lo que verá un `cmd.exe` de Win7; mismas frases, marcador ASCII,
alineación clave-valor):

```
+-------------------------------------+
|          AGE LAN SERVER             |
|          v1.4.0                     |
+-------------------------------------+

  main config file    config.toml
  game config file    config.aos4.toml
  dialog backend      console
  game                aos4

  [warn] Running as administrator, this is not recommended for security
         reasons. It will request isolated admin privileges if/when needed.
  [info] Looking for the game...
  [ ok ] Game found on steam.exe
  [info] 'config-admin-agent' from a previous run is still active,
         stopping it...
  [ ok ] 'config-admin-agent' stopped
  [info] Cleaning up (if needed)...
  [info] Setting up...
  [ ok ] Found 'server' executable path: server\server.exe
  [ ok ] Successfully added host mappings
  [ ok ] Successfully trusted certificate
```

**Propuesta `TierUnicode`** (Win10 con `chcp 65001`, o Linux con locale UTF-8):

```
  ▌AGE LAN SERVER▐  v1.4.0
  ──────────────────────────────────────
  main config file    config.toml
  game config file    config.aos4.toml
  dialog backend      console
  game                aos4

  ⚠ Running as administrator, this is not recommended for security
    reasons. It will request isolated admin privileges if/when needed.
  › Looking for the game...
  ✔ Game found on steam.exe
  › Cleaning up (if needed)...
  ✔ 'config-admin-agent' stopped
  › Setting up...
  ✔ Successfully added host mappings
  ✔ Successfully trusted certificate
```

**Propuesta `TierEmoji`** (Windows Terminal, kitty, alacritty…):

```
  🚀 AGE LAN SERVER  v1.4.0
  ──────────────────────────────────────
  …
  ✅ Game found on steam.exe
  ⚠️  Running as administrator, …
```

Notas de diseño de la maqueta:

- El **banner** es una caja (3 líneas + 2 de marco = 5) y **no** arte figlet de 7–8 líneas: en
  una consola 80×25 el arte grande se come un tercio del buffer visible y desplaza el log
  justo cuando el usuario va a necesitar leer un error. Si se quiere arte grande, se ofrece
  aparte y sólo con `Columns >= 100` y tier >= Unicode, generado con
  `figlet -f slant -w 60 "AgeLANServer"` — no se escribe a mano en el repo para evitar que
  quede mal alineado.
- El banner se imprime **una vez**, al principio, y sólo si la salida es un TTY. Si `Columns == 0`
  (pipe a fichero) no se imprime: un banner en un log es ruido.
- El marcador va **antes** del texto y la continuación se indenta con `markerWidth`, de modo que
  un mensaje largo sigue siendo legible como bloque.
- `Dialog backend: console.` → `dialog backend  console`. Es un cambio de *separador*, no de
  información, permitido por el contrato de paridad de §7. La frase conserva las dos palabras
  con token (`dialog`, `backend`, `console`) presentes en el original.

#### 6.8.2 `launcher-config setup` — el flujo más largo

Hoy:
```
Setting up configuration for aos4...
Adding user certificate, authorize it if needed...
Successfully added user certificate
Backing up metadata
Successfully backed up metadata
Failed to back up metadata
Error message: <...>
```

`TierASCII`:
```
[info] Setting up configuration for aos4
[info] Adding user certificate, authorize it if needed
[ ok ] Successfully added user certificate
[info] Backing up metadata
[fail] Failed to back up metadata
       Error message: <...>
```

Aquí se ve el valor de `Detail`: el `Error message:` pasa a ser subordinado del fallo en vez de
una línea suelta al mismo nivel. En `TierASCII` es **exactamente** la misma información con la
misma frase.

#### 6.8.3 Tablas: sólo cuando son realmente tabulares

`printCandidates` (`launcher/internal/dialog/console.go:49-55`) y el listado de battle servers
sí son tabulares. Requisitos:

- `TierASCII` → **numeración plana, sin bordes**, idéntica a hoy. Esto es lo que mantiene
  vivos `console_test.go:152` y `select_server_test.go:225` sin tocarlos.
- `TierUnicode`/`TierEmoji` → `lipgloss/table` con `BorderStyle(lipgloss.RoundedBorder())` y
  celdas sin emoji (regla 3 de §6.4).
- `ASCIIBorder()` existe pero **no se usa**: `+---+` en una fuente raster de 80 columnas queda
  desalineado con cualquier fuente proporcional y ocupa 4 líneas por fila.

#### 6.8.4 `dialog/console.go`

```
┌ Found the following 'server's ────────────────┐
│  1  192.168.1.50:27015   aos4   (8 ms)        │
│  2  192.168.1.51:27015   aos4  (31 ms)        │
└───────────────────────────────────────────────┘
  [ ? ] Enter the number of the 'server' (1-2):
```

En `TierASCII` es exactamente la lista numerada actual + el prompt actual con `[ ? ]` delante.
La lógica de reintento (`console.go:16-31`) no se toca: el test
`TestConsoleConfirmStartServerEOFReturnsTrue` depende de que un EOF no aborte el arranque.

### 6.9 `chcp 65001` en `common/resources/start.bat`

`common/resources/start.bat` es hoy la entrada por doble clic:

```bat
@echo off
cd /d "%~dp0"
%*
```

Recomendación (una línea, inocua si falla):

```bat
@echo off
chcp 65001 >nul 2>&1
cd /d "%~dp0"
%*
```

Por qué es seguro: la regla de §6.2 paso 6 exige **build >= 10586 Y codepage 65001**, así que
un `chcp` en Win7 no sube el tier y el output sigue siendo ASCII. En Win10 sube de
`TierASCII` a `TierUnicode` sin tocar nada más. `common/resources/start.sh` no se toca.

---

## 7. Contrato de paridad: la garantía para terminales antiguas

### 7.1 Enunciado

> Para todo mensaje `m` emitido por `launcher`, `config` y `config-admin`:
> en `TierASCII` la salida satisface:
> **(P1)** no contiene ningún byte fuera de `0x09` y `0x0A`–`0x7E`;
> **(P2)** no contiene ninguna secuencia ANSI (`ESC` / `CSI`);
> **(P3)** el multiconjunto de tokens de información de `m` es un subconjunto del multiconjunto
> de tokens de la línea ASCII emitida;
> **(P4)** el orden relativo de esos tokens se conserva;
> **(P5)** el número de líneas no disminuye respecto al original (no se puede "ganar"
> información a costa de la misma: un mensaje de 1 línea en Win7 no puede desaparecer);
> **(P6)** el código de salida del proceso no cambia.

### 7.2 Cómo se verifica mecánicamente

`ui/parity_test.go` mantiene un **catálogo de mensajes** (los ~40 textos canónicos de
`launcher/internal/cmd/root.go`, `launcher-config/internal/cmd/setUp.go` y
`launcher-config-admin/internal/cmd/setUp.go`) y para cada par `(original, restyled)`:

```go
// assertParity exige P3 y P4.
func assertParity(t *testing.T, original, restyled string) {
	t.Helper()
	// El texto base del tier ASCII debe ser recuperable sin ambigüedad.
	plain := stripDecorators(restyled)
	origTokens := tokens(original)
	newTokens := tokens(plain)
	for tok, n := range origTokens {
		if newTokens[tok] < n {
			t.Fatalf("paridad rota: falta %q x%d en\n original: %q\n restyled: %q", tok, n, original, plain)
		}
	}
	// Ni un token nuevo sin revisar.
	for tok, n := range newTokens {
		if _, ok := origTokens[tok]; !ok && !decoratorTokens[tok] {
			t.Fatalf("paridad: token nuevo %q x%d en %q", tok, n, plain)
		}
	}
}
```

Y `capability_test.go` para P1/P2: con la `probe` de `TierASCII`, recorrer el catálogo completo y
comprobar `utf8.Valid`, `allASCII(s)`, `!strings.Contains(s, "\x1b")` y
`lipgloss.Width(line) <= columns`.

`stripDecorators` quita el marcador (`[ ok ]`, `[fail]`, `[warn]`, `[info]`, `[ ? ]`) y la
sangría; `decoratorTokens` es la lista cerrada de tokens de decoración permitidos
(`ok`, `fail`, `warn`, `info`, `?`). Un token nuevo obliga a revisarlo explícitamente.

### 7.3 Pruebas de que Win7 cae en `TierASCII`

`capability_test.go` con la `probe` inyectada:

| Caso | Entorno / probe | `Tier` esperado | `Profile` esperado |
|---|---|---|---|
| Win7 `cmd.exe` | `probe.vtUsable=false`, `osVersion=(6,1)`, sin `WT_SESSION` | `TierASCII` | `NoTTY` |
| Win7 + `ANSICON=1` | `vtUsable=true`, `osVersion=(6,1)` | `TierASCII` (el code page no es 65001) | `ANSI` |
| Win7 + `chcp 65001` | `outputCode=65001`, `osVersion=(6,1)` | `TierASCII` (build < 10586) | `NoTTY` |
| Win10 `conhost` sin VT | `osVersion=(10,17763)`, `vtUsable=false` | `TierASCII` | `NoTTY` ← **caso H1** |
| Win10 `conhost` + VT, cp 437 | `vtUsable=true`, `outputCode=437` | `TierASCII` | `ANSI256`/`TrueColor` |
| Win10 + VT + cp 65001 | `osVersion=(10,19045)`, `vtUsable=true`, `outputCode=65001` | `TierUnicode` | `TrueColor` |
| Windows Terminal | `WT_SESSION` presente | `TierEmoji` | `TrueColor` |
| Linux `LANG=C` | `TERM=xterm-256color` | `TierASCII` | `ANSI256` |
| Linux `LANG=es_ES.UTF-8` | `TERM=xterm-256color` | `TierEmoji` | `ANSI256` |
| Pipe (test de Go) | `probe.isTerminal=false` | `TierASCII` | `NoTTY` ← **salva §2.6** |
| `NO_COLOR=` (vacío) | presente | `Tier` según el resto, `Profile=NoTTY` | `NoTTY` ← **caso H2** |
| `AGE_LANSERVER_OUTPUT=ascii` | lo que sea | `TierASCII` | `NoTTY` |

---

## 8. Evaluación de Bubble Tea, Bubbles y el resto de Charm

### 8.1 API verificada de Bubble Tea v2 (módulo `charm.land/bubbletea/v2`)

- `tea.NewProgram(model Model, opts ...ProgramOption) *Program`.
- `Model` = `Init() tea.Cmd`, `Update(tea.Msg) (tea.Model, tea.Cmd)`, **`View() tea.View`**
  (en v2 `View` devuelve un `tea.View`, no un `string`).
- `tea.View` es un struct con `Content string`, `AltScreen bool`, `ReportAlternateKeys bool`,
  `OnMouse`. El alt-screen **deja de ser un `ProgramOption`** (`WithAltScreen` ya no está en
  `options.go`) y pasa a ser un campo por frame.
- Opciones relevantes: `WithOutput(io.Writer)`, `WithInput(io.Reader)`, `WithContext`,
  `WithFPS(int)`, `WithColorProfile(colorprofile.Profile)`, `WithWindowSize(w, h)`,
  `WithFilter`, `WithoutSignalHandler()`, `WithoutSignals()`, `WithoutRenderer()`,
  `WithoutCatchPanics()`.
- Impresión sobre el TUI: `(*Program).Println` / `(*Program).Printf`.

Bubbles v2 (`charm.land/bubbles/v2`): `spinner`, `textinput`, `textarea`, `table`, `progress`,
`paginator`, `viewport`, `list`, `filepicker`, `timer`, `stopwatch`, `help`, `key`.

### 8.2 Dónde sí aportaría valor

| Caso | Componente | Valor real |
|---|---|---|
| Selector de `'server'` (`dialog/console.go`) | `bubbles/list` + `viewport` | Hoy es un `for` con `fmt.Scanf` de un entero (`console.go:13-32`): sin filtros, sin flechas, sin resaltado, sin scroll, y un `EOF` obliga a abortar. Una `list` con filtro por subcadena y `Viewport` para descripciones largas es una mejora UX grande y **encaja en el alcance existente** (`dialog` ya es un `Dialog` intercambiable con `zenity`: `launcher/ZENITY.md`) |
| Progreso de la fase de setup | `bubbles/spinner` + `bubbles/progress` | `config setup` puede tardar: generar certificados, escribir el hosts, `FlushDns`. Una línea viva con spinner ahorra la sensación de "colgado" |
| Pie de atajos | `bubbles/help` + `bubbles/key` | Coherencia si se adopta la `list` |

### 8.3 Por qué **no** en el flujo principal (razones duras)

| # | Razón | Detalle |
|---|---|---|
| R1 | **Colisión con el modelo de salida actual** | Bubble Tea toma el control de stdin/stdout en modo raw y redibuja la pantalla. El output del launcher es un **log lineal** que además se duplica a un fichero vía `log.Logger`. No es un modelo de "frame", es un modelo de "línea" |
| R2 | **El contrato de cleanup lo rompe** | `launcher/internal/cmd/root.go:254-278`: un `teardown` con mutex, un `defer` de panic-recovery y un handler de SIGINT que hace `os.Exit` desde **otra goroutine** (`:524-530`). Con Bubble Tea, un `os.Exit` a mitad de un frame deja el terminal en modo raw / alt-screen, y este repo ya ha 수입ido varios bugs de "kill what we just started" y "publish the pipe before its slow work" (§ commits recientes). Añadir un `os.Exit` remoto a un TUI es exactamente el tipo de regresión que este código evita por diseño |
| R3 | **El launcher se ejecuta a menudo sin TTY** | `launcher-common/configRevert.go:219-221` redirige stdout/stderr del `config` hijo a un buffer. Y `battleServerManager`/`server` se lanzan con `exec.Options`. Una UI TUI en un proceso sin consola es absurda |
| R4 | **`signal.Notify` propio** | El launcher ya gestiona `SIGINT`/`SIGTERM`. Bubble Tea los captura; habría que usar `tea.WithoutSignalHandler()`, que es una opción que existe precisamente porque esto es un problema conocido |
| R5 | **Coste de binario desproporcionado** | `lipgloss` ya añade ~2 deps directas. `bubbletea`+`bubbles` arrastran `ultraviolet`, `cursed_renderer`, `x/exp/*`, `x/input`, `x/term`, `x/ansi`, `colorprofile`… Para un programa cuyo valor es "funciona", no es justificable |
| R6 | **Ya hay una GUI** | `zenity` (`launcher/internal/dialog/zenity.go`) cubre el caso interactivo gráfico. El fallback de consola existe para cuando no hay GUI, y en ese contexto una TUI peor que líneas simples no es una mejora |

### 8.4 Veredicto

| Ámbito | Decisión |
|---|---|
| `lipgloss` (+ `colorprofile`, `x/ansi`, `x/term`, subpaquetes `table`/`list`) | **Sí, en el flujo principal.** Es el 90% del valor con el 10% del riesgo |
| `bubbletea` + `bubbles/list` en `dialog` | **Aplazado y aislado.** Si algún día se hace: un `dialog` nuevo (`rich`) seleccionado **solo** si `isTTY && Columns >= 60 && AGE_LANSERVER_DIALOG=rich` (opt-in, nunca `auto`), aislado en una goroutine con `defer` de restauración del terminal, `tea.WithoutSignalHandler()`, `tea.WithWindowSize(...)` y fallback inmediato a la `console` actual ante cualquier error. La `console` actual **no se toca**, así que el riesgo está contenido |
| `bubbletea` para el progreso del setup | **No.** Solución sin TUI: una línea de estado con `\r` reescrita, sólo si `isTTY && Tier > TierASCII`, y `Printf("\r%s", ...)` + `Printf("\n")` al terminar. Coste ~30 líneas, cero dependencias, cero modo raw (§8.3 R1/R2 siguen sin berlaku) |
| `charm.land/lipgloss/v2/table` | **Sí**, sólo para listados tabulares (§6.8.3) |
| `charm.land/lipgloss/v2/list` | **Sí**, candidatos |
| `charm.land/lipgloss/v2/tree` | **No.** Su lugar natural sería el volcado de hosts/certificados, que va al **log de fichero** y debe seguir plano |
| `charmbracelet/x/ansi` | **Sí, útil**: `ansi.Strip` para el test de paridad y para garantizar por construcción que el sink de fichero no recibe ANSI; `ansi.Truncate`/`ansi.Wrap` para el ancho |
| `charmbracelet/x/term` | **Sí**: tamaño de terminal y TTY coherentes con la propia lógica de Charm |
| `charmbracelet/fang` | **No.** Es un wrapper de **cobra** (`fang.Execute(ctx, *cobra.Command)`) con `--version`, manpages y completions. El repo usa `pflag` con dispatchers propios en `common/cmd/flagSet.go`, compartido con módulos fuera de alcance |
| `bubbles/{textinput,filepicker,textarea}` | **No.** No hay entrada de texto libre en ningún punto de estos cuatro módulos |
| `glamour`, `huh`, `bubbletea` TUI completa | **No** |

---

## 9. Dependencias

```
# Desde F:\ageLANServer\launcher-common
go get charm.land/lipgloss/v2@v2.0.6
go get github.com/charmbracelet/colorprofile@v0.4.3
go get github.com/charmbracelet/x/ansi@latest
```

`charm.land/lipgloss/v2@v2.0.6` y `github.com/charmbracelet/colorprofile@v0.4.3` son exactamente
las versiones ya resueltas en `go.work.sum`, así que no hay resolución nueva que pueda fallar.
`x/ansi` hay que resolverlo.

```
# OJO: go mod tidy y go work sync NO son seguros en este workspace (ver launcher/ZENITY.md §7.7)
#   - `go mod tidy` intenta resolver common y launcher-common desde el proxy y falla
#   - `go mod tidy -e` añade requirements de common/launcher-common al go.mod
#   - `go work sync` propaga eso a los 17 módulos
# Se conserva el grafo que produce `go get` y NO se editan go.mod/go.sum a mano.
```

Además, para que `go.sum` de cada módulo consumidor lleve los hashes (necesario para los builds
por módulo que hace goreleaser, que no usan el workspace):

```
# Desde F:\ageLANServer\launcher, F:\ageLANServer\launcher-config, F:\ageLANServer\launcher-config-admin
go get charm.land/lipgloss/v2@v2.0.6
```

Módulos afectados: `launcher-common`, `launcher`, `launcher-config`, `launcher-config-admin`.
Los otros 13 no se tocan.

---

## 10. Plan de pruebas

### 10.1 `launcher-common/ui/capability_test.go` (nuevo)

`probe` inyectada con `t.Run` por cada fila de la tabla de §7.3. Comprueba `Tier`, `Profile` y
`Columns`. Incluye los 4 casos de hueco: H1 (Win10 sin VT), H2 (`NO_COLOR=` vacío), Win7 con
`chcp 65001`, y pipe.

### 10.2 `launcher-common/ui/glyph_test.go` (nuevo)

Dos tests de tabla sobre el catálogo completo:
- `TestGlyphASCIIIsPureASCII`: cada `g.ASCII` cumple `0x20 <= b <= 0x7E` o `b == ' '`; ningún
  `\t`, `\n`, `\x1b`, `\u200b`.
- `TestGlyphUnicodeIsSingleWidth`: `runewidth.StringWidth(g.Unicode) == 1`.
- `TestGRespectsTier`: para cada `Tier`, `G(name)` devuelve el campo esperado.

### 10.3 `launcher-common/ui/theme_test.go` (nuevo)

- `TestThemeIsIdentityInASCII`: con `Profile=NoTTY`, para cada estilo, `s.Render("hola") == "hola"`
  y `s.Render("a\tb") == "a\tb"` (comprueba `NoTabConversion`).
- `TestThemeNoANSIWhenColorOff`: con `Profile=colorprofile.ASCII` (NO_COLOR), ningún `Render`
  emite `\x1b`; sólo texto.

### 10.4 `launcher-common/ui/print_test.go` (nuevo)

`SetOutput(&bytes.Buffer{})` y tabla de `Tier` × función. Comprueba, para `TierASCII`:
- el buffer es ASCII puro;
- cada marcador aparece exactamente una vez y al principio de la línea;
- el texto de entrada aparece íntegro;
- `KV` alinea la columna de claves en los tres tiers.

### 10.5 `launcher-common/ui/parity_test.go` (nuevo)

El catálogo de mensajes de §7.2 + `assertParity`. Es el test que **materializa el requisito del
usuario**. Debe ejecutarse también en la matriz de build de Windows.

### 10.6 `launcher-common/ui/banner_test.go` (nuevo)

- El banner cabe en `Columns` (con `Columns = 80` y con `Columns = 40`).
- En `TierASCII` sólo usa los caracteres `+`, `-`, `|`, espacio y letras.
- No se imprime cuando `Columns == 0`.

### 10.7 `launcher/internal/dialog/console_test.go` (existente, **sin cambios**)

`console_test.go:152` (igualdad byte a byte) y `select_server_test.go:225` **deben seguir
pasando sin tocarse**. Eso es el test que verifica que el tier por defecto en `go test` es
`TierASCII` (§6.2 regla 3). Si alguno falla, es que se ha introducido estilado en un camino que
no debe, o que la detección ha dejado de treatar un pipe como no-TTY.

### 10.8 `launcher/internal/cmdUtils/logger/log_test.go` (existente)

Añadir un caso que verifique el **dual sink**: con `LogEnabled=false`, el texto que llega al
buffer del log de fichero y el que llega a la consola coinciden tras `ansi.Strip`. Es la defensa
contra la fuga de ANSI al log.

### 10.9 `launcher/internal/cmd/root.go` — tests existentes

`runRoot_test.go:1733,1736` usan `Contains("Dialog backend: console.")`. Con KV el texto pasa a
`dialog backend      console` y **rompen**. Dos opciones, y hay que elegir una explícitamente:
- (a) no usar `KV` para los mensajes que hoy llevan `": "` (se conserva `Dialog backend: console.`);
- (b) actualizar los dos `Contains` y añadir el caso de paridad del mensaje.
  Recomendación: **(a)** en la v1. `KV` se reserva para mensajes **nuevos** o para mensajes que ya
  son tablas de pares. Es la opción de menor riesgo y mantiene el contrato de §7.

### 10.10 Higiene

```powershell
$env:CGO_ENABLED = "1"
go test ./launcher-common/... ./launcher/... ./launcher-config/... ./launcher-config-admin/... -race
go vet ./launcher-common/... ./launcher/... ./launcher-config/... ./launcher-config-admin/...
gofmt -l ./launcher-common/ui
```

```powershell
# Verificación manual de los tres tiers (PowerShell)
$env:AGE_LANSERVER_OUTPUT = "ascii";  .\launcher.exe --game aos4 --help
$env:AGE_LANSERVER_OUTPUT = "color";  .\launcher.exe --game aos4 --help
$env:AGE_LANSERVER_OUTPUT = "auto";   .\launcher.exe --game aos4 --help | Out-File x.txt   # pipe -> ASCII
$env:NO_COLOR = "";                  .\launcher.exe --game aos4 --help                # H2 -> sin color
.\launcher.exe --game aos4 --help                                                          # TTY -> auto
```

---

## 11. Orden de implementación

| # | Paso | Entregable | Sin él no compila |
|---|---|---|---|
| 1 | Dependencias | `go get` en `launcher-common` + los 3 consumidores | — |
| 2 | `ui/capability.go` + `ui/vt_windows.go` + `ui/vt_other.go` + `capability_test.go` | detección con los 12 casos de §7.3 | 3 |
| 3 | `ui/glyph.go` + `ui/glyph_test.go` | catálogo de §6.4, `G(name)` | 2 |
| 4 | `ui/theme.go` + `ui/theme_test.go` | estilos, identidad en ASCII | 2 |
| 5 | `ui/sanitize.go`, `ui/width.go` | `StripANSI`, envoltura a `ColumnLimit` | 2 |
| 6 | `ui/print.go` + `ui/print_test.go` + `ui/banner_test.go` | `Initialize/Print*/Ok/Fail/…/Banner` | 4, 5 |
| 7 | `ui/parity_test.go` + catálogo de mensajes | **la garantía de §7** | 6 |
| 8 | `launcher-common`: `main.go` no cambia; inicializar `ui` en los `main` de los 4 binarios | `ui.Initialize(os.Stdout, os.Environ())` antes de `commonLogger.Initialize` | 6 |
| 9 | `launcher/internal/cmdUtils/logger/log.go:143-151` | dual sink: fichero intacto, consola estilada + `Ok/Fail/Warn/Info/Detail` | 8 |
| 10 | `launcher/internal/cmd/root.go` | `Step`/`Detail`/`KV`/`Banner` en los ~20 puntos de §6.8.1 | 9 |
| 11 | `launcher/internal/dialog/console.go` | marcador `[ ? ]` y tabla en tiers superiores | 9 |
| 12 | `launcher-config/internal/cmd/{setUp,revert,flushCache,stopAgent}.go`, `internal/cacert.go`, `internal/userData/*.go`, `internal/admin/admin.go` | sustitución mecánica (~125 puntos) | 8 |
| 13 | `launcher-config-admin/internal/cmd/{setUp,revert,flushCache}.go`, `internal/hosts/*.go` | sustitución mecánica (~30 puntos) | 8 |
| 14 | `common/resources/start.bat` | `chcp 65001` | — |
| 15 | `launcher/README.md`, `launcher-config/README.md`, `launcher-config-admin/README.md` | documentar `AGE_LANSERVER_OUTPUT`, `--output`, el tier automático y el `chcp` | 2 |
| 16 | `common/cmd/flagSet.go` | **deuda**: exponer `Fs()` en `RootFlagSet` para registrar `--output` también en `config`/`config-admin`. Requiere tocar `common` → fuera de alcance aquí; se abre como issue | 2 |

---

## 12. Riesgos

| Riesgo | Impacto | Mitigación |
|---|---|---|
| ANSI literal en `conhost` de Win10/11 (hueco H1 de `colorprofile`) | El output se ensucia con `←[32m` | Verificación propia de `ENABLE_VIRTUAL_TERMINAL_PROCESSING` con re-lectura. Test de §10.1. Worst case: sin color, que es el comportamiento de hoy |
| `NO_COLOR` mal interpretado (H2) | El usuario con `NO_COLOR=` ve color | Detección por presencia antes de `colorprofile`. Test dedicado |
| Fuga de ANSI/emoji al **log de fichero** | `launcher.txt` deja de ser grepeable; diagnóstico roto | El estilado se aplica solo en la mitad de stdout (`logger/log.go:143-151`). Test de dual sink (§10.8) |
| `commonLogger.Initialize` pierde su comparador de punteros (§2.5) | Aparecen timestamps + `|MAIN|` en terminal interactiva: regresión | No se toca `Initialize`. Prohibido envolver el writer (§2.5, §4-C). Revisar en review |
| Cambio de flags de `log.Logger` al pasar de `writer != os.Stdout` | como el anterior | Ídem |
| Romper los tests golden de `dialog`/`select_server` | CI rojo | Regla 3 de §6.2: **no TTY ⇒ `TierASCII`** y en `TierASCII` el texto base no cambia. Los tests son la prueba de la regla |
| Emoji de doble ancho en tablas | Columnas desalineadas, aspecto peor que hoy | Regla 3 de §6.4 + test de anchura. Emoji solo en marcadores de línea |
| Fuente raster sin el glifo Unicode elegido | `?` en pantalla | Solo se usan bloques BMP presentes en CP437/CP850 y en las fuentes de consola comunes; test de 1 celda |
| El banner empuja los mensajes importantes fuera de la vista | Peor usabilidad en 80×25 | Banner de 5 líneas, sólo TTY, y arte grande sólo con `Columns >= 100` |
| Reescribir mensajes y perder información por descuido | Paridad rota en Win7 | Contrato de §7 + `assertParity` que **falla la build**. La reescritura no está en la v1 |
| `SetConsoleMode` molesta a alguien | Modo de consola alterado | Sólo se hace `OR` del bit VT si no estaba activo, sólo en stdout, y es reversible con `AGE_LANSERVER_OUTPUT=ascii`. Win7 no se toca (early return por versión) |
| `chcp 65001` en `start.bat` causa problemas en Win7 | `chcp` falla en algunas instalaciones | `>nul 2>&1` y, sobre todo, la regla 6 exige `build >= 10586`, así que el tier no sube en Win7 |
| `gofmt` reescribiendo ficheros ajenos | Ruido en el diff | `gofmt -l ./launcher-common/ui` acotado, como en `ZENITY.md` §7.7 |
| Acoplamiento de `launcher-common/ui` con `launcher` | Rompe la reutilización por `config` | El paquete sólo depende de `lipgloss`, `colorprofile`, `x/ansi`, `x/sys`. Sin `common/logger` |
| El stylado se extiende a `launcher-agent` / `config-admin-agent` "sin querer" | Violación de alcance | Los paquetes no se importan desde ahí; §5.2 lo deja escrito y `go.mod` de esos módulos no cambia |
| Tamaño de binario | ±1–3 MB por dep de Charm | Midiendo (§13). El budget del proyecto es releases de 10–40 MB, así que es aceptable; si no lo fuera, `lipgloss` solo aporta casi todo el valor |

---

## 13. Criterios de aceptación

- [ ] `cmd.exe` de Windows 7 (build 7601, code page 437, sin ConEmu/ANSICON) ejecuta
      `launcher`, `config` y `config-admin` y su salida es **ASCII puro**, sin ningún `ESC`,
      sin ningún `?` de reemplazo, y con **toda** la información del output actual.
- [ ] Un pipe de un `conhost` de Windows 10 sin VT no emite secuencias ANSI (caso H1).
- [ ] `NO_COLOR=` (vacío) desactiva el color (caso H2).
- [ ] El log de fichero `logs/.../launcher.txt` es byte a byte idéntico antes y después del
      cambio (salvo los nuevos `[ ok ]` que se-applicationen en `logger`, si los hay; la v1 los
      escribe **también** en el log, y eso es intencionado y debe documentarse).
- [ ] `launcher/internal/dialog/console_test.go` y
      `launcher/internal/cmdUtils/select_server_test.go` pasan **sin modificar**.
- [ ] `ui/parity_test.go` pasa para el 100% del catálogo de mensajes.
- [ ] `go test -race` verde en los 4 módulos; `go vet` limpio; `gofmt -l` sin salida.
- [ ] `go.mod` de los 13 módulos fuera de alcance, sin cambios.
- [ ] Ningún emoji o símbolo Unicode literal en el código de los 4 módulos (verificado por `grep`).
- [ ] Ningún mensaje pierde información en ningún tier.
- [ ] Documentados en los README: `AGE_LANSERVER_OUTPUT`, `--output`, el `chcp 65001` de
      `start.bat` y qué se espera ver en cada terminal.
- [ ] Documentado que la barra de progreso con `\r` y `bubbletea` quedan fuera (§8.4).
- [ ] Los ficheros de log de `config`/`config-admin` siguen sin ANSI ni glifos.

### 13.1 Delta de binario medido

Pendiente de medir (el mismo procedimiento que `ZENITY.md` §10.1):

```powershell
# OJO: en PowerShell `go build -o $null ./launcher` NO vale, $null se expande a cadena vacía.
$env:CGO_ENABLED = "0"
$tmp = Join-Path $env:TEMP "launcher-build-matrix"
New-Item -ItemType Directory -Force -Path $tmp | Out-Null
foreach ($m in @("launcher", "launcher-common", "launcher-config", "launcher-config-admin")) {
    $out = Join-Path $tmp "$m-before.exe"
    go build -o $out "./$m"
    "{0,-22} {1,10:N0} KB" -f $m, ((Get-Item $out).Length / 1KB)
}
```

Referencia a medir antes/después. Si el delta de `launcher` supera ~3 MB, evaluar
`charmbracelet/x/ansi` (la mayor parte del peso de `lipgloss` v2 está en `x/ansi` + `x/term`),
que es LOW ya está en el árbol por `colorprofile`.

---

## 14. Anexo A — tabla de lo verificado en el código fuente

| Afirmación | Fuente |
|---|---|
| `Style.Render` es puro y no hace downsampling | `lipgloss/style.go` (`Render`); la nota de downsampling solo aparece en `lipgloss/writer.go` |
| `lipgloss.Complete(profile)` es el mecanismo de downsampling | `lipgloss/color.go`, sección "Complete Colors" del README |
| `var Writer = colorprofile.NewWriter(os.Stdout, os.Environ())` y `Print*/Fprint*/Sprint*` usan `Writer.Profile` | `lipgloss/writer.go` |
| `TabWidth` convierte `\t` a 4 espacios; `NoTabConversion` lo desactiva | `lipgloss/style.go`, sección "Tabs" del README |
| `colorprofile.Detect` respeta `NO_COLOR`, `CLICOLOR`, `CLICOLOR_FORCE`, `TTY_FORCE`, `TERM=dumb`, `COLORTERM`, terminfo, tmux | `colorprofile/env.go`, doc-comment de `Detect` |
| `envNoColor` usa `strconv.ParseBool`, luego `NO_COLOR=` vacío no desactiva | `colorprofile/env.go`, `func envNoColor` |
| En Windows, `major < 10 || build < 10586` ⇒ `NoTTY` salvo `ANSICON` | `colorprofile/env_windows.go`, `windowsColorProfile` |
| En Windows ≥ 14931 ⇒ `TrueColor` sin comprobar `ENABLE_VIRTUAL_TERMINAL_PROCESSING` (hueco H1) | `colorprofile/env_windows.go`, última línea de `windowsColorProfile` |
| `WT_SESSION` ⇒ `TrueColor` | `colorprofile/env.go`, `envColorProfile` |
| `colorprofile.NoTTY` y `colorprofile.ASCII` son los perfiles más bajos | `colorprofile/colorprofile.go`, constantes de `Profile` |
| Bubbles v2 ofrece spinner, textinput, textarea, table, progress, paginator, viewport, list, filepicker, timer, stopwatch, help, key | `bubbles/README.md`, índice |
| Bubble Tea v2: `View() tea.View`, `tea.NewView(s)`, `View.AltScreen` (alt-screen ya no es `ProgramOption`), `WithoutSignalHandler`, `WithoutRenderer`, `WithOutput`, `WithColorProfile` | `bubbletea/tea.go` (struct `View`, `func NewView`), `bubbletea/options.go`, `bubbletea/tea.go` (`(*Program).Println/.Printf`) |
| `fang` es un wrapper de **cobra** (`fang.Execute(ctx, *cobra.Command)`) | `fang` README y `pkg.go.dev/github.com/charmbracelet/fang` |