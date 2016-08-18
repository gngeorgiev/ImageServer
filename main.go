package main

import (
	"log"

	"github.com/gngeorgiev/ImageServer/imageServer"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	imageServer.Run()
}
