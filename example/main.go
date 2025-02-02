package main

import (
	"d3framework"
	"fmt"
	"net"
	"net/http"
	"google.golang.org/grpc"
	"golang.org/x/net/websocket"
	"context"
	"log"
)

// CustomDataHandler handles data fetching
type CustomDataHandler struct{}

func (d *CustomDataHandler) FetchData(params map[string]string) (interface{}, error) {
	name := params["name"]
	if name == "" {
		name = "World"
	}
	return fmt.Sprintf("Hello, %s!", name), nil
}

// CustomInputHandler handles different types of inputs
type CustomInputHandler struct {
	d3framework.BaseInputHandler
}

func (i *CustomInputHandler) HandleHTTPInput(r *http.Request) (map[string]string, error) {
	params := map[string]string{
		"name": r.URL.Query().Get("name"),
	}
	return params, nil
}

func (i *CustomInputHandler) HandleTCPInput(conn net.Conn) (map[string]string, error) {
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}
	params := map[string]string{"message": string(buffer[:n])}
	return params, nil
}

func (i *CustomInputHandler) HandleWebSocketInput(conn *websocket.Conn) (map[string]string, error) {
	var message string
	if err := websocket.Message.Receive(conn, &message); err != nil {
		return nil, err
	}
	return map[string]string{"message": message}, nil
}

// CustomOutputHandler handles different types of outputs
type CustomOutputHandler struct {
	d3framework.BaseOutputHandler
}

func (o *CustomOutputHandler) SendHTTPResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, data)
}

func (o *CustomOutputHandler) SendTCPResponse(conn net.Conn, data interface{}) error {
	_, err := conn.Write([]byte(fmt.Sprintf("%v", data)))
	return err
}

func (o *CustomOutputHandler) SendWebSocketResponse(conn *websocket.Conn, data interface{}) error {
	return websocket.Message.Send(conn, data)
}

// CustomDeliveryHandler processes data
type CustomDeliveryHandler struct {
	d3framework.BaseDeliveryHandler
}

func (d *CustomDeliveryHandler) ProcessData(data interface{}) (interface{}, error) {
	return fmt.Sprintf("Processed Data: %v", data), nil
}

func startHTTPServer(framework *d3framework.Framework) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		params, _ := framework.Input.HandleHTTPInput(r)
		data, _ := framework.Data.FetchData(params)
		processedData, _ := framework.Delivery.ProcessData(data)
		framework.Output.SendHTTPResponse(w, processedData)
	})
	fmt.Println("Starting HTTP server on :8080")
	http.ListenAndServe(":8080", nil)
}

func startTCPServer(framework *d3framework.Framework) {
	listener, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("Failed to start TCP server: %v", err)
	}
	fmt.Println("Starting TCP server on :8081")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go func(conn net.Conn) {
			params, _ := framework.Input.HandleTCPInput(conn)
			data, _ := framework.Data.FetchData(params)
			processedData, _ := framework.Delivery.ProcessData(data)
			framework.Output.SendTCPResponse(conn, processedData)
			conn.Close()
		}(conn)
	}
}

func startWebSocketServer(framework *d3framework.Framework) {
	http.Handle("/ws", websocket.Handler(func(conn *websocket.Conn) {
		params, _ := framework.Input.HandleWebSocketInput(conn)
		data, _ := framework.Data.FetchData(params)
		processedData, _ := framework.Delivery.ProcessData(data)
		framework.Output.SendWebSocketResponse(conn, processedData)
	}))
	fmt.Println("Starting WebSocket server on :8082")
	http.ListenAndServe(":8082", nil)
}

func main() {
	framework := &d3framework.Framework{
		Data:    &CustomDataHandler{},
		Input:   &CustomInputHandler{},
		Output:  &CustomOutputHandler{},
		Delivery: &CustomDeliveryHandler{},
	}

	go startHTTPServer(framework)
	go startTCPServer(framework)
	go startWebSocketServer(framework)

	select {} // Keep main running
}
