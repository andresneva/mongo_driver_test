package main

import (
	"strconv"

	"github.com/andresneva/mongo_driver_test/config"
	"github.com/andresneva/mongo_driver_test/http"
	"github.com/sirupsen/logrus"
)

func main() {

	appConfig := config.LoadConfig()

	handler := http.NewRequestHandler()

	server, err := http.ConfigureRoutes(handler, appConfig)
	if err != nil {
		logrus.Fatal(err)
	}

	server.Run(":" + strconv.Itoa(appConfig.Port))

}
