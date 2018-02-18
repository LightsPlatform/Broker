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
	"log"

	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
	mgo "gopkg.in/mgo.v2"
)

// Config represents main configuration
var Config = struct {
	DB struct {
		URL string `default:"127.0.0.1" env:"db_url"`
	}
}{}

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
}
