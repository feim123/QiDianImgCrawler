package main

import (
	"strconv"
	"net/http"
	"regexp"
	"log"
	"bufio"
	"io/ioutil"
	"os"
	"io"
	"bytes"
	"time"
)

const hhh = `https://www.qidian.com/all?orderId=&style=1&pageSize=20&siteid=1&pubflag=0&hiddenField=0&page=`
const pageNum = 43871
var imgRe = regexp.MustCompile(`<img src="(//qidian.qpic.cn/qdbimg/349573/[0-9]+/150)">`)
var imgCount = 1
var rateLimiter = time.Tick(100 * time.Millisecond)

func main() {

	var url string

	urlIn := make(chan string)
	bytesOut := make(chan []byte)
	imgOut := make(chan string)

	for i:=0; i<50; i++{
		go fetch(urlIn, bytesOut)
	}

	for i:=0; i<20; i++{
		go parser(bytesOut, imgOut)
	}

	for i:=0; i<10; i++{
		go imgDownload(imgOut, &imgCount)
	}

	for i:=1; i<=pageNum; i++{
		url = hhh + strconv.Itoa(i)
		urlIn <- url
	}
}

func fetch(in chan string, out chan []byte){
	<- rateLimiter
	for {
		url := <- in
		resp, err := http.Get(url)
		if err != nil{
			log.Printf("fetch %s failed", url)
		}
		defer resp.Body.Close()
		bodyReader := bufio.NewReader(resp.Body)
		body, _ := ioutil.ReadAll(bodyReader)
		out <- body
	}
}

func parser(in chan []byte, out chan string)  {
	for{
		contents := <- in
		matches := imgRe.FindAllSubmatch(contents, -1)
		for _, v := range matches {
			out <- "https:" + string(v[1])
		}
	}
}

func imgDownload(in chan string, imgCount *int)  {
	<- rateLimiter
	for {
		url := <- in
		log.Printf("downloading imgs %s", url)
		resp, err := http.Get(url)
		if err != nil {
			log.Println("download img error url is ", url)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		out, _ := os.Create("I://imgs//" + strconv.Itoa(*imgCount) + ".jpg")
		*imgCount++
		io.Copy(out, bytes.NewReader(body))
	}

}
