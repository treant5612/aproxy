package config

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/treant5612/aproxy/filter"
	"github.com/treant5612/aproxy/socks5"
	"github.com/treant5612/aproxy/tunnel"
	"io"
	"log"
	"net"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"
)

type Config struct {
	Proxy           *Proxy
	Tunnel          *Tunnel
	Server          *Server
	WebsocketServer *WebsocketServer
	Forward         *Forward
}

const anyHost = "0.0.0.0"

func (c *Config) Run() {
	var err error
	var wg sync.WaitGroup
	//向下一级socks代理转发
	if c.Forward != nil {
		tunnel.LocalTunnelConf.SocksForward = c.Forward.Address
	}
	//代理隧道设置
	var transConf *tunnel.Conf = nil
	if c.Tunnel != nil {
		transConf = tunnel.NewTunnelConf(c.Tunnel.Type, c.Tunnel.Key, c.Tunnel.Address, c.Tunnel.Port)
	}
	if c.Proxy != nil {
		//初始化url过滤
		if c.Proxy.Filter != "" {
			if err = filter.Init(c.Proxy.Filter); err != nil {
				log.Printf("failed to use %s as filter:%v", c.Proxy.Filter, err)
			} else {
				log.Printf("use %s as filter\n", c.Proxy.Filter)
			}
			if c.Proxy.AutoProxy {
				SetWindowsProxy(c.Proxy.SocksPort)
			} else {
				UnsetWindowProxy()
			}
			socks := socks5.NewServer("", transConf)
			listenAddr := anyHost
			if c.Proxy.SocksLocal {
				listenAddr = "localhost"
			}
			wg.Add(1)
			go keepRun(&wg, func() {
				socks.ListenAndServe("tcp", net.JoinHostPort(listenAddr, c.Proxy.SocksPort))
			})
		}
	}

	if c.Server != nil {
		server := tunnel.NewRemoteTunnelServer(c.Server.Key)
		wg.Add(1)
		go keepRun(&wg, func() {
			if err := server.ListenAndServe("tcp", net.JoinHostPort(anyHost, c.Server.Port)); err != nil {
				log.Println(err)
			}

		})
	}
	if c.WebsocketServer != nil {
		ws, _ := tunnel.NewWebsocketServer(c.WebsocketServer.Key, "/")
		wg.Add(1)
		go keepRun(&wg, func() {
			ws.ListenWs(net.JoinHostPort(anyHost, c.WebsocketServer.Port))
		})
	}
	wg.Wait()
}

func keepRun(wg *sync.WaitGroup, f func()) {
	defer wg.Done()
	t := time.Now()
	ok := true
	//暂时用这种方式来避免panic导致的跳出
	//即如果服务正常运行超过一个小时那么在它panic的时候再次运行它
	for ok {
		func() {
			defer func() {
				if err := recover(); err != nil {
					log.Println(err)
				}
				if time.Since(t) > time.Hour {
					t = time.Now()
				} else {
					ok = false
				}
			}()
			f()
		}()
	}
}

func OpenConfig(fileName string) *Config {
	file, err := os.Open(fileName)
	if err != nil {
		log.Println(err)
		return nil
	}
	bufReader := bufio.NewReader(file)
	config, section := &Config{}, ""
	configVal := reflect.ValueOf(config)
	for {
		line, _, err := bufReader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err)
			return nil
		}
		if err = setVal(configVal, &section, line); err != nil {
			log.Println(err)
		}
	}
	return config
}

func setVal(target reflect.Value, section *string, line []byte) (err error) {
	line = bytes.TrimSpace(line)
	if len(line) == 0 {
		return
	}
	switch line[0] {
	case '#', '!':
		return
	case '[':
		if line[len(line)-1] == ']' {
			*section = string(line[1 : len(line)-1])
		}
		return
	}
	kv := bytes.Split(line, []byte{'='})
	if len(kv) != 2 {
		return fmt.Errorf("error in %q ", line)
	}
	k, v := bytes.TrimSpace(kv[0]), bytes.TrimSpace(kv[1])
	sub := target.Elem().FieldByName(*section)
	if sub.IsValid() {
		if sub.IsZero() {
			if sub.Kind() == reflect.Ptr {
				sub.Set(reflect.New(sub.Type().Elem()))
			}
		}
		sub = sub.Elem()
		if field := sub.FieldByName(string(k)); field.IsValid() {
			var value interface{}
			switch field.Kind() {
			case reflect.Int:
				value, err = strconv.Atoi(string(v))
			case reflect.Bool:
				value, err = strconv.ParseBool(string(v))
			case reflect.String:
				value = string(v)
			}
			if err != nil {
				return fmt.Errorf("wrong value %q:%v", v, err)
			}
			field.Set(reflect.ValueOf(value))
		}
	}
	return nil
}

type Proxy struct {
	AutoProxy  bool
	Filter     string
	SocksLocal bool
	SocksPort  string
}
type Tunnel struct {
	Type    string
	Key     string
	Address string
	Port    string
}
type Server struct {
	Key  string
	Port string
}
type WebsocketServer struct {
	Key  string
	Port string
}
type Forward struct {
	Address string
}
