package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"math/rand/v2"
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
	ctx, cancel := context.WithTimeout(ctx, time.Second*40)
	defer cancel()

	stream, err := cacheClient.GetStream(ctx, &pb.Tkr{Tkr: "GOOGL"})
	if err != nil {
		panic(err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		fmt.Println(res)
	}
}

// func set(cacheClient pb.CacheClient, tkrData *pb.TkrData) {

// }

func setStream(cacheClient pb.CacheClient) {

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*40)
	defer cancel()

	stream, err := cacheClient.SetStream(ctx)
	if err != nil {
		panic(err)
	}

	for {
		tkrData := &pb.TkrData{
			Tkr:       "GOOGL",
			Timestamp: time.Now().Unix(),
			Price:     rand.Float64() * 100,
			Volume:    rand.Int64(),
		}
		if err := stream.Send(tkrData); err != nil {
			panic(err)
		}
		ack, err := stream.Recv()
		if err != nil {
			panic(err)
		}
		fmt.Println(ack.Success)
		time.Sleep(time.Duration(rand.Int32N(10)) * time.Second)
	}
}

var serverAddr = flag.String("addr", "localhost:50051", "The server address in the format of host:port")

func main() {
	flag.Parse()

	grpcClient, err := grpc.NewClient(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	defer grpcClient.Close()

	cacheClient := pb.NewCacheClient(grpcClient)

	// get(cacheClient, "GOOGL")
	getStream(cacheClient)
	// set(cacheClient)
	setStream(cacheClient)
}
