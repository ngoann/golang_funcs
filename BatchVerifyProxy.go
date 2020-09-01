package main

import (
	// "bytes"
	"log"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/proxy"
	"bufio"
	"os"
	"io/ioutil"
)

var proxyServers = []string{}

func init() {
	env := os.Getenv("ENV")

	if env == "production" {
		file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}

		log.SetOutput(file)
	}
}

func main() {
	GetProxies()
	GetDataFromFreeProxy()

	log.Println("Ready proxy:", len(proxyServers))

	c := colly.NewCollector()
  log.Println("========START========")

	createErr := ioutil.WriteFile("ready_proxy_servers.txt", []byte(""), 0644)
  if createErr != nil {
    log.Fatal(createErr)
  }

	successProxyFile, err := os.OpenFile("ready_proxy_servers.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalf("Failed open file: %s", err)
	}

	dataWriter := bufio.NewWriter(successProxyFile)

	for _, proxyServer := range proxyServers {
    c = colly.NewCollector(
      colly.AllowURLRevisit(),
    )

  	c.OnResponse(func(r *colly.Response) {
			dataWriter.WriteString(proxyServer + "\n")
			dataWriter.Flush()
      log.Println("Checking IP:", proxyServer, "->OK")
  	})

		c.OnError(func(r *colly.Response, err error) {
			log.Println("Checking IP:", proxyServer, "->Failed")
		})

    rp, err := proxy.RoundRobinProxySwitcher(proxyServer)
    if err != nil {
      log.Fatal(err)
    }
    c.SetProxyFunc(rp)
		c.Visit("https://httpbin.org/ip")
	}

	successProxyFile.Close()
  log.Println("=====DONE=====")
}

func GetProxies() {
	log.Println("Visiting to file...")
	file, err := os.Open("proxy_servers.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		proxyServers = append(proxyServers, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func GetDataFromFreeProxy() {
	urlList := []string{"https://free-proxy-list.net"}

	c := colly.NewCollector(
		colly.Async(true),
	)

	for _, url := range urlList {
		c.OnHTML("#proxylisttable tbody tr", func(e *colly.HTMLElement) {
			if e.ChildText("td:nth-child(7)") == "yes" {
				proxyFull := "http://" + e.ChildText("td:nth-child(1)") + ":" + e.ChildText("td:nth-child(2)")
				proxyServers = append(proxyServers, proxyFull)
			}
    })

    c.OnRequest(func(r *colly.Request) {
      log.Println("Visiting:", r.URL)
    })

		c.Visit(url)
	}
	c.Wait()
}
