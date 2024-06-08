package executor

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	mapset "github.com/deckarep/golang-set/v2"
	"shared"
	"shared/executor"
)

const processName string = "launcher.admin.exe"

func run(elevate bool, action string, subArguments map[string]interface{}) bool {
	var subArgumentsJsonStr string
	if subArguments != nil {
		jsonBytes, err := json.Marshal(subArguments)
		if err != nil {
			return false
		}
		subArgumentsJsonStr = base64.StdEncoding.EncodeToString(jsonBytes)
	} else {
		subArgumentsJsonStr = "{}"
	}
	args := []string{"-action=" + action, "-subArguments=" + subArgumentsJsonStr}
	if elevate {
		return ElevateCustomExecutable(processName, args...)
	}
	return executor.RunCustomExecutable("./"+processName, args...)
}

func AddHosts(elevate bool, ips mapset.Set[string]) bool {
	return run(elevate, "addHosts", map[string]interface{}{"ips": ips.ToSlice()})
}

func RemoveHosts(elevate bool) bool {
	return run(elevate, "removeHosts", nil)
}

func AddCertificate(elevate bool, cert x509.Certificate) bool {
	base64Cert := base64.StdEncoding.EncodeToString(cert.Raw)
	return run(elevate, "addCert", map[string]interface{}{"certData": base64Cert})
}

func RemoveCertificate(elevate bool) bool {
	return run(elevate, "removeCert", nil)
}

func AddCertificateInternal(elevate bool, cert *x509.Certificate) bool {
	if elevate {
		return AddCertificate(elevate, *cert)
	}
	return shared.TrustCertificate(true, cert)
}

func RemoveCertificateInternal(elevate bool) bool {
	if elevate {
		return RemoveCertificate(elevate)
	}
	return shared.UntrustCertificate(true)
}

func AddHostsInternal(elevate bool, host string) bool {
	ips := shared.HostOrIpToIps(host)
	var ok bool
	if elevate {
		ok = AddHosts(elevate, ips)
	}
	ok = shared.AddHosts(ips)
	shared.ClearResolveCache()
	return ok
}

func RemoveHostsInternal(elevate bool) bool {
	var ok bool
	if elevate {
		ok = RemoveHosts(elevate)
	}
	ok = shared.RemoveHosts()
	shared.ClearResolveCache()
	return ok
}
