package config

import (
	"aproxy/filter"
	"aproxy/socks5"
	"aproxy/tunnel"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sync"
)

type Config struct {
	Filter      string
	Socks       *SocksConfig
	Transporter *TunnelConfig
	Server      *ServerConfig
}

func (c *Config) Run() {
	if c.Filter != "" {
		if err := filter.Init(c.Filter); err != nil {
			log.Printf("failed to use %s for proxy filter", c.Filter)
		} else {
			log.Printf("use %s as proxy filter", c.Filter)
		}
	}
	SetWindowsProxy(c.Socks.Port)
	defer UnsetWindowProxy()

	wg := &sync.WaitGroup{}
	var transConf *tunnel.Conf = nil
	if c.Transporter != nil {

		transConf = tunnel.NewTunnelConf("", c.Transporter.Key, c.Transporter.Address, c.Transporter.Port)
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
			err := socks.ListenAndServe("tcp", net.JoinHostPort(listenAddr, c.Socks.Port))
			if err != nil {
				log.Println(err)
			}
		}()
	}
	if c.Server != nil {
		server := new(tunnel.RemoteTunnelServer)
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
type TunnelConfig struct {
	Key     string
	Address string
	Port    string
}
type ServerConfig struct {
	Key  string
	Port string
}
