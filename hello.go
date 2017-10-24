package main

import (
  "fmt"
  "net/http"
  "net"
  "log"
  "io/ioutil"
  "strings"
  "github.com/oschwald/geoip2-golang"
  "github.com/kelseyhightower/envconfig"
  "time"
)

const (
  version = "3.1"
)

var (
  CrashAppCounter = 0
  GeoLocation = ""
  ExternalIP = ""
  port = ":80"
)

type EnvVars struct {
    DisplayExternalIP bool `default:"False"`
    DisplayGeoLocation bool `default:"False"`
    CrashApp bool `default:"False"`
    CrashAppCount int `default:"5"`
    Debug bool `default:"False"`
    SimulateReady bool `default:"False"`
    WaitBeforeReady int `default:"30"`
    Port string 
}

type Machine struct {
  ExternalIP string
  LocalIP string
  GeoLocation string
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

func GetGeoLocation(IPAddress string) (IP string, err error) {
    db, err := geoip2.Open("GeoLite2-City.mmdb")
    if err != nil {
      return "", err
    }
    defer db.Close()
    // If you are using strings that may be invalid, check that ip is not nil
    ip := net.ParseIP(IPAddress)
    record, err := db.City(ip)
    if err != nil {
      return "", err
    }
    LocationString := record.Subdivisions[0].Names["en"] + ", " + record.Country.Names["en"]

    return LocationString, nil
}

func GetExternalIP() (ExternalIP string, err error){

  resp, err := http.Get("http://myexternalip.com/raw")

  if err != nil {
    return "", err
  }

  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  return strings.TrimSpace(string(body[:])), err
}

func HelloWorld(w http.ResponseWriter, r *http.Request, env EnvVars, machine Machine) {  
  fmt.Fprintf(w, "Hello, world! version: %s ", version)
  fmt.Fprintf(w, "Local IP: %s \n", machine.LocalIP)
  if env.DisplayExternalIP { fmt.Fprintf(w, "External IP: %s \n", machine.ExternalIP)}
  if env.DisplayGeoLocation { fmt.Fprintf(w, "Location: %s \n", machine.GeoLocation) } 
  CrashAppCounter = CrashAppCounter + 1
}

func healthz(w http.ResponseWriter, r *http.Request, CrashApp bool, CrashAppCount int) {
  if CrashApp && CrashAppCounter >= CrashAppCount {
    // do nothing
  } else {
    w.Write([]byte("OK"))
  }
}

func readiness(w http.ResponseWriter, r *http.Request, SimulateReady bool, WaitBeforeReady int) {
  if SimulateReady {
    time.Sleep(time.Duration(WaitBeforeReady)*time.Second)
    } 
  w.Write([]byte("OK"))
}

func main() {
  
  var env EnvVars
  var machine Machine

  err := envconfig.Process("HelloWorld", &env)
  if err != nil {
    log.Fatal(err.Error())
  }

  if env.Port != "" {port = env.Port}
  
  if env.Debug {
    fmt.Printf("DisplayExternalIP: %v\nDisplayGeoLocation: %v\nCrashApp: %v\nCrashAppCount: %v\nPort: %s\n", 
      env.DisplayExternalIP, env.DisplayGeoLocation, env.CrashApp, env.CrashAppCount, port)
  }

  log.Printf("Started Application version: %s \n", version)
  machine.LocalIP = GetLocalIP()
  log.Printf("Local IP: %s \n", machine.LocalIP) 

  if env.DisplayExternalIP {
    machine.ExternalIP, err = GetExternalIP()
    if err != nil {
      log.Fatal(err.Error())
    }   
    log.Printf("External IP: %s \n", machine.ExternalIP )
  }
  if env.DisplayGeoLocation {
    machine.GeoLocation, err = GetGeoLocation(machine.ExternalIP)
    if err != nil {
      log.Fatal(err.Error())
    }     
    log.Printf("Location: %s \n", machine.GeoLocation) 
  } 
 
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    HelloWorld(w, r, env, machine)
  })
  
  http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
    healthz(w, r, env.CrashApp, env.CrashAppCount)
  })

  http.HandleFunc("/readiness", func(w http.ResponseWriter, r *http.Request) {
    readiness(w, r, env.SimulateReady, env.WaitBeforeReady)
  })

  err = http.ListenAndServe(port, nil)
  if err != nil {
    log.Fatal(err.Error())
  }

}
