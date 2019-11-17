package config

import (
	"aproxy/socks5"
	"aproxy/transport"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sync"
)

type Config struct {
	Socks       *SocksConfig
	Transporter *TransporterConfig
	Server      *ServerConfig
}

func (c *Config) Run() {
	SetWindowsProxy(c.Socks.Port)
	defer UnsetWindowProxy()

	wg := &sync.WaitGroup{}
	var transConf *transport.TransConf = nil
	if c.Transporter != nil {

		transConf = transport.NewTransConf("", c.Transporter.Key, c.Transporter.Address, c.Transporter.Port)
	}

	if c.Socks != nil {
		socks := socks5.NewServer(c.Socks.Auth, transConf)
		listenAddr := "0.0.0.0"
		if c.Socks.Local {
			listenAddr = "localhost"
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			socks.ListenAndServe("tcp", net.JoinHostPort(listenAddr, c.Socks.Port))
		}()
	}
	if c.Server != nil {
		server := new(transport.RemoteTransporterServer)
		server.Key = c.Server.Key
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := server.ListenAndServe("tcp", net.JoinHostPort("0.0.0.0", c.Server.Port)); err != nil {
				log.Println(err)
			}
		}()
	}
	wg.Wait()
}
func ParseConfigFromFile(filepath string) (*Config, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return ParseConfig(b)

}
func ParseConfig(configJsonBytes []byte) (*Config, error) {
	conf := new(Config)
	err := json.Unmarshal(configJsonBytes, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

type SocksConfig struct {
	Local    bool
	Port     string
	Auth     string
	Accounts []string
}
type TransporterConfig struct {
	Key     string
	Address string
	Port    string
}
type ServerConfig struct {
	Key  string
	Port string
}
