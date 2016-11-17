// https://tour.golang.org/
// https://www.golang-book.com/books/intro

// go get github.com/parnurzeal/gorequest
// go get github.com/PuerkitoBio/goquery
// go get gopkg.in/go-playground/pool.v3

package main

import (
    "log"
    // "fmt"
    "time"
    "net/http"
    "strings"
    // "sync"
    "github.com/parnurzeal/gorequest"
    "github.com/PuerkitoBio/goquery"
    "gopkg.in/go-playground/pool.v3"
)

type Url struct {
    uri string
    visited bool
    response gorequest.Response
    body string
    duration time.Duration
}

var visitedURL map[string]Url = make(map[string]Url)
var request = gorequest.New().Timeout(1000 * time.Millisecond)

var baseURL string = "http://www.exemple.com/"


func main() {
    log.Println("Starting...")


    p := pool.NewLimited(10)
    batch := p.Batch()
    defer p.Close()


    url := Url{uri: baseURL}
    log.Println("Parsing " + url.uri)
    url.parseUrl()
    visitedURL[baseURL] = url

    go func() {
        // Copy the map to avoid overwriting
        work := make(map[string]Url)
        for k,v := range visitedURL {
            work[k] = v
        }
        for _, newUrl := range work {
            if newUrl.visited == false {
                batch.Queue(handleUrl(newUrl))
            }
        }

        // DO NOT FORGET THIS OR GOROUTINES WILL DEADLOCK
        // if calling Cancel() it calles QueueComplete() internally
        batch.QueueComplete()
    }()


    for crawl := range batch.Results() {
        if err := crawl.Error(); err != nil {
            // handle error
        }

        // use return value (url object)
        // log.Println(crawl.Value())
    }

    nop := 0
    yep := 0
    for _, newUrl := range visitedURL {
        if newUrl.visited == false {
            nop += 1
        } else {
            yep += 1
        }
    }

    // log.Println(visitedURL[baseURL])
    log.Println(nop)
    log.Println(yep)
    log.Println(len(visitedURL))
}

func handleUrl(url Url) pool.WorkFunc  {
    return func(wu pool.WorkUnit) (interface{}, error) {
        if wu.IsCancelled() {
            // return values not used
            return nil, nil
        }

        log.Println("Parsing " + url.uri)

        url.parseUrl()

        visitedURL[url.uri] = url

        return url, nil // everything ok, send nil as 2nd parameter if no error
    }
}

func getUrl(url string) (gorequest.Response, string, time.Duration, []error) {
    timeStart := time.Now()

    response, body, err := request.
        Get(url).
        RedirectPolicy(func(req gorequest.Request, via []gorequest.Request) error {
            return http.ErrUseLastResponse
        }).End()

    return response, body, time.Since(timeStart), err
}

func (url *Url) parseUrl() ([]error) {
    response, body, duration, requestError := getUrl(url.uri)

    url.response = response
    url.body = body
    url.duration = duration
    url.visited = true

    if len(requestError) > 0 {
        return requestError
    }

    responseCode := response.StatusCode
    location, locationExist := response.Header["Location"]

    if responseCode == 301 && locationExist {
        location := location[0]

        visitedURL[location] = Url{uri: location}

        return nil
    }

    // Parse
    doc, parserErr := goquery.NewDocumentFromResponse(response)
    if parserErr != nil {
        errors := []error{parserErr}
        return errors
    }

    // Find the review items
    // log.Println(doc.Find("title").Text())
    doc.Find("a").Each(func(i int, element *goquery.Selection) {
        link, linkExist := element.Attr("href")
        _, exists := visitedURL[link]
        if linkExist && link != "#" && strings.HasPrefix(link, baseURL) && ! exists {
            visitedURL[link] = Url{uri: link}
        }
    })

    return nil
}
