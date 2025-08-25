package main

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/prithivilaksh/streaming-cache/proto/cache"
)

// func get(cacheClient pb.CacheClient, tkr string) {
// 	ctx := context.Background()
// 	res, err := cacheClient.Get(ctx, &pb.Tkr{Tkr: tkr})
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Println(res)
// }

func getStream(cacheClient pb.CacheClient) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	stream, err := cacheClient.GetStream(ctx, &pb.Tkr{Tkr: "GOOGL"})
	if err != nil {
		panic(err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("Server closed the stream")
			return
		}
		if err != nil {
			// if context timed out/canceled, it will show up here
			fmt.Println("Stream error:", err)
			return
		}
		fmt.Println(res)
	}
}

// func set(cacheClient pb.CacheClient, tkrData *pb.TkrData) {

// }

func setStream(cacheClient pb.CacheClient) {

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	stream, err := cacheClient.SetStream(ctx)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context timed out, closing stream")
			stream.CloseSend() // optional, signals server
			return
		default:
			tkrData := &pb.TkrData{
				Tkr:       "GOOGL",
				Timestamp: time.Now().Unix(),
				Price:     100 + rand.Float64()*4,
				Volume:    100000000 - int64(rand.Int32N(1000000)),
			}
			if err := stream.Send(tkrData); err != nil {
				fmt.Println("Send error:", err)
				return
			}

			ack, err := stream.Recv()
			if err == io.EOF {
				fmt.Println("Server closed stream")
				return
			}
			if err != nil {
				fmt.Println("Recv error:", err)
				return
			}
			fmt.Println("Set ack received from server:", ack.Success)
			time.Sleep(time.Duration(rand.Int32N(10)) * time.Second)
		}
	}
}

func main() {

	addr := os.Getenv("SERVER_ADDRESS")
	if addr == "" {
		addr = "localhost:50051" // default
	}
	grpcClient, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	defer grpcClient.Close()

	cacheClient := pb.NewCacheClient(grpcClient)

	// get(cacheClient, "GOOGL")
	// set(cacheClient)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go setStream(cacheClient)
	go getStream(cacheClient)
	wg.Wait()

}
