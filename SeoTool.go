package main

import (
  "bytes"
  "encoding/json"
  "github.com/gocolly/colly"
  "github.com/gorilla/mux"
  "github.com/rs/cors"
  "html/template"
  "io"
  "log"
  "math/rand"
  "net"
  "net/http"
  "net/url"
  "os"
  "os/exec"
  "regexp"
  "runtime"
  "strconv"
  "fmt"
)

type PageVariables struct {}

type Response struct {
  Url    string
  Status int
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const timeChecking = 20
const version = "1"
var apiURL = "http://128.199.164.114:8888"
var processCount int
var enabled bool
var result map[string]interface{}
var messageAlert string

func init() {
  env := os.Getenv("ENV")
  if env == "development" {
    apiURL = "http://127.0.0.1:8000"
    log.Println("Development environment:", apiURL)
  }

  messageAlert = "Welcome to SEOS TOOL V" + version
  log.Println("Your computer CODE:", strconv.Itoa(int(macUint64())))
  checkingValid()
}

func checkingValid() {
  resp, err := http.PostForm(apiURL + "/checking/valid", url.Values{
    "computer_code": {strconv.Itoa(int(macUint64()))},
    "count": {strconv.Itoa(processCount)},
    "version": {version},
  })

  if err != nil {
    enabled = false
    log.Println("[ERROR] May chu xay ra loi, vui long lien he ADMIN de duoc xu ly!")
    messageAlert = `
      <b style="text-align: center; display: block; font-size: 24px; color: #ff5c97;">
        Máy chủ đang gặp 1 số vấn đề <br /> Vui lòng liên hệ QTV để được xử lý!
      </b>
    `
    openBrowser("http://127.0.0.1:8080/alert")
  } else {
    if resp.StatusCode == 200 {
      // json.NewDecoder(resp.Body).Decode(&result)
      enabled = true
    } else {
      log.Println("[WARN] Tool cua ban da het han hoac may cua ban khong duoc phep su dung!")
      enabled = false
      messageAlert = `
        <b style="text-align: center; display: block; font-size: 24px; color: #ff5c97;">
        TOOL của bạn đã hết hạn hoặc máy của bạn không có quyền sử dụng <br /> Vui lòng liên hệ QTV để được xử lý
      </b>
      `
      openBrowser("http://127.0.0.1:8080/alert")
    }
  }
}

func main() {
  r := mux.NewRouter()
  r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("images/"))))

  api := r.PathPrefix("/").Subrouter()
  api.HandleFunc("/", HomePage)
  api.HandleFunc("/alert", AlertHandler)
  api.HandleFunc("/download", DownloadHandler)

  handlerCORS := cors.Default().Handler(r)

  srv := &http.Server{
    Handler: handlerCORS,
    Addr: "127.0.0.1:8080",
  }
  if enabled {
    log.Println("Visit to http://127.0.0.1:8080", "Version:", version)
    openBrowser("http://127.0.0.1:8080")
  }
  log.Fatal(srv.ListenAndServe())
}

func AlertHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, messageAlert)
}

func HomePage(w http.ResponseWriter, r *http.Request){
    HomePageVars := PageVariables{}

    t, err := template.ParseFiles("index.html")
    if err != nil {
  	  log.Print("Template parsing error: ", err)
  	}

    err = t.Execute(w, HomePageVars)
    if err != nil {
  	  log.Print("Template executing error: ", err)
  	}
}

func DownloadHandler(w http.ResponseWriter, r *http.Request){
  if !enabled {
    return
  }
  r.ParseForm()

  if processCount == timeChecking {
    checkingValid()
    processCount = 0
  }

  if r.Method == http.MethodPost {
    productId := r.Form.Get("product_id")
    link, status := CrawlWithoutProxy(productId)
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

func RandomString() string {
  b := make([]byte, rand.Intn(10)+10)
  for i := range b {
    b[i] = letterBytes[rand.Intn(len(letterBytes))]
  }
  return string(b)
}

func CrawlWithoutProxy(productId string) (string, int) {
  var status int
  var fileUrl string
  var fileName string
  c := colly.NewCollector()

  c.OnHTML("#imgTagWrapperId img", func(e *colly.HTMLElement) {
    decodedValue, _ := url.QueryUnescape(e.Attr("src"))
    designRegex, _ := regexp.Compile(`\|(.{15})\|`)

    if len(designRegex.FindStringSubmatch(decodedValue)) > 0 {
      filePath := designRegex.FindStringSubmatch(decodedValue)[1]
      fileUrl = "https://m.media-amazon.com/images/I/" + filePath
      fileName = "images/" + productId + ".png"

      DownloadFile(fileName, fileUrl)
      status = 1
      processCount++
      log.Println("Download:", productId, "-> OK")
    } else {
      status = 2
      log.Println("Download:", productId, "-> Reject")
    }
  })

  c.OnRequest(func(r *colly.Request) {
    r.Headers.Set("User-Agent", RandomString())
    log.Println("Visiting:", r.URL)
  })

  c.Visit("https://www.amazon.com/dp/" + productId)

  return fileName, status
}

func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func macUint64() uint64 {
  interfaces, err := net.Interfaces()
  if err != nil {
      return uint64(0)
  }

  for _, i := range interfaces {
    if i.Flags&net.FlagUp != 0 && bytes.Compare(i.HardwareAddr, nil) != 0 {

      // Skip locally administered addresses
      if i.HardwareAddr[0]&2 == 2 {
          continue
      }

      var mac uint64
      for j, b := range i.HardwareAddr {
        if j >= 8 {
            break
        }
        mac <<= 8
        mac += uint64(b)
      }

      return mac
    }
  }

  return uint64(0)
}

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	}
	if err != nil {
		log.Fatal(err)
	}
}
