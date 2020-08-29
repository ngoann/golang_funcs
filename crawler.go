package main

import (
  "github.com/gocolly/colly"
  "github.com/gocolly/colly/proxy"
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
  "fmt"
  "math/rand"
)

var proxyServers = []string{
  "http://161.202.226.194:8123",
  "http://139.162.109.91:3128",
  "http://140.227.123.232:3128",
  "http://163.43.108.114:8080",
  "http://150.95.178.151:8888",
  "http://110.44.128.200:3128",
}

func main() {
  for {
    Crawl()
  }
}

func Crawl() {
  fmt.Print("Enter product ID (B07BDXRQPC): ")
  productId := "B07BDXRQPC"
  fmt.Scanln(&productId)

  timeNowStr := strconv.Itoa(int(time.Now().UnixNano()))
  zipSavePath := "assets/" + timeNowStr + ".zip"
  preSavePath := "assets/" + timeNowStr + "/"

  os.MkdirAll(preSavePath, os.ModePerm)

  c := colly.NewCollector()

  var downloaded bool;
  for i := 0; i < 5; i++ {
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
      productName := preSavePath + e.Attr("alt") + ".png"
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
      log.Println(pickProxyServer, "Visiting:", r.URL, i + 1)
    })

    c.Visit("https://www.amazon.com/dp/" + productId)
    if downloaded {
      break
    }
  }

  if downloaded {
    zipit(preSavePath, zipSavePath)
  } else {
    log.Println("Failed to get:", productId)
  }

  os.RemoveAll(preSavePath)
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
