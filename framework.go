package d3framework

import (
	"fmt"
	"net"
	"net/http"

	"golang.org/x/net/websocket"
	"google.golang.org/grpc"
)

// DataHandler is responsible for fetching and managing data
type DataHandler interface {
	FetchData(params map[string]string) (interface{}, error)
	HandleWebSocketInput(conn *websocket.Conn) (map[string]string, error)
	SendWebSocketResponse(conn *websocket.Conn, data interface{}) error
	HandleTCPInput(conn net.Conn) (map[string]string, error)
}

// InputHandler provides default implementations for input handling
type BaseInputHandler struct{}

func (b *BaseInputHandler) HandleHTTPInput(r *http.Request) (map[string]string, error) {
	return nil, fmt.Errorf("HTTP input handling not implemented")
}
func (b *BaseInputHandler) HandleJSONRPCInput(r *http.Request) (map[string]string, error) {
	return nil, fmt.Errorf("JSON-RPC input handling not implemented")
}
func (b *BaseInputHandler) HandleGRPCInput(stream grpc.ServerStream) (map[string]string, error) {
	return nil, fmt.Errorf("gRPC input handling not implemented")
}
func (b *BaseInputHandler) HandleWebSocketInput(conn *websocket.Conn) (map[string]string, error) {
	return nil, fmt.Errorf("WebSocket input handling not implemented")
}
func (b *BaseInputHandler) HandleTCPInput(conn net.Conn) (map[string]string, error) {
	return nil, fmt.Errorf("TCP input handling not implemented")
}

// OutputHandler provides default implementations for output handling
type BaseOutputHandler struct{}

func (b *BaseOutputHandler) SendHTTPResponse(w http.ResponseWriter, data interface{}) {
	http.Error(w, "HTTP response handling not implemented", http.StatusNotImplemented)
}
func (b *BaseOutputHandler) SendJSONRPCResponse(w http.ResponseWriter, data interface{}) {
	http.Error(w, "JSON-RPC response handling not implemented", http.StatusNotImplemented)
}
func (b *BaseOutputHandler) SendGRPCResponse(stream grpc.ServerStream, data interface{}) error {
	return fmt.Errorf("gRPC response handling not implemented")
}
func (b *BaseOutputHandler) SendWebSocketResponse(conn *websocket.Conn, data interface{}) error {
	return fmt.Errorf("WebSocket response handling not implemented")
}
func (b *BaseOutputHandler) SendTCPResponse(conn net.Conn, data interface{}) error {
	return fmt.Errorf("TCP response handling not implemented")
}

// DeliveryHandler provides default data processing
type BaseDeliveryHandler struct{}

func (b *BaseDeliveryHandler) ProcessData(data interface{}) (interface{}, error) {
	return data, nil
}

// DisplayHandler provides default display handling
type BaseDisplayHandler struct{}

func (b *BaseDisplayHandler) RenderData(data interface{}) (interface{}, error) {
	return data, nil
}
func (b *BaseDisplayHandler) Display(w http.ResponseWriter, data interface{}) {
	http.Error(w, "Display handling not implemented", http.StatusNotImplemented)
}

// Framework structure that binds the D3 components
type Framework struct {
	Data     DataHandler
	Input    InputHandler
	Output   OutputHandler
	Delivery DeliveryHandler
	Display  DisplayHandler
}

type InputHandler interface {
	FetchData(params map[string]string) (interface{}, error)
	HandleTCPInput(conn net.Conn) (map[string]string, error)
	HandleHTTPInput(r *http.Request) (map[string]string, error)
	HandleWebSocketInput(conn *websocket.Conn) (map[string]string, error)
}

type OutputHandler interface {
	SendHTTPResponse(w http.ResponseWriter, data interface{})
	SendTCPResponse(conn net.Conn, data interface{}) error
	SendWebSocketResponse(conn *websocket.Conn, data interface{}) error
}
type DeliveryHandler interface {
	ProcessData(data interface{}) (interface{}, error)
}
type DisplayHandler interface {
	RenderData(data interface{}) (interface{}, error)
	Display(w http.ResponseWriter, data interface{})
}

// HTTPServer handles HTTP-specific requests
func (f *Framework) HTTPServer(addr string) {
	http.HandleFunc("/", f.ServeHTTP)
	fmt.Printf("Starting HTTP server on %s\n", addr)
	http.ListenAndServe(addr, nil)
}

// GRPCServer handles gRPC-specific requests
func (f *Framework) GRPCServer(addr string, server *grpc.Server) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Failed to start gRPC server: %v\n", err)
		return
	}
	fmt.Printf("Starting gRPC server on %s\n", addr)
	server.Serve(listener)
}
func (f *Framework) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params, _ := f.Input.HandleHTTPInput(r)
	data, _ := f.Data.FetchData(params)
	processedData, _ := f.Delivery.ProcessData(data)
	f.Output.SendHTTPResponse(w, processedData)
}

// WebSocketServer handles WebSocket-specific connections
func (f *Framework) WebSocketServer(addr string) {
	http.Handle("/ws", websocket.Handler(func(conn *websocket.Conn) {
		params, _ := f.Input.HandleWebSocketInput(conn)
		data, _ := f.Data.FetchData(params)
		processedData, _ := f.Delivery.ProcessData(data)
		f.Output.SendWebSocketResponse(conn, processedData)
	}))
	fmt.Printf("Starting WebSocket server on %s\n", addr)
	http.ListenAndServe(addr, nil)
}

// TCPServer handles TCP-specific requests
func (f *Framework) TCPServer(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Failed to start TCP server: %v\n", err)
		return
	}
	fmt.Printf("Starting TCP server on %s\n", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\n", err)
			continue
		}
		go func(conn net.Conn) {
			params, _ := f.Input.HandleTCPInput(conn)
			data, _ := f.Data.FetchData(params)
			processedData, _ := f.Delivery.ProcessData(data)
			f.Output.SendTCPResponse(conn, processedData)
			conn.Close()
		}(conn)
	}
}
