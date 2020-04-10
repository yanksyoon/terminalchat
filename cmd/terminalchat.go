package main

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/charlie4284/terminalchat/pkg/color"
	terminalchat "github.com/charlie4284/terminalchat/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// Server - server type
type Server struct {
	terminalchat.TerminalchatServer
	Connections map[string]Connection
}

// Connection - describes connection
type Connection struct {
	Conn  terminalchat.Terminalchat_JoinServer
	Color string
	Err   chan error
}

func main() {
	srv := grpc.NewServer()
	srvInstance := &Server{}
	srvInstance.Connections = make(map[string]Connection)
	srvInstance.Connections["server"] = Connection{
		Conn:  nil,
		Color: color.Reset,
		Err:   nil,
	}
	terminalchat.RegisterTerminalchatServer(srv, srvInstance)
	reflection.Register(srv)

	listener, listenerErr := net.Listen("tcp", ":9999")
	if listenerErr != nil {
		panic(listenerErr)
	}

	fmt.Println("Server started")
	err := srv.Serve(listener)
	if err != nil {
		panic(err)
	}
	// Bind OS signal(terminate) to signal channel
	// stopChan := make(chan os.Signal)
	// signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT)
	defer srv.GracefulStop()
}

// Broadcast - Broadcast message to server
func (s *Server) Broadcast(ctx context.Context, in *terminalchat.Message) (*terminalchat.Empty, error) {
	for connUser, srv := range s.Connections {
		if connUser == "server" {
			continue
		}
		username := strings.TrimSpace(in.GetUsername())
		// fmt.Println(in.GetUsername(), in.GetMessage())
		clr := s.Connections[username].Color
		// coloredMsg := fmt.Sprintf("[%v %v %v] %v \n", clr, username, color.Reset, strings.TrimSpace(in.GetMessage()))
		// in.Message = coloredMsg
		newMsg := terminalchat.Message{
			Username: strings.TrimSpace(in.GetUsername()),
			Message:  fmt.Sprintf("[%v%v%v]: %v", clr, strings.TrimSpace(in.GetUsername()), color.Reset, strings.TrimSpace(in.GetMessage())),
		}
		if err := srv.Conn.Send(&newMsg); err != nil {
			srv.Err <- err
		}
	}
	return &terminalchat.Empty{}, nil
}

// Join - Join message server
func (s *Server) Join(in *terminalchat.User, joinserver terminalchat.Terminalchat_JoinServer) error {
	if s.Connections == nil {
		s.Connections = make(map[string]Connection)
	}
	if _, ok := s.Connections[in.GetUsername()]; ok {
		return status.Error(codes.AlreadyExists, "Current username is already taken")
	}

	newConn := Connection{
		Conn:  joinserver,
		Color: color.Random(),
		Err:   make(chan error),
	}
	s.Connections[strings.TrimSpace(in.GetUsername())] = newConn
	s.Broadcast(context.Background(), &terminalchat.Message{
		Username: "server",
		Message:  fmt.Sprintf("%v has joined the chat. Currently online: %v\n", in.GetUsername(), strings.Join(getCurrentConnections(s), ", ")),
	})
	err := <-newConn.Err
	delete(s.Connections, in.GetUsername())
	return err
}

func getCurrentConnections(s *Server) []string {
	userlist := []string{}
	for username := range s.Connections {
		userlist = append(userlist, username)
	}
	return userlist
}
