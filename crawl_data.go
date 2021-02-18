package main

import (
	// "bytes"
	"encoding/csv"
	"fmt"
	"github.com/gocolly/colly"
	"log"
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
	var number int
	fmt.Print("Nhap so luong: ")
	fmt.Scanf("%d", &number)

	GetData(number)
}

func GetData(number int) {
	successCount := 0

	c := colly.NewCollector(
		colly.Async(true),
	)

	file, err := os.Create("result.csv")

	checkError("Cannot create file", err)
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	c.OnHTML(".frame-1", func(e *colly.HTMLElement) {
		if len(e.ChildText(".basic-face p.name")) > 0 {
			log.Println(e.ChildText(".basic-face p.name"))
			log.Println(e.ChildText(".basic-face .col-md-8 p:nth-child(5) > b"))
			log.Println(e.ChildAttr(".no-margin div:nth-child(22) > input", "value"))

			name := e.ChildText(".basic-face p.name")
			address := e.ChildText(".basic-face .col-md-8 p:nth-child(5) > b")
			ssn := e.ChildAttr(".no-margin div:nth-child(22) > input", "value")

			var data = []string{name, address, ssn}

			err = writer.Write(data)
			checkError("Cannot write to file", err)

			successCount++
		}
	})

	// c.OnRequest(func(r *colly.Request) {
	// 	log.Println("Visiting:", r.URL)
	// })

	for successCount <= number {
		now := time.Now()

		c.Visit("https://www.bestrandoms.com/random-identity?new=fresh&state=AK&" +  strconv.Itoa(now.Nanosecond()))
		time.Sleep(1 * time.Second)
	}

	c.Wait()
}

func checkError(message string, err error) {
    if err != nil {
        log.Fatal(message, err)
    }
}
