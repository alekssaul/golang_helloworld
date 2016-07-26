package main

import (
  "fmt"
  "net/http"
  "net"
  "os" 
  "log"
  "io/ioutil"
  "strings"
  "github.com/oschwald/geoip2-golang"
  "github.com/kelseyhightower/envconfig"
  "time"
)

const (
  port = ":80"
  version = "1.0"
)

var (
  LocalIP = GetLocalIP()
  ExternalIP = GetExternalIP() 
  GeoLocation = GetGeoLocation(ExternalIP)
  CrashAppCounter = 0
)

type Specification struct {
    DisplayExternalIP bool `default:"False"`
    DisplayGeoLocation bool `default:"False"`
    CrashApp bool `default:"False"`
    CrashAppCount int `default:"5"`
    Debug bool `default:"False"`
    SimulateReady bool `default:"False"`
    WaitBeforeReady int `default:"30"`
}

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

func GetGeoLocation(IPAddress string) string {
    db, err := geoip2.Open("GeoLite2-City.mmdb")
    if err != nil {
      os.Stderr.WriteString(err.Error())
    }
    defer db.Close()
    // If you are using strings that may be invalid, check that ip is not nil
    ip := net.ParseIP(IPAddress)
    record, err := db.City(ip)
    if err != nil {
      os.Stderr.WriteString(err.Error())
    }
    LocationString := record.Subdivisions[0].Names["en"] + ", " + record.Country.Names["en"]

    return LocationString
}

func GetExternalIP() string {

  resp, err := http.Get("http://myexternalip.com/raw")

  if err != nil {
    os.Stderr.WriteString(err.Error())
    os.Exit(1)
  }

  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  return strings.TrimSpace(string(body[:]))
}

func HelloWorld(w http.ResponseWriter, r *http.Request, DisplayExternalIP bool, DisplayGeoLocation bool) {  
  fmt.Fprintf(w, "Hello, world! version: %s ", version)
  fmt.Fprintf(w, "Local IP: %s \n", LocalIP)
  if DisplayExternalIP ==true { fmt.Fprintf(w, "External IP: %s \n", ExternalIP)}
  if DisplayGeoLocation ==true { fmt.Fprintf(w, "Location: %s \n", GeoLocation) } 
  CrashAppCounter = CrashAppCounter + 1
}

func healthz(w http.ResponseWriter, r *http.Request, CrashApp bool, CrashAppCount int) {
  if CrashApp ==true && CrashAppCounter >= CrashAppCount {
    // do nothing
  } else {
    w.Write([]byte("OK"))
  }
}

func readiness(w http.ResponseWriter, r *http.Request, SimulateReady bool, WaitBeforeReady int) {
  if SimulateReady ==true {
    time.Sleep(time.Duration(WaitBeforeReady)*time.Second)
    } 
  w.Write([]byte("OK"))
}

func main() {
  
  var s Specification
  err := envconfig.Process("HelloWorld", &s)
  if err != nil {
    log.Fatal(err.Error())
  }

  if s.Debug ==true {
    format := "DisplayExternalIP: %v\nDisplayGeoLocation: %v\nCrashApp: %v\nCrashAppCount: %v\n"
    _, err = fmt.Printf(format, s.DisplayExternalIP, s.DisplayGeoLocation, s.CrashApp, s.CrashAppCount)
  }

  fmt.Printf("Started Application version: %s \n", version)
  fmt.Printf("Local IP: %s \n", LocalIP)  
  if s.DisplayExternalIP ==true { fmt.Printf("External IP: %s \n", ExternalIP) }
  if s.DisplayGeoLocation ==true { fmt.Printf("Location: %s \n", GeoLocation) } 
 
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    HelloWorld(w, r, s.DisplayExternalIP, s.DisplayGeoLocation)
  })
  
  http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
    healthz(w, r, s.CrashApp, s.CrashAppCount)
  })

  http.HandleFunc("/readiness", func(w http.ResponseWriter, r *http.Request) {
    readiness(w, r, s.SimulateReady, s.WaitBeforeReady)
  })

  http.ListenAndServe(port, nil)   

}
