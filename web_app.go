package main

import (
  "encoding/json"
  "github.com/gocolly/colly"
  "github.com/gocolly/colly/proxy"
  "html/template"
  "io"
  "log"
  "net/http"
  "net/url"
  "os"
  "regexp"
  "archive/zip"
  "path/filepath"
  "strings"
  "math/rand"
  "github.com/gorilla/mux"
  "github.com/rs/cors"
)

type PageVariables struct {}

type Response struct {
  Url    string
  Status bool
}

var proxyServers = []string{
  "http://110.44.128.200:3128",
  "http://117.2.165.159:53281",
  "http://121.50.44.14:3128",
  "http://133.130.111.34:60088",
  "http://139.162.109.91:3128",
  "http://140.227.123.218:3128",
  "http://140.227.123.220:3128",
  "http://140.227.123.232:3128",
  "http://150.95.178.151:8888",
  "http://153.142.70.170:8080",
  "http://161.202.226.194:8123",
  "http://163.43.108.114:8080",
  "http://173.82.74.62:5836",
  "http://173.82.78.187:5836",
  "http://18.166.13.99:8080",
  "http://209.97.137.39:80",
  "http://27.72.29.159:8080",
  "http://3.19.234.208:3128",
  "http://34.68.103.187:3128",
  "http://34.72.12.158:3128",
  "http://34.95.207.212:3128",
  "http://45.77.27.87:8080",
  "http://68.183.121.227:3128",
}

func main() {
  port := os.Getenv("PORT")
  r := mux.NewRouter()

  // Handle API routes
  api := r.PathPrefix("/").Subrouter()
  api.HandleFunc("/", HomePage)
  api.HandleFunc("/download", DownloadHandler)

  // Serve static files
  r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets/"))))

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

func HomePage(w http.ResponseWriter, r *http.Request){
    HomePageVars := PageVariables{}

    t, err := template.ParseFiles("homepage.html") //parse the html file homepage.html
    if err != nil { // if there is an error
  	  log.Print("template parsing error: ", err) // log it
  	}
    err = t.Execute(w, HomePageVars) //execute the template and pass it the HomePageVars struct to fill in the gaps
    if err != nil { // if there is an error
  	  log.Print("template executing error: ", err) //log it
  	}
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
  preSavePath := "assets/"
  productName := preSavePath + productId + ".png"
  os.MkdirAll(preSavePath, os.ModePerm)

  c := colly.NewCollector()

  var downloaded bool;
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
        fileUrl := "https://m.media-amazon.com/images/I/" + filePath
        DownloadFile(productName, fileUrl)
        log.Println("Downloaded:", productName)
        downloaded = true
      } else {
        log.Println("Failed:", productName)
      }
    })

    c.OnRequest(func(r *colly.Request) {
      log.Println("Visiting:", r.URL, i + 1, pickProxyServer)
    })

    c.Visit("https://www.amazon.com/dp/" + productId)
    if downloaded {
      break
    }
  }
  return productName, downloaded
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

// ZIP FILE
func zipit(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

  log.Println("Ziped:", target)
	return err
}
