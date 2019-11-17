package filter

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
)

var pacListMap map[string][0]byte

var initialization bool = false

func Init(filepath string) error {
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
	//todo 由于过滤列表大部分为完整的域名，可以用hash表进行查找而无需作为正则进行遍历
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
	if !strings.ContainsAny(str, "*^") {
	}
	str = strings.ReplaceAll(str, ".", `\.`)
	reg, err := regexp.Compile(str)
	if err != nil {
		log.Println(str)
		return
	}
	rules.add(reg, line, "")
}
