package main

import (
    "log"
    "io"
    "os"
    "strconv"
    "time"
    "net/url"
    "net/http"
    "encoding/csv"
    "github.com/jawher/mow.cli"
    "github.com/parnurzeal/gorequest"
    "gopkg.in/go-playground/pool.v3"
)

type Url struct {
    uri string
    response gorequest.Response
    duration time.Duration
}

var urls = make([]string, 0)
var visitedURL map[string]Url = make(map[string]Url)
var request = gorequest.New().Timeout(10000 * time.Millisecond)

func main() {
    app := cli.App("http-status-check", "CLI tool to concurrently checks URLs from a CSV file and check HTTP status code")
    app.Spec = "[-c=<concurrency>] [-i=<input-file-path>] [-o=<output-file-path>] [-n=<new-uri>]"

    var (
        concurrency = app.IntOpt("c concurrency",  5,            "Concurrency")
        inputPath   = app.StringOpt("i input",     "input.csv",  "Input CSV file path")
        outputPath  = app.StringOpt("o output",    "output.csv", "Output CSV file path")
        newuri      = app.StringOpt("n newuri",    "",           "New URI for scheme/host replacements (ie: https://staging.exemple.com/)")
    )

    app.Action = func() {
        if *newuri != "" {
            _, newuriErr := url.ParseRequestURI(*newuri)
            if newuriErr != nil {
                log.Fatalln(newuriErr)
            }
        }
        if _, inputPathErr := os.Stat(*inputPath); os.IsNotExist(inputPathErr) {
            log.Fatalln(inputPathErr)
        }

        log.Println("Starting...")
        timeStart := time.Now()

        log.Println("Concurrency set to", *concurrency)

        err := process(*inputPath, *outputPath, uint(*concurrency), *newuri)
        if err != nil {
            log.Fatalln(err)
        }

        totalTime := time.Since(timeStart)
        log.Println("Done in", totalTime.String())
    }

    app.Run(os.Args)
}

func process(inputPath string, outputPath string, concurrency uint, newuri string) error {
    errRead := readCSV(inputPath)
    if errRead != nil {
        return errRead
    }

    log.Println(len(urls), "urls to parse")

    p := pool.NewLimited(concurrency)
    batch := p.Batch()
    defer p.Close()

    go func() {
        for _, currentUrl := range urls {
            batch.Queue(handleUrl(currentUrl, newuri))
        }

        // DO NOT FORGET THIS OR GOROUTINES WILL DEADLOCK
        // if calling Cancel() it calles QueueComplete() internally
        batch.QueueComplete()
    }()

    for crawl := range batch.Results() {
        if errCrawl := crawl.Error(); errCrawl != nil {
            // handle error
            log.Println("Error: ", errCrawl)
            continue
        }

        currentUrl := crawl.Value().(Url)

        visitedURL[currentUrl.uri] = currentUrl
    }

    errWrite := writeCSV(outputPath, visitedURL)
    if errWrite != nil {
        return errWrite
    }

    return nil
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

    for _, currentUrl := range datas {
        responseCode := currentUrl.response.StatusCode
        duration := currentUrl.duration.String()

        locationValue, locationExist := currentUrl.response.Header["Location"]
        location := ""
        if locationExist {
            location = locationValue[0]
        }

        line := []string{currentUrl.uri, strconv.Itoa(responseCode), duration, location}

        errWrite := writer.Write(line)
        if errWrite != nil {
            log.Fatal("Cannot create file", errWrite)
        }
    }

    defer writer.Flush()

    return nil
}

func handleUrl(currentUrl string, newuri string) pool.WorkFunc  {
    return func(wu pool.WorkUnit) (interface{}, error) {
        if wu.IsCancelled() {
            // return values not used
            return nil, nil
        }

        // Change destination URL
        if newuri != "" {
            currentUrlParsed, currentUrlErr := url.ParseRequestURI(currentUrl)
            if currentUrlErr != nil {
                return nil, currentUrlErr
            }

            newuriParsed, newuriErr := url.ParseRequestURI(newuri)
            if newuriErr != nil {
                return nil, newuriErr
            }
            currentUrlParsed.Scheme = newuriParsed.Scheme
            currentUrlParsed.Host   = newuriParsed.Host
            currentUrl = currentUrlParsed.String()
        }

        log.Println("Checking", currentUrl)

        newUrl := Url{uri: currentUrl}
        errs := newUrl.parseUrl()

        if errs != nil {
            return nil, errs[0]
        }

        return newUrl, nil // everything ok, send nil as 2nd parameter if no error
    }
}

func getUrl(currentUrl string) (gorequest.Response, time.Duration, []error) {
    timeStart := time.Now()

    response, _, err := request.
        Get(currentUrl).
        RedirectPolicy(func(req gorequest.Request, via []gorequest.Request) error {
            return http.ErrUseLastResponse
        }).End()

    return response, time.Since(timeStart), err
}

func (currentUrl *Url) parseUrl() ([]error) {
    response, duration, requestError := getUrl(currentUrl.uri)

    currentUrl.response = response
    currentUrl.duration = duration

    if len(requestError) > 0 {
        return requestError
    }

    return nil
}
