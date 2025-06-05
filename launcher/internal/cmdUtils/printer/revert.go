package printer

import (
	"fmt"
	launcherCommon "github.com/luskaner/ageLANServer/launcher-common"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
)

func ConfigRevertPrinter() *launcherCommon.ConfigRevertPrinter {
	return &launcherCommon.ConfigRevertPrinter{
		Revert: func(all bool, requiresRevertAdminElevation bool, stopAgent bool) {
			texts := []*StyledText{T("Reverting ")}
			if all {
				texts = append(texts, T("all possible "))
			}
			texts = append(texts, T("configuration"))
			if stopAgent {
				texts = append(texts, T(" and stopping its agent"))
			}
			if requiresRevertAdminElevation {
				texts = append(
					texts,
					T(", authorize "),
					TS("config-admin", ComponentStyle),
					T(" if needed"),
				)
			}
			texts = append(texts, T("..."))
			fmt.Print(Gen(Configuration, "", texts...))
		},
		RevertFlagsErr: func(err error) {
			PrintSimpln(
				Warning,
				"Failed to get revert flags",
			)
			PrintError(err)
		},
		ClearRevertFlagsErr: func(err error) {
			PrintSimpln(
				Warning,
				"Failed to clear revert flags",
			)
			PrintError(err)
		},
		RevertResult: func(result *exec.Result) {
			if result.Success() {
				PrintSucceeded()
			} else {
				PrintFailedResultError(result)
			}
		},
	}
}
