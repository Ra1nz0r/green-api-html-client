package main

import (
	"green-api-html-client/internal/server"

	_ "green-api-html-client/docs"
)

// @title GREEN-API HTML Client
// @version 1.0
// @description HTML client and backend proxy for GREEN-API methods getSettings, getStateInstance, sendMessage and sendFileByUrl.
// @contact.name Artem Rylskii
// @contact.url https://t.me/Rainz0r
// @contact.email n52rus@gmail.com
// @host localhost:8080
// @BasePath /
func main() {
	server.Run()
}
