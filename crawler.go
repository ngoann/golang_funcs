package main

import (
  "fmt"
  "github.com/gocolly/colly"
  "net/url"
  "regexp"
)

func main() {
  for {
    crawl()
  }
}

func crawl() {
  fmt.Print("Enter product ID (B07BDXRQPC): ")
  product_id := "B07BDXRQPC"
  fmt.Scanln(&product_id)

  c := colly.NewCollector(
    colly.Async(true),
  )

  c.OnHTML("#imgTagWrapperId img", func(e *colly.HTMLElement) {
    decodedValue, _ := url.QueryUnescape(e.Attr("src"))
    r, _ := regexp.Compile(`\|(.{15})\|`)

    fmt.Println(">>>>>> Image link:", "https://m.media-amazon.com/images/I/" + r.FindStringSubmatch(decodedValue)[1])
  })

  c.OnRequest(func(r *colly.Request) {
    fmt.Println("Visiting", r.URL)
  })

  c.Visit("https://www.amazon.com/dp/" + product_id)
  c.Wait()
}
