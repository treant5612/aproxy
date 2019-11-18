package filter

import (
	"fmt"
	"log"
	"testing"
)

func TestInit(t *testing.T) {
	err := Init("../gfwlist.txt")
	if err != nil {
		log.Println(err)
		t.FailNow()
	}
	fmt.Println(len(plainRules.m))
	fmt.Println(Proxy("www.google.com"))
	fmt.Println(Proxy("www.sogou.com"))
}

func BenchmarkProxy(b *testing.B) {
	Init("../gfwlist.txt")
	for i := 0; i < b.N; i++ {
		Proxy("www.sogou.com")
	}
}

func TestHash(t *testing.T) {
	str := "www.google.com"
	h0 := hash(str[:8])
	h1 := hash(str[1:9])
	h2 := hash(str[2:10])
	log.Println(h0, h1, h2)
	log.Println((h0-uint(str[0])*prime7)*primeRK + uint(str[8]))
	fmt.Println((h1-uint(str[1])*prime7)*primeRK + uint(str[9]))
}
