package main

import (
    "log"
    "io"
    "os"
    "strconv"
    "encoding/csv"
)

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
