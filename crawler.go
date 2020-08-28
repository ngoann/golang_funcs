package main

import (
  "fmt"
  "github.com/gocolly/colly"
  "net/url"
  "regexp"
  "strings"
  "io"
  "net/http"
  "os"
)

func main() {
  for {
    crawl()
  }
}

func crawl() {
  fmt.Print("Enter product ID (B07BDXRQPC,B07BDXRQP2): ")
  id_input := "B07BDXRQPC"
  fmt.Scanln(&id_input)

  product_ids := strings.Split(id_input, ",")

  c := colly.NewCollector(
    colly.Async(true),
  )

  c.OnHTML("#imgTagWrapperId img", func(e *colly.HTMLElement) {
    decodedValue, _ := url.QueryUnescape(e.Attr("src"))
    product_name := e.Attr("alt") + ".png"
    r, _ := regexp.Compile(`\|(.{15})\|`)

    if len(r.FindStringSubmatch(decodedValue)) > 0 {
      filePath := r.FindStringSubmatch(decodedValue)[1]
      fileUrl := "https://m.media-amazon.com/images/I/" + filePath
      DownloadFile(product_name, fileUrl)
      fmt.Println("Downloaded: " + product_name)
    } else {
      fmt.Println("Failed:", product_name)
    }
  })

  c.OnRequest(func(r *colly.Request) {
    fmt.Println("Visiting", r.URL)
  })

  for _, product_id := range product_ids {
    c.Visit("https://www.amazon.com/dp/" + product_id)
  }

  c.Wait()
}

func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
