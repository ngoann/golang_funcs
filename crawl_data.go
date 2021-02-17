package main

import (
	// "bytes"
	"log"
	"github.com/gocolly/colly"
	"os"
	"strconv"
	"time"
)

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
	GetData()
}

func GetData() {
	c := colly.NewCollector(
		colly.Async(true),
	)

	c.OnHTML(".frame-1", func(e *colly.HTMLElement) {
		log.Println(e.ChildText(".basic-face p.name"))
		log.Println(e.ChildText(".basic-face .col-md-8 p:nth-child(5) > b"))
		// log.Println(e.ChildText(".no-margin .input:nth-child(11)", "class"))
		log.Println(e.ChildAttr(".no-margin div:nth-child(22) > input", "value"))
		// if e.ChildText("td:nth-child(7)") == "yes" {
			// 	// proxyFull := "http://" + e.ChildText("td:nth-child(1)") + ":" + e.ChildText("td:nth-child(2)")
			// 	// proxyServers = append(proxyServers, proxyFull)
			// }
	})

	// c.OnRequest(func(r *colly.Request) {
	// 	log.Println("Visiting:", r.URL)
	// })

	for i := 0; i < 999999; i++ {
		c.Visit("https://www.bestrandoms.com/random-identity?new=fresh&state=AK&" +  strconv.Itoa(i))
		time.Sleep(2 * time.Second)
	}

	c.Wait()
}
