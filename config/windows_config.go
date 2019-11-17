package config

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
)

func SetWindowsProxy(port string) {
	if runtime.GOOS != "windows" {
		return
	}
	setProxyRegistry(PROXY_ENABLE, REG_DWROD, "0x01")
	serverValue := fmt.Sprintf("http=127.0.0.1:%s;https=127.0.0.1:%s;socks=127.0.0.1:%s", port, port, port)
	//serverValue := fmt.Sprintf("http=127.0.0.1:%s", port)
	setProxyRegistry(PROXY_SERVER, REG_SZ, serverValue)

}

func UnsetWindowProxy() {
	if runtime.GOOS != "windows" {
		return
	}
	setProxyRegistry(PROXY_ENABLE, REG_DWROD, "0X0")
	setProxyRegistry(PROXY_SERVER, REG_SZ, "")

}

const (
	REG_INTERNET_SETTINGS = `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Internet Settings`
	PROXY_ENABLE          = `ProxyEnable`
	PROXY_SERVER          = "ProxyServer"
	REG_DWROD             = "REG_DWORD"
	REG_SZ                = "REG_SZ"
)

func setProxyRegistry(valueName string, valueType string, valueData string) {
	setRegistry(REG_INTERNET_SETTINGS, valueName, valueType, valueData)
}
func setRegistry(keyName string, valueName string, valueType string, valueData string) {
	cmd := exec.Command("reg", "add", keyName, `/v`, valueName, `/t`, valueType, `/d`, valueData, `/f`)
	result, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("%v\n%s", err, result)
	}
}
