package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"servidor-tcp/server"
	"syscall"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	srv := server.InitServer(":3000")
	
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	
	errCh := make(chan error, 1)
	go func() {
		fmt.Println("Servidor iniciado na porta 3000")
		errCh <- srv.Start()
	}()
	
	select {
	case err := <-errCh:
		if err != nil {
			log.Printf("Erro ao iniciar o servidor: %v\n", err)
		}
	case sig := <-sigCh:
		log.Printf("Recebido sinal: %v. Iniciando encerramento...\n", sig)
		
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		if err := srv.Stop(); err != nil {
			log.Printf("Erro ao encerrar o servidor: %v\n", err)
		}
		
		<-ctx.Done()
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Timeout de encerramento excedido. Forçando encerramento.")
		}
	}
	
	log.Println("Servidor encerrado")
}
