package main

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

const (
	Port             = 8091                // HTTP server port
	Delay            = 5 * time.Second     // Delay upon connection establishment
	GracefulShutdown = "graceful shutdown" // command to shutdown server
)

var Stop = make(chan struct{})

func main() {

	portstring := ":" + strconv.Itoa(Port)
	webserver := &http.Server{Addr: portstring}
	http.HandleFunc("/", hashValue)
	listener, err := net.Listen("tcp", "127.0.0.1"+portstring)
	if err != nil {
		panic(err)
	}

	// Main web server
	go func() {
		if err := webserver.Serve(listener); err != http.ErrServerClosed {
			log.Fatal("Can't start web server listener:  ", err)
		}
	}()

	<-Stop

	// Prevent any more connections
	listener.Close()

	// Let remaining connections be processed to completion
	ctx, _ := context.WithTimeout(context.Background(), 24*time.Hour)
	webserver.Shutdown(ctx)
}

func hashValue(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		time.Sleep(Delay)

		err := request.ParseForm()
		if err == nil {

			// Look for shutdown command or first non-zero length value
			for k := range request.Form {
				if k == GracefulShutdown {
					fmt.Println("graceful shutdown")
					close(Stop)
					fmt.Fprint(writer, GracefulShutdown)
					break
				} else {
					value := request.Form[k][0]
					if len(value) > 0 {
						h := sha512.New()
						h.Write([]byte(value))
						sha1_hash := base64.URLEncoding.EncodeToString(h.Sum(nil))
						fmt.Fprint(writer, sha1_hash)
						break
					}
				}
			}
		}
	}
}
