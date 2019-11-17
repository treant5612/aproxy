package socks5

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
)

var httpHeaderReg = regexp.MustCompile(`(?i).*\s(.*)\sHTTP/`)

func isHttpRequest(header []byte) bool {
	if httpHeaderReg.Match(header) {
		return true
	}
	return false
}

var (
	httpConnectMethod = []byte("CONNECT")
)

func doHttpProxy(header []byte, request *CommonRequest) (err error) {
	if bytes.HasPrefix(header, httpConnectMethod) {
		return proxyHttps(header, request)
	}
	return proxyHttp(header, request)
}

var (
	getDomainReg = regexp.MustCompile(`(.*?)\shttps?://(.*?)(/.*)\s(.*)`)
)

func proxyHttp(header []byte, request *CommonRequest) (err error) {
	httpRequest := newHttpRequest(request)
	subMatch := getDomainReg.FindSubmatch(header)
	if len(subMatch) != 5 {
		return fmt.Errorf("can not parse http domain")
	}
	method, domain, url, protocol := subMatch[1], subMatch[2], subMatch[3], subMatch[4]
	httpRequest.DstAddr, httpRequest.DstPort = string(domain), "80"
	newHeaderParts := [][]byte{method, url, protocol}
	newHeader := bytes.Join(newHeaderParts, []byte{' '})
	httpRequest.headBuf = append(newHeader, '\r', '\n')
	t, err := httpRequest.newTransporter(httpRequest.DstAddr, httpRequest.DstPort)
	if err != nil {
		return err
	}
	log.Printf("http_proxy: %s\n", request.DstAddr)
	return t.Transport(httpRequest)
}

//我在这里踩了一个坑，刚开始用\r\n作为结尾，导致无法进行代理，连接也被不会关闭
var connectionEstablished = []byte("HTTP/1.1 200 Connection Established\r\n\r\n")

func proxyHttps(header []byte, request *CommonRequest) (err error) {
	//从请求中分离出目标地址
	start, end, split := bytes.IndexByte(header, 0x20), bytes.LastIndexByte(header, 0x20), bytes.IndexByte(header, byte(':'))
	request.DstAddr, request.DstPort = string(header[start+1:split]), string(header[split+1:end])
	t, err := request.newTransporter(request.DstAddr, request.DstPort)
	if err != nil {
		return err
	}
	//丢弃
	request.bufReader.Reset(request.source)
	if _, err = request.source.Write(connectionEstablished); err != nil {
		return err
	}

	log.Printf("https_proxy: %s:%s", request.DstAddr, request.DstPort)
	r := newHttpRequest(request)
	return t.Transport(r.source)
}

type httpRequest struct {
	*CommonRequest
	headBuf []byte
}

func newHttpRequest(r *CommonRequest) *httpRequest {
	return &httpRequest{r, nil}
}

//由于HTTP代理需要修改请求头
func (hr *httpRequest) Read(p []byte) (n int, err error) {
	if len(hr.headBuf) != 0 {
		nHead := copy(p, hr.headBuf)
		hr.headBuf = hr.headBuf[nHead:]
		nReader, err := hr.bufReader.Read(p[nHead:])
		return nReader + nHead, err
	}
	return hr.bufReader.Read(p)
}

func (hr *httpRequest) Write(p []byte) (n int, err error) {
	return hr.source.Write(p)
}
