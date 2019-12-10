package filter

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var pacListMap map[string][0]byte

var initialization bool = false
var initLock sync.RWMutex

func Init(filepath string) error {
	initLock.Lock()
	defer initLock.Unlock()
	if initialization {
		return nil
	}
	text, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	initFromText(text)
	initialization = true
	return nil
}

//如果初始化完成则使用地址过滤
func Proxy(address string) bool {
	if initialization {
		return rules.Test(address)
	}
	return true
}

func initFromText(text []byte) {
	bufReader := bufio.NewReader(bytes.NewReader(text))

	var preline []byte
	for {
		lineBuf, isPrefix, err := bufReader.ReadLine()
		if err != nil {
			return
		}
		if isPrefix {
			preline = lineBuf
			continue
		} else {
			preline = preline[:]
		}
		line := append(lineBuf, preline...)
		if len(line) == 0 || line[0] == '!' || line[0] == '[' || line[0] == '@' {
			continue
		}
		buildReg(line)
	}

}

func buildReg(line []byte) (*regexp.Regexp, error) {
	toRegex(line)
	return nil, nil
}

//用于转换adb的域名匹配规则到正则表达式
//比如 将.转为 \.
//将表示开头或结尾的|转为 ^或$
type filterRule struct {
	reg *regexp.Regexp
	raw []byte
	key string
}
type Rules struct {
	rules []*filterRule
}

func (r *Rules) add(reg *regexp.Regexp, raw []byte, key string) {
	r.rules = append(r.rules, &filterRule{reg, raw, key})
}

func (r Rules) Test(s string) bool {
	if plainRules.contains(s) {
		return true
	}
	for _, fr := range r.rules {
		if fr.reg.MatchString(s) {
			return true
		}
	}
	return false
}

var rules Rules

var (
	adbRule1 = regexp.MustCompile(`^\|`)
	adbRule2 = regexp.MustCompile(`\|$`)
	adbRule3 = regexp.MustCompile(`https?://(.*)/.*|/$`)
)

func toRegex(line []byte) {
	if line[0] == '/' && line[len(line)-1] == '/' {
		reg, err := regexp.Compile(string(line[1 : len(line)-1]))
		if err != nil {
			return
		}
		rules.add(reg, line, "")
	}

	lineStr := string(line)
	str := strings.ReplaceAll(lineStr, "||", "")
	str = strings.ReplaceAll(str, "*", ".*")
	str = adbRule1.ReplaceAllString(str, "^")
	str = adbRule2.ReplaceAllString(str, "$")
	if adbRule3.MatchString(str) {
		str = adbRule3.ReplaceAllString(str, "${1}")
	}
	if len(str) >= 8 || !strings.ContainsAny(str, "*^") {

		reg, err := regexp.Compile(str)
		if err != nil {
			log.Println(str)
			return
		}
		plainRules.add(&filterRule{reg, line, str,})
		return
	}
	str = strings.ReplaceAll(str, ".", `\.`)
	reg, err := regexp.Compile(str)
	if err != nil {
		log.Println(str)
		return
	}
	rules.add(reg, line, "")
}

type plainRuleList struct {
	m map[uint][]*filterRule
}

var plainRules *plainRuleList = newPlainRules()

func newPlainRules() *plainRuleList {
	m := make(map[uint][]*filterRule)
	return &plainRuleList{m}
}

func (p *plainRuleList) add(rule *filterRule) {
	var key string
	if len(rule.key) < 8 {
		key = rule.key
	} else {
		key = rule.key[:8]
	}
	h := hash(key)
	p.m[h] = append(p.m[h], rule)
}

func (p *plainRuleList) contains(str string) bool {
	var hashes = make([]uint, 0)
	if len(str) < 8 {
		hashes = append(hashes, hash(str))
	} else {
		h0 := hash(str[:8])
		hashes = append(hashes, h0)
		for i := 0; i < len(str)-8; i++ {
			h := (hashes[i]-uint(str[i])*prime7)*primeRK + uint(str[i+8])
			hashes = append(hashes, h)
		}

	}
	for _, v := range hashes {
		if len(p.m[v]) > 0 {
			for _, rule := range p.m[v] {
				if rule.reg.MatchString(str) {
					return true
				}
			}
		}
	}
	return false
}

const primeRK = 16777619
const prime7 uint = 12960422244463762683

//var prime7 = primeRK * primeRK * primeRK * primeRK * primeRK * primeRK * primeRK

func hash(str string) uint {
	var n uint
	for _, v := range str {
		n = n*primeRK + uint(v)
	}
	return n
}

var gfwlistUrl = "https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt"
var DefaultGfwlist = "gfwlist.txt"

func UpdateGfwlist(proxyUrl string) {
	updateGfwlist(proxyUrl)
}

func updateGfwlist(proxyUrl string) {
	if !shouldUpdateGfwlist(DefaultGfwlist) {
		return
	}
	log.Println("start downloading gfwlist.txt...")
	proxy := func(_ *http.Request) (*url.URL, error) {
		return url.Parse(proxyUrl)
	}
	transport := &http.Transport{Proxy: proxy}
	httpClient := http.Client{Transport: transport}

	url := "https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt"
	resp, err := httpClient.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	decoder := base64.NewDecoder(base64.StdEncoding, resp.Body)
	f, err := os.Create(DefaultGfwlist)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	io.Copy(f, decoder)
	log.Println("download gfwlist.txt complete!")
}

//check if necessary to update gfwlist file
func shouldUpdateGfwlist(fileName string) bool {
	stat, _ := os.Stat(fileName)
	if stat == nil {
		return true //file not exist
	}
	if time.Since(stat.ModTime()) > time.Hour*24*7 {
		return true
	}
	return false
}
