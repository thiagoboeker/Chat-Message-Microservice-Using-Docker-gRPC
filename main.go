package main

import (
	protos "github.com/chatapp/server/protos"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	glog "google.golang.org/grpc/grpclog"
	"log"
	"net"
	"os"
	"sync"
)

//Package level variables
var grpclog glog.LoggerV2 //The logger from grpc/grpclog

func init() {
	grpclog = glog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

//Conection - Struct to wrap some useful info
type Conection struct {
	Stream protos.Broadcast_CreateStreamServer
	XId    string
	Active bool
	Error  chan error
}

//Server - To gather the conection
type Server struct {
	Conections []*Conection
}

//CreateStream - Receive the request and create an Stream with clients
func (s *Server) CreateStream(r *protos.Conect, stream protos.Broadcast_CreateStreamServer) error {

	//The conection
	cnx := &Conection{
		Stream: stream,
		XId:    r.User.XId,
		Active: true,
		Error:  make(chan error),
	}

	//Gather in the conections pool
	s.Conections = append(s.Conections, cnx)

	//Wait for any error to return in the channel refering to this cnx object
	return <-cnx.Error

}

//BroadcastMessage - Gets the Message from the user
func (s *Server) BroadcastMessage(ctx context.Context, msg *protos.UserMessage) (*protos.Done, error) {

	mu := sync.WaitGroup{}
	done := make(chan int)

	//Loop through all connections
	for _, cnx := range s.Conections {
		mu.Add(1)
		go func(msg *protos.UserMessage, cnx *Conection) {
			defer mu.Done()
			if cnx.Active {
				//Send the message
				err := cnx.Stream.Send(msg)
				grpclog.Info("Sending message to ", cnx.Stream)
				if err != nil {
					grpclog.Errorf("Error with Stream: %v - Error: %v", cnx.Stream, err)
					cnx.Active = false
					cnx.Error <- err
				}
			}
		}(msg, cnx)
	}

	go func() {
		mu.Wait()
		close(done)
	}()

	<-done
	return &protos.Done{}, nil
}

func main() {

	var Conections []*Conection

	Srv := &Server{Conections}
	grpcServer := grpc.NewServer()

	li, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Error crating the server %v", err)
	}

	grpclog.Info("Starting server at port :8080")

	protos.RegisterBroadcastServer(grpcServer, Srv)
	grpcServer.Serve(li)
}
