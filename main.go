/*
 * Power control block
 *
 * This API was created to monitor states of Power Control Block and send some commands to it.
 *
 * API version: 1.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"./control"
	sw "./go"
	"github.com/natefinch/lumberjack"
	// WARNING!
	// Change this to a fully-qualified import path
	// once you place this file into your project.
	// For example,
	//
	//    sw "github.com/myname/myrepo/go"
	//
)

func main() {
	f, errf := os.OpenFile("output.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if errf != nil {
		log.Fatalf("Error open log file: %v", errf)
	}
	defer f.Close()

	log.SetOutput(&lumberjack.Logger{
		Filename:   "output.log",
		MaxSize:    1,  // megabytes after which new file is created
		MaxBackups: 3,  // number of backups
		MaxAge:     28, // days
	})

	log.Printf("Hello programm")

	// control.HttpTest()

	time.Sleep(time.Second)
	control.InitProtocol()

	control.ProcStart()

	log.Printf("Server started")
	router := sw.NewRouter()
	http.ListenAndServe(":8080", router)
}
