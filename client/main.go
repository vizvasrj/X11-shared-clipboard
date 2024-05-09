package main

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	pb "src/proto"

	cp "github.com/atotto/clipboard"
	"golang.design/x/clipboard"
	"google.golang.org/grpc"
)

func main() {
	// dail server
	conn, err := grpc.Dial(":50005", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}

	// create stream
	client := pb.NewMathClient(conn)
	stream, err := client.Max(context.Background())
	if err != nil {
		log.Fatalf("openn stream error %v", err)
	}

	var max string
	ctx := stream.Context()
	done := make(chan bool)

	isRecieved := false
	mutex := &sync.Mutex{}
	// first goroutine sends random increasing numbers to stream
	// and closes it after 10 iterations
	ch := clipboard.Watch(context.TODO(), clipboard.FmtText)
	go func() {
		for data := range ch {
			if !isRecieved {
				req := pb.Request{String_: string(data)}
				if err := stream.Send(&req); err != nil {
					log.Printf("can not send %v\n", err)
					continue
				}
				// log.Printf("%s sent", req.String_)
				time.Sleep(time.Millisecond * 200)
			} else {
				// fmt.Println("isRecieved data that is i am not running send clipboard because i just recieved")
				mutex.Lock()
				isRecieved = false
				mutex.Unlock()
			}
		}
		if err := stream.CloseSend(); err != nil {
			log.Println(err)
		}
	}()

	// second goroutine receives data from stream
	// and saves result in max variable
	//
	// if stream is finished it closes done channel
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			max = resp.String_
			// log.Printf("recieved Clipboard [%s]", max)
			mutex.Lock()
			isRecieved = true
			mutex.Unlock()
			// clipboard.Write(clipboard.FmtText, []byte(max))
			cp.WriteAll(max)
		}
	}()

	// third goroutine closes done channel
	// if context is done
	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil {
			log.Println(err)
		}
		close(done)
	}()

	<-done
	// log.Printf("finished with max=%s", max)

}
