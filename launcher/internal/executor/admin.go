package executor

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
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

func AddHost(elevate bool, ip string) bool {
	return run(elevate, "addHost", map[string]interface{}{"ip": ip})
}

func RemoveHost(elevate bool) bool {
	return run(elevate, "removeHost", nil)
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
	return shared.TrustCertificate(cert)
}

func RemoveCertificateInternal(elevate bool) bool {
	if elevate {
		return RemoveCertificate(elevate)
	}
	return shared.UntrustCertificate()
}

func AddHostInternal(elevate bool, ip string) bool {
	if elevate {
		return AddHost(elevate, ip)
	}
	return shared.AddHost(ip)
}

func RemoveHostInternal(elevate bool) bool {
	if elevate {
		return RemoveHost(elevate)
	}
	return shared.RemoveHost()
}
