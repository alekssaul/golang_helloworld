//go:generate statik -src=../frontend

package main

import (
	"log"
	"net/http"

	_ "github.com/alekssaul/golang_helloworld/server/statik"
	"github.com/kelseyhightower/envconfig"
	"github.com/rakyll/statik/fs"
)

const (
	version = "3.1"
)

var (
	CrashAppCounter = 0
)

type EnvVars struct {
	DisplayExternalIP  bool   `default:"True"`
	DisplayGeoLocation bool   `default:"True"`
	CrashAppCount      int    `default:"5"`
	SimulateReady      bool   `default:"False"`
	WaitBeforeReady    int    `default:"30"`
	Port               string `default:"80"`
}

type Machine struct {
	ExternalIP  string
	LocalIP     string
	GeoLocation local
	version     string
}

type local struct {
	State     string
	Country   string
	Latitude  float64
	Longitude float64
}

func main() {

	var env EnvVars
	var machine Machine

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		healthz(w, r)
	})

	err := envconfig.Process("HelloWorld", &env)
	if err != nil {
		log.Fatal(err.Error())
	}

	http.HandleFunc("/kill", func(w http.ResponseWriter, r *http.Request) {
		KillServer(w, r, env.CrashAppCount)
	})

	log.Printf("- Running with Flags - \nDISPLAYEXTERNALIP: %v\nDISPLAYGEOLOCATION: %v\nCrashAppCount: %v\nPort: %v\n",
		env.DisplayExternalIP, env.DisplayGeoLocation, env.CrashAppCount, env.Port)

	machine.version = version
	log.Printf("Started Application version: %s \n", machine.version)
	machine.LocalIP = GetLocalIP()
	log.Printf("Local IP: %s \n", machine.LocalIP)

	if env.DisplayExternalIP {
		machine.ExternalIP, err = GetExternalIP()
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Printf("External IP: %s \n", machine.ExternalIP)
	}

	if env.DisplayGeoLocation {
		machine.GeoLocation, err = GetGeoLocation(machine.ExternalIP)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Printf("Location: %s , %s \n", machine.GeoLocation.State, machine.GeoLocation.Country)
	}

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", http.StripPrefix("/", http.FileServer(statikFS)))

	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		MainApi(w, r, env, machine)
	})

	http.HandleFunc("/readiness", func(w http.ResponseWriter, r *http.Request) {
		readiness(w, r, env.SimulateReady, env.WaitBeforeReady)
	})

	log.Fatal(http.ListenAndServe(":"+env.Port, nil))

}
