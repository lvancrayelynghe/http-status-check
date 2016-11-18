package main

import (
    "log"
    "io"
    "os"
    "encoding/csv"
)

func readCSV(filepath string) ([][]string, error) {
    file, err := os.Open(filepath)
    if err != nil {
        return nil, err
    }

    // automatically call Close() at the end of current method
    defer file.Close()

    // options are available at: http://golang.org/src/pkg/encoding/csv/reader.go?s=3213:3671#L94
    reader := csv.NewReader(file)
    reader.Comma = ';'

    datas := make([][]string, 0)
    for {
        // read just one record, but we could ReadAll() as well
        record, err := reader.Read()

        // end-of-file is fitted into err
        if err == io.EOF {
            break
        } else if err != nil {
            return nil, err
        }

        datas = append(datas, record)
    }

    return datas, nil
}

func writeCSV(filepath string, datas [][]string) error {
    file, err := os.Create(filepath)
    if err != nil {
        log.Fatal("Cannot create file", err)
    }

    defer file.Close()

    writer := csv.NewWriter(file)
    writer.Comma = ';'

    for _, line := range datas {
        errWrite := writer.Write(line)
        if errWrite != nil {
            log.Fatal("Cannot create file", errWrite)
        }
    }

    defer writer.Flush()

    return nil
}
