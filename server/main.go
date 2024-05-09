package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	pb "src/proto"
	"sync"
	"time"

	cp "github.com/atotto/clipboard"
	"golang.design/x/clipboard"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedMathServer
}

func (s server) Max(srv pb.Math_MaxServer) error {
	isRecieved := false
	mutex := &sync.Mutex{}
	ch := clipboard.Watch(context.TODO(), clipboard.FmtText)
	go func() {
		for data := range ch {
			if !isRecieved {
				req := pb.Response{String_: string(data)}
				if err := srv.Send(&req); err != nil {
					log.Printf("can not send %v\n", err)
					continue
				}
				// log.Printf("from gnome to lkde [%s] sent", req.String_)
				time.Sleep(time.Millisecond * 200)
			} else {
				mutex.Lock()
				isRecieved = false
				mutex.Unlock()
			}
		}
	}()

	log.Println("start new server")
	// var max = "a"
	ctx := srv.Context()

	for {

		// exit if context is done
		// or continue
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// receive data from stream
		req, err := srv.Recv()
		if err == io.EOF {
			// return will close stream from server side
			log.Println("exit")
			return nil
		}
		if err != nil {
			log.Printf("receive error %v", err)
			continue
		}

		mutex.Lock()
		isRecieved = true
		mutex.Unlock()

		err = cp.WriteAll(req.String_)
		if err != nil {
			fmt.Println("error in writing to clipboard", err)
			continue
		}

		// continue if number reveived from stream
		// less than max
		// if req.String_ == max {
		// 	continue
		// }

		// // update max and send it to stream
		// max = req.String_
		// now := time.Now().String()
		// resp := pb.Response{String_: max + "CHANGED at" + now}
		// if err := srv.Send(&resp); err != nil {
		// 	log.Printf("send error %v", err)
		// }
		// log.Printf("send new max=%s", max)
	}
}

func main() {
	// create listener
	lis, err := net.Listen("tcp", ":50005")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// create grpc server
	s := grpc.NewServer()

	pb.RegisterMathServer(s, &server{})

	// and start...
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
