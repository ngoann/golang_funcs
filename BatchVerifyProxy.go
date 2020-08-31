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

func main() {
	GetProxies()

	c := colly.NewCollector()
  log.Println("=====START=====")

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
      log.Println("IP OK:", proxyServer)
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
	proxyServers = []string{}

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
