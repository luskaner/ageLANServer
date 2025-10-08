package cmdUtils

import (
	"crypto/x509"
	"fmt"

	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher/internal"
	"github.com/luskaner/ageLANServer/launcher/internal/executor"
)

func (c *Config) AddCACertToGame(gameId string, cert *x509.Certificate, gamePath string) (errorCode int) {
	fmt.Println("Adding CA certificate to game...")
	if result := executor.RunSetUp(gameId, nil, nil, nil, cert.Raw, false, false, false, false, "", "", gamePath); !result.Success() {
		fmt.Println("Failed to save CA certificate to game")
		errorCode = internal.ErrConfigCACertAdd
		if result.Err != nil {
			fmt.Println("Error message: " + result.Err.Error())
		}
		if result.ExitCode != common.ErrSuccess {
			fmt.Printf(`Exit code: %d.`+"\n", result.ExitCode)
		}
	}
	return
}
