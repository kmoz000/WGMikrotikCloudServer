package main

import (
	"log"

	"github.com/joho/godotenv"
	WGMCS "github.com/kmoz000/wgmikrotikcontrolserver/WGController"
)

func main() {
	log.SetPrefix("WGMCS#:")
	if err := godotenv.Load(); err != nil {
		log.Fatalf("godotenv %s", err.Error())
	}
	wgmcs := WGMCS.CloudServer{}
	if err := wgmcs.GenDevices(); err != nil {
		log.Fatalf("godotenv %s", err.Error())
	}
	wgmcs.UpDevice()
}
