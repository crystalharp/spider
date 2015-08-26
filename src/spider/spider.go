package spider

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
    "sync"
    "runtime"
)

type Spider struct {
	maxLevel int
    ratePerSite int
    downBytes   map[string]int
    lock sync.RWMutex
}

type urlInfo struct {
	level int
	url string
	retry int
}

func NewSpider(maxLevel int, ratePerSite int) Spider {
    spider := Spider{maxLevel:maxLevel, ratePerSite:ratePerSite}
    spider.downBytes = make(map[string]int)
//    spider.lock = sync.RWMutex{}
    return spider
}

func (spider Spider) Submit(start string) {
//	num := runtime.NumCPU()
    runtime.GOMAXPROCS(runtime.NumCPU())
    num := 1
    fmt.Println("cpu num", num)
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
		domain := getDomain(info.url)
		if spider.downBytes[domain] > spider.ratePerSite {
			fmt.Println("speed reached for ", domain, ", skip")
			urls <- info
		} else {
			fmt.Println("downloading level ", info.level, " url:", info.url)
			content, err := download(info.url)
			if err != nil {
				if info.retry < 3 {
					info.retry++
					urls <- info
				}
				fmt.Println(info.url + " download failed")
				continue
			}

            spider.lock.Lock()
			spider.downBytes[domain] += len(content)
            spider.lock.Unlock()
			go spider.parsePage(info, content, urls)
		}
	}
	//close(urls)
	//endInfo <- interface{}
	endInfo <- 1

}

func (spider Spider) parsePage(parent urlInfo, content []byte, urls chan<- urlInfo) {
	if parent.level >= spider.maxLevel {
		return
	}
	scontent := string(content)
	var splits = strings.FieldsFunc(scontent, spliter)
	for _, url := range splits {
		if strings.Index(url, "http://") == 0 {
			fmt.Println("found url:", url)
			urls <- urlInfo{parent.level + 1, url, 0}
		}
	}
    fname := strings.Replace(parent.url, "http://", "", 1)
    fname = strings.Replace(fname, "/", "_", -1)
//    fname = string(parent.level) + "/" + fname
    fname = fmt.Sprintf("%d/%s", parent.level, fname)
	destFile, err := os.Create(fname)
    fmt.Println("create file:", fname, " parent lvl:", parent.level)
	if err != nil {
		fmt.Println("failed to create file:", fname)
		fmt.Println(err)
		return
	}
	defer destFile.Close()
	destFile.WriteString(scontent)
}

func spliter(s rune) bool {
	if s == ' ' || s == '\'' || s == ',' || s == '"'{
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
        spider.lock.Lock()
		for k, _ := range spider.downBytes {
			spider.downBytes[k] = 0
		}
        spider.lock.Unlock()
	}
}
