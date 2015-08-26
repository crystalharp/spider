package main

import (
	"fmt"
	"os"
	"spider"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage:" + os.Args[0] + " url")
		os.Exit(0)
	}
//	var s = spider.Spider{MaxLevel: 3, RatePerSite: 50}
	var s = spider.NewSpider(3, 500)
	s.Submit(os.Args[1])
	//if err != nil {
	//fmt.Println(err)
	//} else {
	//fmt.Println(content)
	//}
}
