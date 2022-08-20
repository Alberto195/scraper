package scrapers

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"hello/scraper/models"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	userAgent = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:53.0) Gecko/20100101 Firefox/53.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.79 Safari/537.36 Edge/14.14393",
		"Mozilla/5.0 (Linux; Android 6.0.1; SAMSUNG SM-G570Y Build/MMB29K) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/4.0 Chrome/44.0.2403.133 Mobile Safari/537.36",
		"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
	}
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type OidParser struct {
	urlCache   *sync.Map
	httpClient HTTPClient
}

func NewOIDParser(urlCache *sync.Map, httpClient HTTPClient) *OidParser {
	return &OidParser{
		urlCache:   urlCache,
		httpClient: httpClient,
	}
}

func (p *OidParser) Parse(urls chan string, paths chan<- string, pathsToCache chan<- string) error {
	for url := range urls {
		body, err := p.getBody(baseUrl + url)
		if err != nil {
			log.Printf("Couldn`t get body of url %v: %v; Starting retries", url, err.Error())
			body, err = p.startRetries(err, url)
			if err != nil {
				urls <- url
				continue
			}
		}

		data, err := p.filter(body)
		if err != nil {
			return err
		}

		for link := range data {
			pathsToCache <- link
		}

		for link, tableInfo := range data {
			if _, ok := p.urlCache.Load(link); ok {
				continue
			}
			p.urlCache.Store(link, tableInfo)
			paths <- link
		}
	}
	return nil
}

func (p *OidParser) filter(text []byte) (map[string]*models.TableInfo, error) {
	mibData := make(map[string]*models.TableInfo, 10)
	tableData := make([]string, 5)
	isBrothers := false
	doc, err := html.Parse(bytes.NewReader(text))
	if err != nil {
		return nil, fmt.Errorf("can`t parse text: %v", err)
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h3" {
			if strings.Contains(n.FirstChild.Data, "Brothers") {
				isBrothers = true
			}
		}
		if n.Type == html.ElementNode && n.Data == "tr" && !isBrothers {
			if n.FirstChild.FirstChild != nil {
				if n.FirstChild.FirstChild.Data != "Node" && n.FirstChild.FirstChild.Data != "OID" {
					child := n.FirstChild.NextSibling
					for child != nil && len(tableData) < 5 {
						tableData = append(tableData, child.FirstChild.Data)
						child = child.NextSibling
					}
				}
			} else {
				if n.FirstChild.NextSibling.FirstChild.FirstChild.Data != "OID" && !isBrothers {
					child := n.FirstChild.NextSibling.NextSibling.NextSibling
					for child != nil && len(tableData) < 5 {
						if child.FirstChild != nil {
							tableData = append(tableData, child.FirstChild.Data) // 2 ee
							child = child.NextSibling.NextSibling
						} else {
							break
						}
					}
				}
			}
		}
		if n.Type == html.ElementNode && n.Data == "a" && !isBrothers {
			attrValue := n.Attr[0].Val
			if n.Attr[0].Key == "href" &&
				(len(attrValue) < 3 || strings.Contains(attrValue, ".")) &&
				!strings.Contains(attrValue, "http") &&
				!strings.Contains(attrValue, "mib") {
				mibData[attrValue] = mapToTableInfo(tableData)
				tableData = tableData[:0]
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return mibData, nil
}

func (p *OidParser) getBody(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent[rand.Intn(5)])
	response, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal("can`t close body: ", err)
		}
	}(response.Body)
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (p *OidParser) startRetries(err error, url string) ([]byte, error) {
	var body []byte
	retries := 10

	for err != nil || retries == 0 {
		body, err = p.getBody(baseUrl + url)
		retries--
	}
	if err != nil {
		log.Printf("Couldn`t get body of url %v after 10 retries: %v; Timeout for 1 minute", url, err.Error())
		time.Sleep(time.Minute)
		return nil, err
	}

	return body, nil
}

func mapToTableInfo(data []string) *models.TableInfo {
	switch len(data) {
	case 0:
		return &models.TableInfo{}
	case 1:
		return &models.TableInfo{
			Name: data[0],
		}
	case 2:
		subCh, _ := strconv.Atoi(data[1])
		return &models.TableInfo{
			Name:  data[0],
			SubCh: subCh,
		}
	case 3:
		subCh, _ := strconv.Atoi(data[1])
		totalCh, _ := strconv.Atoi(data[2])
		return &models.TableInfo{
			Name:     data[0],
			SubCh:    subCh,
			SubTotal: totalCh,
		}
	case 4:
		subCh, _ := strconv.Atoi(data[1])
		totalCh, _ := strconv.Atoi(data[2])
		return &models.TableInfo{
			Name:     data[0],
			SubCh:    subCh,
			SubTotal: totalCh,
			Desc:     data[3],
		}
	default:
		subCh, _ := strconv.Atoi(data[1])
		totalCh, _ := strconv.Atoi(data[2])
		return &models.TableInfo{
			Name:     data[0],
			SubCh:    subCh,
			SubTotal: totalCh,
			Desc:     data[3],
			Inf:      data[4],
		}
	}
}
