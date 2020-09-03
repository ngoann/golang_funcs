package main

import (
  "encoding/json"
  "github.com/gocolly/colly"
  "github.com/gocolly/colly/proxy"
  "log"
  "net/http"
  "net/url"
  "os"
  "regexp"
  "math/rand"
  "github.com/gorilla/mux"
  "github.com/rs/cors"
  "bufio"
)

type PageVariables struct {}

type Response struct {
  Url    string
  Status bool
}

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
  port := os.Getenv("PORT")
  r := mux.NewRouter()

  // Handle API routes
  api := r.PathPrefix("/").Subrouter()
  api.HandleFunc("/download", DownloadHandler)

  handlerCORS := cors.Default().Handler(r)

  srv := &http.Server{
    Handler:      handlerCORS,
    Addr:         "0.0.0.0:" + port,
    // WriteTimeout: 15 * time.Second,
    // ReadTimeout:  15 * time.Second,
  }
  log.Println("http://0.0.0.0:" + port)
  log.Fatal(srv.ListenAndServe())
}

func DownloadHandler(w http.ResponseWriter, r *http.Request){
  r.ParseForm()

  if r.Method == http.MethodPost {
    productId := r.Form.Get("product_id")
    link, status := Crawl(productId)
    res := Response{link, status}

    js, err := json.Marshal(res)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(js)
  } else {
    http.Error(w, "Not found!", 404)
  }
}

func Crawl(productId string) (string, bool) {
  GetProxyFromFile()

  crawlStatus, fileUrl := CrawlWithoutProxy(productId)

  if crawlStatus {
    return fileUrl, crawlStatus
  }

  c := colly.NewCollector()

  var getSuccess bool;
  for i := 0; i < 10; i++ {
    randomIndex := rand.Intn(len(proxyServers))
    pickProxyServer := proxyServers[randomIndex]

    c = colly.NewCollector(
      colly.AllowURLRevisit(),
    )

    rp, err := proxy.RoundRobinProxySwitcher(pickProxyServer)
    if err != nil {
    	log.Fatal(err)
    }
    c.SetProxyFunc(rp)

    c.OnHTML("#imgTagWrapperId img", func(e *colly.HTMLElement) {
      decodedValue, _ := url.QueryUnescape(e.Attr("src"))
      designRegex, _ := regexp.Compile(`\|(.{15})\|`)

      if len(designRegex.FindStringSubmatch(decodedValue)) > 0 {
        filePath := designRegex.FindStringSubmatch(decodedValue)[1]
        fileUrl = "https://m.media-amazon.com/images/I/" + filePath
        log.Println(productId, ":", fileUrl, " -> OK")
        getSuccess = true
      } else {
        log.Println(productId, ":", " -> Failed")
      }
    })

    c.OnRequest(func(r *colly.Request) {
      log.Println("Visiting:", r.URL, i + 1, pickProxyServer)
    })

    c.Visit("https://www.amazon.com/dp/" + productId)
    if getSuccess {
      break
    }
  }
  return fileUrl, getSuccess
}

func CrawlWithoutProxy(productId string) (bool, string) {
  var status bool
  var fileUrl string
  c := colly.NewCollector()

  c.OnHTML("#imgTagWrapperId img", func(e *colly.HTMLElement) {
    decodedValue, _ := url.QueryUnescape(e.Attr("src"))
    designRegex, _ := regexp.Compile(`\|(.{15})\|`)

    if len(designRegex.FindStringSubmatch(decodedValue)) > 0 {
      filePath := designRegex.FindStringSubmatch(decodedValue)[1]
      fileUrl = "https://m.media-amazon.com/images/I/" + filePath
      log.Println(productId, ":", fileUrl, " -> OK")
      status = true
    } else {
      log.Println(productId, ":", " -> Failed")
    }
  })

  c.OnRequest(func(r *colly.Request) {
    log.Println("Visiting:", r.URL)
  })

  c.Visit("https://www.amazon.com/dp/" + productId)

  return status, fileUrl
}

func GetProxyFromFile() {
	proxyServers = []string{}

	file, err := os.Open("ready_proxy_servers.txt")
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
