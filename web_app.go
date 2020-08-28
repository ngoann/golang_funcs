package main

import (
  "encoding/json"
  "github.com/gocolly/colly"
  "html/template"
  "io"
  "log"
  "net/http"
  "net/url"
  "os"
  "regexp"
  "time"
  "strconv"
  "archive/zip"
  "path/filepath"
  "strings"
)

type PageVariables struct {}

type Response struct {
  Url    string
  Downloaded int
}

func main() {
	http.HandleFunc("/", HomePage)
	http.HandleFunc("/download", DownloadHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
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
    productIds := r.Form["product_ids[]"]
    link, successCount := crawl(productIds)
    res := Response{link, successCount}

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

func crawl(productIds []string) (string, int) {
  var successCount int
  timeNowStr := strconv.Itoa(int(time.Now().UnixNano()))
  zipSavePath := "assets/" + timeNowStr + ".zip"
  preSavePath := "assets/" + timeNowStr + "/"
  os.MkdirAll(preSavePath, os.ModePerm)

  c := colly.NewCollector(
    // colly.Async(true),
  )

  c.OnHTML("#imgTagWrapperId img", func(e *colly.HTMLElement) {
    decodedValue, _ := url.QueryUnescape(e.Attr("src"))
    productName := preSavePath + e.Attr("alt") + ".png"
    designRegex, _ := regexp.Compile(`\|(.{15})\|`)

    if len(designRegex.FindStringSubmatch(decodedValue)) > 0 {
      filePath := designRegex.FindStringSubmatch(decodedValue)[1]
      fileUrl := "https://m.media-amazon.com/images/I/" + filePath
      DownloadFile(productName, fileUrl)
      log.Println("Downloaded:", productName)
      successCount ++
    } else {
      log.Println("Failed:", productName)
    }
  })

  c.OnRequest(func(r *colly.Request) {
    log.Println("Visiting:", r.URL)
  })

  for _, productId := range productIds {
    c.Visit("https://www.amazon.com/dp/" + productId)
  }

  // c.Wait()
  zipit(preSavePath, zipSavePath)
  os.RemoveAll(preSavePath)
  return zipSavePath, successCount
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
