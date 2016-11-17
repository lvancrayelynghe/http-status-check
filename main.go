// https://tour.golang.org/
// https://www.golang-book.com/books/intro

// go get github.com/parnurzeal/gorequest
// go get gopkg.in/go-playground/pool.v3

package main

import (
    "log"
    "io"
    "os"
    "encoding/csv"
    "strconv"
    "time"
    "net/http"
    "strings"
    "github.com/parnurzeal/gorequest"
    "gopkg.in/go-playground/pool.v3"
)

type Url struct {
    uri string
    response gorequest.Response
    duration time.Duration
}

var visitedURL map[string]Url = make(map[string]Url)
var request = gorequest.New().Timeout(10000 * time.Millisecond)

var baseURL string = "http://www.exemple.com/"
var replaceURL string = "https://recette.exemple.com/"

var urls = make([]string, 0)

func main() {
    log.Println("Starting...")

    err := readCSV("input.csv")
    if err != nil {
        log.Fatalln(err)
    }

    p := pool.NewLimited(10)
    batch := p.Batch()
    defer p.Close()

    go func() {
        for _, url := range urls {
            batch.Queue(handleUrl(url))
        }

        // DO NOT FORGET THIS OR GOROUTINES WILL DEADLOCK
        // if calling Cancel() it calles QueueComplete() internally
        batch.QueueComplete()
    }()

    for crawl := range batch.Results() {
        if err := crawl.Error(); err != nil {
            // handle error
            log.Println("Error: ", err)
        }

        url := crawl.Value().(Url)

        visitedURL[url.uri] = url
    }

    writeCSV("output.csv", visitedURL)

    log.Println("Done !")
}

func readCSV(filepath string) error {
    file, err := os.Open(filepath)
    if err != nil {
        return err
    }

    // automatically call Close() at the end of current method
    defer file.Close()

    // options are available at: http://golang.org/src/pkg/encoding/csv/reader.go?s=3213:3671#L94
    reader := csv.NewReader(file)
    reader.Comma = ';'

    for {
        // read just one record, but we could ReadAll() as well
        record, err := reader.Read()

        // end-of-file is fitted into err
        if err == io.EOF {
            break
        } else if err != nil {
            return err
        }

        urls = append(urls, record[0])
    }

    return nil
}

func writeCSV(filepath string, datas map[string]Url) error {
    file, err := os.Create(filepath)
    if err != nil {
        log.Fatal("Cannot create file", err)
    }

    defer file.Close()

    writer := csv.NewWriter(file)
    writer.Comma = ';'

    line := []string{"Url", "Response Code", "Duration", "Redirect"}
    errWrite := writer.Write(line)
    if errWrite != nil {
        log.Fatal("Cannot create file", errWrite)
    }

    for _, url := range datas {
        responseCode := url.response.StatusCode
        duration := url.duration.String()

        locationValue, locationExist := url.response.Header["Location"]
        location := ""
        if locationExist {
            location = locationValue[0]
        }

        line := []string{url.uri, strconv.Itoa(responseCode), duration, location}

        errWrite := writer.Write(line)
        if errWrite != nil {
            log.Fatal("Cannot create file", errWrite)
        }
    }

    defer writer.Flush()

    return nil
}

func handleUrl(url string) pool.WorkFunc  {
    return func(wu pool.WorkUnit) (interface{}, error) {
        if wu.IsCancelled() {
            // return values not used
            return nil, nil
        }

        // Change destination URL
        url = strings.Replace(url, baseURL, replaceURL, -1)

        log.Println("Parsing " + url)

        newUrl := Url{uri: url}
        errs := newUrl.parseUrl()

        if errs != nil {
            return nil, errs[0]
        }

        return newUrl, nil // everything ok, send nil as 2nd parameter if no error
    }
}

func getUrl(url string) (gorequest.Response, time.Duration, []error) {
    timeStart := time.Now()

    response, _, err := request.
        Get(url).
        RedirectPolicy(func(req gorequest.Request, via []gorequest.Request) error {
            return http.ErrUseLastResponse
        }).End()

    return response, time.Since(timeStart), err
}

func (url *Url) parseUrl() ([]error) {
    response, duration, requestError := getUrl(url.uri)

    url.response = response
    url.duration = duration

    if len(requestError) > 0 {
        return requestError
    }

    return nil
}
