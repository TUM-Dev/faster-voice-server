package main

import "faster-voice-service/pkg/server"

func main() {
	s := server.NewServer()
	s.Run()
}
