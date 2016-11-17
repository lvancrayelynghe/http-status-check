package main

import (
    "time"
    "github.com/parnurzeal/gorequest"
)

type Url struct {
    uri string
    response gorequest.Response
    duration time.Duration
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
