package main

import (
    "fmt"
    "io"
    "net/http"
    "sync"

    "golang.org/x/net/websocket"
)

type Server struct {
    mu sync.Mutex
    conns map[*websocket.Conn]struct{}
}

func (s *Server) addConn(conn *websocket.Conn) {
    s.mu.Lock()
    defer s.mu.Unlock()

    s.conns[conn] = struct{}{}
}

func (s *Server) removeConn(conn *websocket.Conn) {
    s.mu.Lock()
    defer s.mu.Unlock()

    delete(s.conns, conn)
}

//could process each conn write in a separate goroutine for better
//delivery speed
func (s *Server) broadcast(msg []byte) {
    s.mu.Lock()
    defer s.mu.Unlock()

    for conn := range s.conns {
        conn.Write(msg) 
    }
}

func (s *Server) startReadLoop(conn *websocket.Conn) {
    buf := make([]byte, 1024)
    for {
        n, err := conn.Read(buf) 
        if err == io.EOF {
            fmt.Println("Connection ended")
            s.removeConn(conn)
            break
        }
        if err != nil {
            msg := []byte(fmt.Sprintf("error: %s", err)) 
            if _, err := conn.Write(msg); err != nil {
                fmt.Println(err.Error())
            }
            continue
        }

        go s.broadcast(buf[:n])
    }
}

func (s *Server) Handle(conn *websocket.Conn) {
    fmt.Println("new conn: ", conn.RemoteAddr())
    s.addConn(conn) 
    s.startReadLoop(conn)
}

func NewServer() *Server {
    return &Server{
        conns: make(map[*websocket.Conn]struct{}),
    }
}

func main() {
    fmt.Println("starting server")
    server := NewServer()
    http.Handle("/ws", websocket.Handler(server.Handle))

    http.ListenAndServe(":1337", nil)
}
