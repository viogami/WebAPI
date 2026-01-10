package main

import (
	"webapi/conf"
	"webapi/server"
	"log"
)

func main() {
	if err := conf.InitConfig("conf/config.yaml"); err != nil {
		log.Fatalf("Config error: %v", err)
	}

	s := server.NewServer(conf.AppConfig)

	s.UseMiddleware()

	s.RegisterRoutes()

	s.Run()
}
