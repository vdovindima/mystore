package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
)

func main() {
	var socketPath string
	var pathFolder string

	flag.StringVar(&socketPath, "socket", "./socket", "Path to unix socket.")
	flag.StringVar(&pathFolder, "pathFolder", "./data", "path folder file storage")
	flag.Parse()

	if socketPath == "" {
		flag.Usage()
		return
	}
	if pathFolder == "" {
		flag.Usage()
		return
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Println(err.Error())
		return
	}

	// By default, unix socket would only be available to same user.
	// If we want access it from Nginx, we need to loosen permissions.
	err = os.Chmod(socketPath, fs.ModePerm)
	if err != nil {
		log.Println(err)
		return
	}

	httpServer := http.Server{
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			switch request.Method {
			case http.MethodPost:
				// Parse our multipart form, 10 << 20 specifies a maximum
				// upload of 10 MB files.
				request.ParseMultipartForm(1024)
				file, header, err := request.FormFile("file")
				if err != nil {
					http.Error(writer, err.Error(), http.StatusBadRequest)
					return
				}
				defer file.Close()

				data, err := io.ReadAll(file)
				if err != nil {
					http.Error(writer, err.Error(), http.StatusInternalServerError)
					return
				}

				err = os.WriteFile(filepath.Join(pathFolder, request.URL.String(), header.Filename), data, 0644)
				if err != nil {
					http.Error(writer, err.Error(), http.StatusInternalServerError)
					return
				}

				fmt.Fprintf(writer, "File uploaded successfully")

			case http.MethodPut:
				fileName := filepath.Base(request.URL.Path)
				file, err := os.OpenFile(filepath.Join(pathFolder, request.URL.String(), fileName), os.O_WRONLY|os.O_TRUNC, 0644)
				if err != nil {
					http.Error(writer, err.Error(), http.StatusInternalServerError)
					return
				}
				defer file.Close()

				data, err := io.ReadAll(request.Body)
				if err != nil {
					http.Error(writer, err.Error(), http.StatusInternalServerError)
					return
				}

				_, err = file.Write(data)
				if err != nil {
					http.Error(writer, err.Error(), http.StatusInternalServerError)
					return
				}

				fmt.Fprintf(writer, "File updated successfully")

			case http.MethodDelete:
				err := os.Remove(filepath.Join(pathFolder, request.URL.String()))
				if err != nil {
					http.Error(writer, err.Error(), http.StatusInternalServerError)
					return
				}

				fmt.Fprintf(writer, "Deleted successfully")
			default:
				writer.WriteHeader(http.StatusMethodNotAllowed)
				if _, err := writer.Write([]byte("Method not allowed")); err != nil {
					log.Println(err.Error())
				}
			}
		}),
	}

	// Setting up graceful shutdown to clean up Unix socket.
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP Server Shutdown Error: %v", err)
		}
	}()

	log.Printf("Service is listening on socket file %s", socketPath)
	err = httpServer.Serve(listener)
	if err != nil {
		log.Println(err.Error())
		return
	}
}
