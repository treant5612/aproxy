package config

import (
	"fmt"
	"os"
)

/*
根据传入的参数生成相应的配置文件模板
参数为0时生成客户端配置文件，其他生成服务端配置文件
*/
func GenerateConfig(types int, key string) (err error) {
	f, err := os.Create("config.txt")
	if err != nil {
		return err
	}
	defer f.Close()
	switch types {
	case 0:
		_, err = f.WriteString(fmt.Sprintf(templateClient, key))
	default:
		_, err = f.WriteString(fmt.Sprintf(templateServer, key, key))
	}
	return err
}

var templateClient = `
[Proxy]
	AutoProxy=false
	Filter=gfwlist.txt
	SocksLocal=false
	SocksPort=1081

[Tunnel]
	Type	=tcp
	Key     =%s
	Address =192.168.137.1
	Port    =10808
`
var templateServer = `
##以下为服务端设置
[Server]
	Key  =%s
	Port =10808
[WebsocketServer]
    Key=%s
    Port=80
#[Forward]
#    Address=localhost:9050
`
