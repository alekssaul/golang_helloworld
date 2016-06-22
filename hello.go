package main

import (
  "fmt"
  "net/http"
  "net"  
)

const (
  port = ":80"
  version = "1.1"
)

func GetLocalIP() string {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return ""
    }
    for _, address := range addrs {
        // check the address type and if it is not a loopback the display it
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                return ipnet.IP.String()
            }
        }
    }
    return ""
}

func HelloWorld(w http.ResponseWriter, r *http.Request) {  
  fmt.Fprintf(w, "Hello, world! version: %s ", version)
  fmt.Fprintf(w, "Server IP: %s \n", GetLocalIP())
}

func init() {
  fmt.Printf("Started Application version:%s \n", version)
  fmt.Printf("server IP:%s \n", GetLocalIP())  
  http.HandleFunc("/", HelloWorld)
  http.ListenAndServe(port, nil)
}

func main() {}
