/*
 * +===============================================
 * | Author:        Parham Alvani <parham.alvani@gmail.com>
 * |
 * | Creation Date: 18-02-2018
 * |
 * | File Name:     main.go
 * +===============================================
 */

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/LightsPlatform/Broker/group"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
)

// Config represents main configuration
var Config = struct {
	DB struct {
		URL string `default:"127.0.0.1" env:"db_url"`
	}
	Lights struct {
		Sensors   string `default:"127.0.0.1:8080" env:"lights_sensors"`
		Actuators string `default:"127.0.0.1" env:"lights_actuators"`
	}
}{}

var groups map[string]*group.Group

func init() {
	groups = make(map[string]*group.Group)
}

// handle registers apis and create http handler
func handle() http.Handler {
	r := gin.Default()

	api := r.Group("/api")
	{
		api.GET("/about", aboutHandler)

		api.POST("/group", groupCreateHandler)
		api.POST("/group/:id", groupSensorHandler)
		api.GET("/group/:id", groupDataHandler)
		api.GET("/group", groupListHandler)
		api.DELETE("/group/:id", groupDeleteHandler)
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "404 Not Found"})
	})

	return r
}

func main() {
	// Load configuration
	if err := configor.Load(&Config, "config.yml"); err != nil {
		panic(err)
	}

	// Create a Mongo Session
	session, err := mgo.Dial(Config.DB.URL)
	if err != nil {
		log.Fatalf("Mongo session %s: %v", Config.DB.URL, err)
	}
	defer session.Close()
	log.Infof("Mongo session %s has been created\n", Config.DB.URL)

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	fmt.Println("vBroker Light @ 2018")

	srv := &http.Server{
		Addr:    ":1375",
		Handler: handle(),
	}

	go func() {
		fmt.Printf("vBroker Listen: %s\n", srv.Addr)
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("Listen Error:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	fmt.Println("vBroker Shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Shutdown Error:", err)
	}
}

func aboutHandler(c *gin.Context) {
	c.String(http.StatusOK, "18.20 is leaving us")
}

func groupCreateHandler(c *gin.Context) {
	var r struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.BindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	g := group.New(r.Name)
	groups[r.Name] = g

	c.JSON(http.StatusOK, g)
}

func groupDataHandler(c *gin.Context) {
}

func groupSensorHandler(c *gin.Context) {
}

func groupListHandler(c *gin.Context) {
	gs := make([]*group.Group, len(groups))

	i := 0
	for _, g := range groups {
		gs[i] = g
		i++
	}

	c.JSON(http.StatusOK, gs)
}

func groupDeleteHandler(c *gin.Context) {
}
