package main

import (
    "os"
    "io/ioutil"
    "encoding/xml"
)

type Query struct {
    // Have to specify where to find the data since this doesn't match the xml tags go needs to go into
    UrlList []SitemapUrl `xml:"url"`
}

type SitemapUrl struct {
    Loc        string `xml:"loc"`
    LastMod    string `xml:"lastmod"`
    ChangeFreq string `xml:"changefreq"`
    Priority   string `xml:"priority"`
}

func readSitemapXML(filepath string) ([][]string, error) {
    file, err := os.Open(filepath)
    if err != nil {
        return nil, err
    }

    // automatically call Close() at the end of current method
    defer file.Close()

    bytes, _ := ioutil.ReadAll(file)

    var query Query
    xml.Unmarshal(bytes, &query)

    datas := make([][]string, 0)
    for _, sitemapUrl := range query.UrlList {
        data := []string{sitemapUrl.Loc, sitemapUrl.LastMod, sitemapUrl.ChangeFreq, sitemapUrl.Priority}
        datas = append(datas, data)
    }

    return datas, nil
}
