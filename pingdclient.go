package main

import (
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url"
    "os"
    "log"
    "strings"
)

func main() {
    var resp *http.Response
    var err error

    args := os.Args

    if len(args)>=2 {
        if len(args)==2 {
          resp, err = http.Get("http://127.0.0.1:7008/ping4?ip="+url.QueryEscape(args[1]))
        }

        if len(args)>2 {
          resp, err = http.Get("http://127.0.0.1:7008/ping4?ip="+url.QueryEscape(args[1])+"&timeout="+args[2])
        }

        if err!=nil {
            log.Fatal(err)
        }

        sizereturned, err := ioutil.ReadAll(resp.Body)
        resp.Body.Close()

        if err!=nil {
            log.Fatal(err)
        }

        lines := strings.Split(string(sizereturned),"\n")

        fmt.Printf("%s\n",lines[0])
    } else {
        fmt.Println("Usage: pingdclient <ipv4 name/address> <timeout in seconds>")
    }
}
