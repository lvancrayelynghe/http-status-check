package main

import (
    "log"
    "os"
    "time"
    "net/url"
    "github.com/jawher/mow.cli"
    "github.com/parnurzeal/gorequest"
    "gopkg.in/go-playground/pool.v3"
)

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
