package cmdUtils

import (
	"path/filepath"
	"testing"

	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
)

// tempStores reemplaza los ArgsStore globales por ficheros temporales para que
// los tests de Config no ensucien el estado real y se restauren automáticamente.
func tempStores(t *testing.T) {
	t.Helper()
	origConfig := launcherCommon.RevertConfigStore
	origCommand := launcherCommon.RevertCommandStore
	launcherCommon.RevertConfigStore = launcherCommon.NewArgsStore(filepath.Join(t.TempDir(), "revert_config.txt"))
	launcherCommon.RevertCommandStore = launcherCommon.NewArgsStore(filepath.Join(t.TempDir(), "revert_command.txt"))
	t.Cleanup(func() {
		launcherCommon.RevertConfigStore = origConfig
		launcherCommon.RevertCommandStore = origCommand
	})
}

func TestConfigSetGameId(t *testing.T) {
	c := &Config{}
	c.SetGameId("age2")
	if c.gameId != "age2" {
		t.Fatalf("gameId = %q, want age2", c.gameId)
	}
}

func TestConfigRequiresConfigRevert(t *testing.T) {
	tempStores(t)
	c := &Config{}

	// Sin store no requiere revert.
	if c.RequiresConfigRevert() {
		t.Fatal("empty store should not require revert")
	}

	if err := launcherCommon.RevertConfigStore.Store([]string{"--ip"}); err != nil {
		t.Fatal(err)
	}
	if !c.RequiresConfigRevert() {
		t.Fatal("store with args should require revert")
	}
}

func TestConfigRevertCommand(t *testing.T) {
	tempStores(t)
	c := &Config{}

	// Sin setupCommandRan siempre devuelve vacío.
	if got := c.RevertCommand(); len(got) != 0 {
		t.Fatalf("RevertCommand() = %v, want empty", got)
	}

	if err := launcherCommon.RevertCommandStore.Store([]string{"echo", "hi"}); err != nil {
		t.Fatal(err)
	}
	c.setupCommandRan = true
	got := c.RevertCommand()
	if len(got) != 2 || got[0] != "echo" || got[1] != "hi" {
		t.Fatalf("RevertCommand() = %v", got)
	}

	// Volver a desactivar setupCommandRan devuelve vacío aunque haya store.
	c.setupCommandRan = false
	if got := c.RevertCommand(); len(got) != 0 {
		t.Fatalf("RevertCommand() with setupCommandRan=false = %v, want empty", got)
	}
}

func TestConfigRequiresRunningRevertCommand(t *testing.T) {
	tempStores(t)

	// setupCommandRan false: nunca requiere.
	c := &Config{setupCommandRan: false}
	if c.RequiresRunningRevertCommand() {
		t.Fatal("setupCommandRan=false should not require running revert")
	}

	// setupCommandRan true pero store vacío: no requiere.
	c.setupCommandRan = true
	if c.RequiresRunningRevertCommand() {
		t.Fatal("setupCommandRan=true with empty store should not require")
	}

	// setupCommandRan true y store con args: requiere.
	if err := launcherCommon.RevertCommandStore.Store([]string{"run"}); err != nil {
		t.Fatal(err)
	}
	if !c.RequiresRunningRevertCommand() {
		t.Fatal("setupCommandRan=true with store should require")
	}
}
