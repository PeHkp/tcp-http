package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"servidor-tcp/routes"
	"time"
)

var (
	ErrInvalidJSON = errors.New("json inválido")
	ErrBadRoute    = errors.New("rota não encontrada")
	ErrServerClosed = errors.New("servidor encerrado")
)

type Server struct {
	listenAddr string
	ln         net.Listener
	quitch     chan struct{}
	msgch      chan []byte
	maxConn    chan struct{} 
}

type Request struct {
	URL     string                 `json:"url"`
	Content map[string]interface{} `json:"content"`
}

func InitServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr, 
		quitch:     make(chan struct{}), 
		msgch:      make(chan []byte, 10),
		maxConn:    make(chan struct{}, 100), 
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return fmt.Errorf("erro ao iniciar servidor: %w", err)
	}
	fmt.Println("Server Started", ln.Addr())
	defer ln.Close()
	s.ln = ln

	go s.processMessages()
	
	go s.acceptConnections()

	<-s.quitch

	return nil
}

func (s *Server) Stop() error {
	close(s.quitch)
	if s.ln != nil {
		return s.ln.Close()
	}
	return nil
}

func (s *Server) processMessages() {
	for msg := range s.msgch {
		fmt.Println("Processando mensagem:", string(msg))
	}
}

func (s *Server) acceptConnections() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			select {
			case <-s.quitch:
				return 
			default:
				fmt.Println("Erro ao aceitar conexão:", err)
				continue
			}
		}

		fmt.Println("Nova conexão ao servidor:", conn.RemoteAddr())

		select {
		case s.maxConn <- struct{}{}:
			go func(c net.Conn) {
				defer func() { <-s.maxConn }() 
				s.readContentFromRequest(c)
			}(conn)
		default:
			
			fmt.Println("Limite de conexões atingido, rejeitando:", conn.RemoteAddr())
			sendResponse(conn, []byte("Servidor ocupado, tente novamente mais tarde\n"))
			conn.Close()
		}
	}
}

func (s *Server) readContentFromRequest(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	
	for {
		msg, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Conexão fechada pelo cliente:", conn.RemoteAddr())
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("Timeout de leitura:", conn.RemoteAddr())
				sendResponse(conn, []byte("Timeout de conexão\n"))
			} else {
				fmt.Println("Erro de leitura:", err)
			}
			break
		}
		
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		
		fmt.Println(string(msg))
		s.msgch <- msg

		req, err := readJson(msg)
		if err != nil {
			sendResponse(conn, []byte(fmt.Sprintf("Erro: %s\n", err.Error())))
			continue
		}

		if req.URL == "/quit" {
			sendResponse(conn, []byte("Conexão encerrada pelo servidor\n"))
			break
		}

		route, err := routes.Routes(req.URL)
		if err != nil {
			sendResponse(conn, []byte(fmt.Sprintf("Erro: %s\n", err.Error())))
			continue
		}

		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		sendResponse(conn, []byte(route+"\n"))
	}
}

func readJson(msg []byte) (*Request, error) {
	var request Request
	err := json.Unmarshal(msg, &request)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidJSON, err.Error())
	}
	
	if request.URL == "" {
		return nil, errors.New("URL não pode ser vazia")
	}
	
	return &request, nil
}

func sendResponse(conn net.Conn, msg []byte) {
	_, err := conn.Write(msg)
	if err != nil {
		fmt.Println("Erro ao enviar resposta:", err)
	}
}

