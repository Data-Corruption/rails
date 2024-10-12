package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

const CloseTimeout = 3 * time.Second

type Server struct {
	shutdownSignal chan os.Signal // channel for shutdown signals
	router         *chi.Mux
	server         *http.Server
	port           string // e.g. ":80"
}

func (s *Server) Start() {
	port := 9393
	s.port = fmt.Sprintf(":%d", port)

	// Initialize the server
	var err error
	s.router = NewRouter()

	// configure the server
	s.server = &http.Server{
		Addr:    s.port,
		Handler: s.router,
	}

	// server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// listen for shutdown signals
	s.shutdownSignal = make(chan os.Signal, 1)
	signal.Notify(s.shutdownSignal, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-s.shutdownSignal // wait for shutdown signal

		// shutdown signal with grace period
		shutdownCtx, cancel := context.WithTimeout(serverCtx, CloseTimeout)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				panic("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// trigger graceful shutdown
		if err = s.server.Shutdown(shutdownCtx); err != nil {
			panic(err)
		}
		serverStopCtx()
	}()

	waitForServer := true

	// attempt to start the server, if port is in use, increment port and try again
	for attempt := 0; attempt < 10; attempt++ {
		listener, err := net.Listen("tcp", s.port)
		if err == nil {
			link := fmt.Sprintf("http://localhost%s", s.port)
			fmt.Println("Emulator started at:", link)

			// Start the server in a new goroutine
			go func() {
				if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
					fmt.Println("Error serving:", err)
				}
			}()
			break
		} else {
			// If error is port in use, try next port
			opErr, ok := err.(*net.OpError)
			if ok {
				sysErr, ok := opErr.Err.(*os.SyscallError)
				if ok && sysErr.Err == syscall.EADDRINUSE {
					// port is in use
					port++
					s.port = fmt.Sprintf(":%d", port)
					s.server.Addr = s.port
					continue
				}
			}
			// Other errors
			fmt.Println("Error starting server:", err)
			waitForServer = false
			break
		}
	}

	if waitForServer {
		<-serverCtx.Done()
	}
}
