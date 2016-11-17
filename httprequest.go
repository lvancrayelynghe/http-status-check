package main

import (
    "time"
    "net/http"
    "github.com/parnurzeal/gorequest"
)

func getUrl(currentUrl string) (gorequest.Response, time.Duration, []error) {
    timeStart := time.Now()

    response, _, err := request.
        Get(currentUrl).
        RedirectPolicy(func(req gorequest.Request, via []gorequest.Request) error {
            return http.ErrUseLastResponse
        }).End()

    return response, time.Since(timeStart), err
}
