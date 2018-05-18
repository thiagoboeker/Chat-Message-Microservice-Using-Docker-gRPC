package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	protos "github.com/chatapp/server/protos"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	"log"
	"os"
	"sync"
	"time"
)

//Package level variables
var client protos.BroadcastClient
var mu *sync.WaitGroup

func init() {
	mu = &sync.WaitGroup{}
}

//Conect - Conect to the system
func Conect(user *protos.User) error {

	//Variable to persist and return the error
	var StreamError error

	//The stream object from the Client RPC
	stream, err := client.CreateStream(context.Background(), &protos.Conect{
		User:   user,
		Active: true,
	})

	if err != nil {
		return fmt.Errorf("Conection failed :%v", err)
	}

	mu.Add(1)
	go func(str protos.Broadcast_CreateStreamClient) {
		defer mu.Done()
		//Loops forever receiving messages from the stream
		for {
			msg, err := str.Recv()
			if err != nil {
				StreamError = fmt.Errorf("Error reading message: %v", err)
				break
			}
			fmt.Printf("%v : %v\n", msg.XId, msg.Content)
		}
	}(stream)

	return StreamError
}

func main() {

	//Current time to add in the crypto and create an unique ID
	timeStamp := time.Now()
	done := make(chan int)
	name := flag.String("name", "User Undefined", "The name of the user")
	ID := sha256.Sum256([]byte(timeStamp.String() + *name))
	flag.Parse()

	//grpc conection
	conn, err := grpc.Dial("192.168.99.100:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Couldnt conect to service: %v", err)
	}

	//Registry the client
	client = protos.NewBroadcastClient(conn)

	//Create the User
	User := &protos.User{
		XId:  hex.EncodeToString(ID[:]),
		Name: *name,
	}

	//Conect to server
	Conect(User)

	//Code to get the input in the command line
	mu.Add(1)
	go func() {
		defer mu.Done()
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			//The message object to send in the RPC
			msg := &protos.UserMessage{
				XId:     User.XId,
				Content: scanner.Text(),
				Time:    timeStamp.String(),
			}

			//Client RPC
			_, err := client.BroadcastMessage(context.Background(), msg)
			if err != nil {
				fmt.Printf("Error Sending Message:%v", err)
				break
			}
		}

	}()

	go func() {
		mu.Wait()
		close(done)
	}()

	<-done
}
