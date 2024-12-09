package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	pb "src/proto"

	"github.com/urfave/cli/v2"

	cp "github.com/atotto/clipboard"
	"golang.design/x/clipboard"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type server struct {
	pb.UnimplementedClipboardServiceServer
}

func (s server) SendClipboard(srv pb.ClipboardService_SendClipboardServer) error {
	isRecieved := false
	mutex := &sync.Mutex{}
	ch := clipboard.Watch(context.TODO(), clipboard.FmtText)
	go func() {
		for data := range ch {
			if !isRecieved {
				req := pb.Response{Characters: string(data)}
				if err := srv.Send(&req); err != nil {
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
	}()

	log.Println("start new server")
	ctx := srv.Context()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := srv.Recv()
		if err == io.EOF {
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
		err = cp.WriteAll(req.Characters)
		if err != nil {
			fmt.Println("error in writing to clipboard", err)
			continue
		}
	}
}

func main() {
	app := &cli.App{
		Name:  "gRPC Clipboard Server",
		Usage: "A server to send and receive clipboard data via gRPC",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "port",
				Usage:    "Port to run the gRPC server on",
				Value:    "50005",
				Required: false,
			},
			&cli.BoolFlag{
				Name:  "insecure",
				Usage: "Use insecure gRPC connection",
				Value: true,
			},
			&cli.StringFlag{
				Name:  "cert",
				Usage: "Path to the TLS certificate file",
			},
			&cli.StringFlag{
				Name:  "key",
				Usage: "Path to the TLS key file",
			},
		},
		Action: func(c *cli.Context) error {
			port := c.String("port")
			insecureConn := c.Bool("insecure")
			certFile := c.String("cert")
			keyFile := c.String("key")
			log.Println(certFile, keyFile)

			// create listener
			log.SetFlags(log.LstdFlags | log.Lshortfile)

			lis, err := net.Listen("tcp", ":"+port)
			if err != nil {
				log.Fatalf("failed to listen: %v", err)
			}

			// create grpc server
			var opts []grpc.ServerOption
			if insecureConn {
				// opts = append(opts, grpc.Creds(insecure.NewCredentials()))
			} else {
				if certFile == "" || keyFile == "" {
					log.Fatalf("cert and key files must be provided for TLS")
				}
				creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
				if err != nil {
					log.Fatalf("failed to load TLS credentials: %v", err)
				}
				opts = append(opts, grpc.Creds(creds))
			}

			s := grpc.NewServer(opts...)
			pb.RegisterClipboardServiceServer(s, &server{})

			// and start...
			if err := s.Serve(lis); err != nil {
				log.Fatalf("failed to serve: %v", err)
			}
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
