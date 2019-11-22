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
	AutoProxy       bool
	Filter          string
	Socks           *SocksConfig
	Transporter     *TunnelConfig
	Server          *ServerConfig
	WebsocketServer *WebsocketConfig
	Forward         string
}

func (c *Config) Run() {
	if c.Forward != "" {
		tunnel.LocalTunnelConf.SocksForward = c.Forward
	}
	if c.Filter != "" {
		if err := filter.Init(c.Filter); err != nil {
			log.Printf("failed to use %s for proxy filter", c.Filter)
		} else {
			log.Printf("use %s as proxy filter", c.Filter)
		}
	}
	if c.AutoProxy {
		SetWindowsProxy(c.Socks.Port)
		defer UnsetWindowProxy()
	}
	wg := &sync.WaitGroup{}
	var transConf *tunnel.Conf = nil
	if c.Transporter != nil {
		transConf = tunnel.NewTunnelConf(c.Transporter.Type, c.Transporter.Key, c.Transporter.Address, c.Transporter.Port)
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
		server := tunnel.NewRemoteTunnelServer(c.Server.Key)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := server.ListenAndServe("tcp", net.JoinHostPort("0.0.0.0", c.Server.Port)); err != nil {
				log.Println(err)
			}
		}()
	}
	if c.WebsocketServer != nil {
		wsServer, err := tunnel.NewWebsocketServer(c.WebsocketServer.Key, "/")
		if err != nil {
			log.Println(err)
		} else {
			wg.Add(1)
			go func() {
				defer wg.Done()
				wsServer.ListenWs("0.0.0.0:80")
			}()
		}
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
	Type    string
	Key     string
	Address string
	Port    string
}
type ServerConfig struct {
	Key  string
	Port string
}

type WebsocketConfig struct {
	Key  string
	Port string
}
