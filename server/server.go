package server

import (
	"encoding/json"
	"fmt"
	"net"
	"servidor-tcp/routes"
)

type Server struct {
	listenAddr string
	ln         net.Listener
	quitch     chan struct{}
	msgch      chan []byte
}

type Request struct {
	URL     string                 `json:"url"`
	Content map[string]interface{} `json:"content"`
}

func InitServer(listenAddr string) *Server {
	return &Server{listenAddr: listenAddr, quitch: make(chan struct{}), msgch: make(chan []byte, 10)}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	fmt.Println("Server Started", ln.Addr())
	defer ln.Close()
	s.ln = ln

	go s.acceptConnections()

	<-s.quitch

	return nil
}

func (s *Server) acceptConnections() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}

		fmt.Println("New conection to the server: ", conn.RemoteAddr())

		go s.readContentFromRequest(conn)
	}
}

func (s *Server) readContentFromRequest(conn net.Conn) {
	for {
		buf := make([]byte, 2048)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("read error:", err)
			return
		}
		msg := buf[:n]
		s.msgch <- msg

		convertToJsonFormat := readJson(msg)

		if convertToJsonFormat == nil {
			sendResponse(conn, []byte("Bad request\n"))
			continue
		}
		route := routes.Routes(convertToJsonFormat.URL)

		sendResponse(conn, []byte(route))
	}
}

func readJson(msg []byte) *Request {
	var request Request
	err := json.Unmarshal(msg, &request)
	if err != nil {
		fmt.Println("Erro ao decodificar JSON:", err)
		return nil
	}
	return &request
}

func sendResponse(conn net.Conn, msg []byte) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("write error:", err)
		return
	}
}
