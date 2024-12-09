package main

import (
	"context"
	"io"
	"log"
	"os"
	"sync"
	"time"

	pb "src/proto"

	"github.com/urfave/cli/v2"

	cp "github.com/atotto/clipboard"
	"golang.design/x/clipboard"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	app := &cli.App{
		Name:  "gRPC Clipboard Client",
		Usage: "A client to send and receive clipboard data via gRPC",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "ip",
				Usage: "IP address of the gRPC server",
				Value: "localhost:50005",
			},
			&cli.BoolFlag{
				Name:  "insecure",
				Usage: "Use insecure gRPC connection",
				Value: true,
			},
			&cli.StringFlag{
				Name:     "tls",
				Usage:    "Path to TLS certificate",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "group",
				Usage:    "Group name",
				Required: false,
			},
		},
		Action: func(c *cli.Context) error {
			ip := c.String("ip")
			insecureConn := c.Bool("insecure")

			// Set up gRPC connection
			var opts []grpc.DialOption
			if insecureConn {
				opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
			} else {
				// Add your secure credentials here
				if c.String("tls") == "" {
					log.Fatalf("TLS certificate required")
				} else {
					creds, err := credentials.NewClientTLSFromFile(c.String("tls"), "")
					if err != nil {
						log.Fatalf("could not load tls cert: %s", err)
					}
					opts = append(opts, grpc.WithTransportCredentials(creds))
				}
			}

			conn, err := grpc.NewClient(ip, opts...)
			if err != nil {
				log.Fatalf("can not connect with server %v", err)
			}
			defer conn.Close()

			// create stream
			client := pb.NewMathClient(conn)
			stream, err := client.Max(context.Background())
			if err != nil {
				log.Fatalf("open stream error %v", err)
			}

			var max string
			ctx := stream.Context()
			done := make(chan bool)

			isRecieved := false
			mutex := &sync.Mutex{}

			// first goroutine sends clipboard data to stream
			ch := clipboard.Watch(context.TODO(), clipboard.FmtText)
			go func() {
				for data := range ch {
					if !isRecieved {
						req := pb.Request{String_: string(data), Group: c.String("group")}
						if err := stream.Send(&req); err != nil {
							log.Printf("can not send %v\n", err)
							continue
						}
						time.Sleep(time.Millisecond * 200)
					} else {
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
					mutex.Lock()
					isRecieved = true
					mutex.Unlock()
					cp.WriteAll(max)
				}
			}()

			// third goroutine closes done channel if context is done
			go func() {
				<-ctx.Done()
				if err := ctx.Err(); err != nil {
					log.Println(err)
				}
				close(done)
			}()

			<-done
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
