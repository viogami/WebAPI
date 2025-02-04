package main

import (
	"WebAPI/server"
	"log"
)

func main() {
	// 读取配置
	if err := server.InitConfig("config.yaml"); err != nil {
		log.Fatalf("Config error: %v", err)
	}

	s := server.NewServer(server.AppConfig)
	s.RegisterRoutes()
	s.Run()
}
