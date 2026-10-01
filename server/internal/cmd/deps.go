package cmd

import (
	"io"
	"net"
	"net/http"
	"net/netip"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/executor"
	"github.com/luskaner/ageLANServer/common/fileLock"
	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/luskaner/ageLANServer/common/uuid"
	"github.com/luskaner/ageLANServer/server/internal"
	"github.com/luskaner/ageLANServer/server/internal/ip"
	"github.com/luskaner/ageLANServer/server/internal/logger"
	"github.com/luskaner/ageLANServer/server/internal/models"
	"github.com/luskaner/ageLANServer/server/internal/models/initializer"
	"github.com/spf13/pflag"
)

var (
	fileLockNewFn            = func() *fileLock.PidLock { return &fileLock.PidLock{} }
	fileLockLockFn           = func(l *fileLock.PidLock) error { return l.Lock() }
	fileLockUnlockFn         = func(l *fileLock.PidLock) error { return l.Unlock() }
	commonLoggerInitFn       = commonLogger.Initialize
	commonLoggerCloseFn      = commonLogger.CloseFileLog
	loggerOpenMainFileLogFn  = logger.OpenMainFileLog
	isAdminFn                = executor.IsAdmin
	dnsConnectivityFn        = common.DNSConnectivity
	cacheNetworkInterfacesFn = models.CacheNetworkInterfaces
	uuidParseFn              = uuid.Parse
	initConfigFn             = initConfig
	resolveHostsFn           = ip.ResolveHosts
	initializeGameFn         = initializer.InitializeGame
	certificatePairFolderFn  = common.CertificatePairFolder
	queryConnectionsFn       = ip.QueryConnections
	listenQueryConnectionsFn = ip.ListenQueryConnections
	newHTTPServerFn          = func(addr string, handler http.Handler) *http.Server {
		return &http.Server{Addr: addr, Handler: handler}
	}
	initializeRngFn = internal.InitializeRng
	_               = net.ParseIP
	_               = io.Writer(nil)
	_               = pflag.FlagSet{}
	_               = mapset.NewThreadUnsafeSet[string]
	_               = netip.Addr{}
)
