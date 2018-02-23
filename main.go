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
	"strconv"
	"time"

	"github.com/LightsPlatform/Broker/group"
	"github.com/LightsPlatform/vSensor/sensor"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Config represents main configuration
var Config = struct {
	DB struct {
		URL string `default:"127.0.0.1" env:"db_url"`
	}
}{}

var groups map[string]*group.Group

func init() {
	groups = make(map[string]*group.Group)
}

var lightsDB *mgo.Database

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

	lightsDB = session.DB("lights")

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
	g.OnData = func(ds []sensor.Data) {
		c := lightsDB.C(r.Name)

		for _, d := range ds {
			c.Insert(d)
		}
	}
	go g.Run()
	groups[r.Name] = g

	c.JSON(http.StatusOK, g)
}

func groupDataHandler(c *gin.Context) {
	var results []bson.M

	id := c.Param("thingid")

	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := lightsDB.C(id).Find(bson.M{}).Limit(limit).All(&results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, results)
}

func groupSensorHandler(c *gin.Context) {
	id := c.Param("id")

	var r struct {
		Name string `json:"name" binding:"required"`
		URL  string `json:"url" binding:"required"`
	}

	if err := c.BindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	g, ok := groups[id]
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("There is no group named %s", id)})
	}
	g.Add(group.Sensor{
		ID:  r.Name,
		URL: r.URL,
	})

	c.JSON(http.StatusOK, g)

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
