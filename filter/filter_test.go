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
	fmt.Println(Proxy("www.google.com"))
	fmt.Println(Proxy("www.sogou.com"))
}
