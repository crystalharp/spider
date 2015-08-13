package spider

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

type Spider struct {
	MaxLevel    int
	RatePerSite int
	downBytes   map[string]int
}

type urlInfo struct {
	Level int
	Url   string
	retry int
}

func (spider Spider) Submit(start string) {
	spider.downBytes = make(map[string]int)
	num := runtime.NumCPU()
	urls := make(chan urlInfo, num*2)
	urls <- urlInfo{0, start, 0}
	endInfo := make(chan interface{})

	ticker := time.NewTicker(2 * time.Second)
	go spider.clearSpeedRecord(ticker.C)
	for i := 0; i < num; i++ {
		go spider.download(urls, endInfo)
	}
	for i := 0; i < num; i++ {
		<-endInfo
		fmt.Println("one spider ended")
	}
	ticker.Stop()

}

func (spider Spider) download(urls chan urlInfo, endInfo chan<- interface{}) {

	for info := range urls {
		domain := getDomain(info.Url)
		if spider.downBytes[domain] > spider.RatePerSite {
			fmt.Println("speed reached for ", domain, ", skip")
			urls <- info
		} else {
			fmt.Println("downloading level ", info.Level, " url:", info.Url)
			content, err := download(info.Url)
			if err != nil {
				if info.retry < 3 {
					info.retry++
					urls <- info
				}
				fmt.Println(info.Url + " download failed")
				continue
			}

			spider.downBytes[domain] += len(content)
			go spider.parsePage(info, content, urls)
			//urlList := spider.parsePage(info, content)
			//for _, info := range urlList {
			//urls <- info
			//}
		}
	}
	//close(urls)
	//endInfo <- interface{}
	endInfo <- 1

}

func (spider Spider) parsePage(parent urlInfo, content []byte, urls chan<- urlInfo) {
	if parent.Level >= spider.MaxLevel {
		return
	}
	scontent := string(content)
	//var splits = strings.Split(scontent, " ")
	var splits = strings.FieldsFunc(scontent, filter)
	for _, url := range splits {
		if strings.Index(url, "http://") == 0 {
			fmt.Println("found url:", url)
			urls <- urlInfo{parent.Level + 1, url, 0}
		}
	}
	destFile, err := os.Create("./" + string(parent.Level) + "/" + strings.Replace(parent.Url, "http://", "", 1))
	if err != nil {
		fmt.Println("failed to create file:", parent.Url)
		fmt.Println(err)
		return
	}
	defer destFile.Close()
	destFile.WriteString(scontent)
}

func filter(s rune) bool {
	if s == ' ' || s == '\'' || s == ',' {
		return true
	}
	return false
}

/*
func splitToken(String s) []string {
	list := make([]string)
	start := 0
	end := 0
	for i := 0; i < len(s); i++ {
		if s[i]
	}

}*/

func download(url string) ([]byte, error) {
	//fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("get " + url + " failed")
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("read body failed")
		return nil, err
	}
	return body, nil
}

func getDomain(url string) string {
	splits := strings.Split(url, "/")
	return splits[2]
}

func (spider Spider) clearSpeedRecord(c <-chan time.Time) {
	for {
		<-c
		fmt.Println("tick, clearSpeedRecord")
		for k, _ := range spider.downBytes {
			spider.downBytes[k] = 0
		}
	}
}
