package main

import (
	"go_im_ftt/mqtt/models"
	"log"
	"net"
	"os"
)

func main() {
	f, err := os.OpenFile("log.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		return
	}
	log.SetOutput(f)
	l, err := net.Listen("tcp", "localhost:1883")
	if err != nil {
		log.Print("listen: ", err)
		return
	}
	svr := models.NewServer(l)
	svr.Start()
	<-svr.Done
}
