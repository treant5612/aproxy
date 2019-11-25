## aproxy

自用的一个网络代理。支持http/https代理以及socks4/5协议的tcp代理部分。

传输使用aes加密，密钥在配置文件中设置。可以使用tcp直连或者使用websocket协议传输。如使用websocket可以注册一个域名并使用cloudflare提供的免费cdn服务。

支持多级代理，如安装了tor，可在服务端配置下一级代理为tor (localhost:9050)

#### 简单使用说明

**获取**

```shell
go get -u github.com/treant5612/aproxy
```

**作为服务端**

使用 aproxy -s 来生成服务端配置文件config.txt ，然后运行。

默认端口为10808 和 80 (websocket),可以在配置文件中修改。(需要注意防火墙/安全组配置，此外如需使用1024以下的端口需root权限)

```shell
cd ~/go/bin
./aproxy -s -key yourkey
nohup ./aproxy &
```

**作为客户端**

使用aproxy -c 来生成客户端配置文件模板，修改其中的地址和key与服务端相对应之后直接运行即可。

推荐与chrome插件SwitchyOmega配合使用。

在windows下需要后台运行可以使用

```powershell
powershell.exe -WindowStyle Hidden -c ./aproxy.exe
```



#### 配置文件

**客户端**

```
[Proxy]
	# windows下是否修改系统代理设置
	AutoProxy=false

	## Filter指定需要代理的规则列表
	Filter=gfwlist.txt

	## SocksLocal 项设置为true时仅监听来自本机的代理请求
	SocksLocal=false
	
	## 代理端口
	SocksPort=10809

[Tunnel]
	## 数据传输类型 可选项为tcp / websocket 对应服务器配置中的Server/WebsocketServer
	Type=websocket

	## 加密密钥，取其hash作为aes加密的实际密钥
	Key     =defaultKey
	
	## 如果是websocket类型地址应当设置为 ws://aproxy.club/ws 
	Address =ws://aproxy.club/ws
	Port    =80
```

**服务端**

```
#设置密钥以及端口
[Server]
	Key  =defaultKey
	Port =10808
[WebsocketServer]
    Key=defaultKey
    Port=80

#下一级socks代理 比如向tor转发
#[Forward]
#    Address=localhost:9050
```



#### 其它

release中提供了一个已配置好暂时可用的版本。

安卓/ios设备可以在wifi页面点击最右的`>`或`!`符号来配置一个局域网内的HTTP代理。