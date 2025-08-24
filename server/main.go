package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"

	pb "github.com/prithivilaksh/streaming-cache/proto/cache"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func InitRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Username: "default",
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}

type cacheServer struct {
	pb.UnimplementedCacheServer
	rdb *redis.Client
}

func NewCacheServer(rdb *redis.Client) *cacheServer {
	return &cacheServer{
		rdb: rdb,
	}
}

func (s *cacheServer) Get(ctx context.Context, req *pb.Tkr) (*pb.TkrData, error) {
	return &pb.TkrData{}, nil
}

func (s *cacheServer) GetStream(req *pb.Tkr, stream pb.Cache_GetStreamServer) error {
	return nil
}

func (s *cacheServer) Set(ctx context.Context, req *pb.TkrData) (*pb.Ack, error) {
	return &pb.Ack{}, nil
}

func (s *cacheServer) SetStream(stream pb.Cache_SetStreamServer) error {
	return nil
}

var port = flag.String("port", "50051", "port to listen on")

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()
	rdb := InitRedis()
	pb.RegisterCacheServer(grpcServer, NewCacheServer(rdb))
	grpcServer.Serve(lis)

	// ctx := context.Background()

	// err := rdb.Set(ctx, "foo", "bar", 0).Err()
	// if err != nil {
	// 	panic(err)
	// }

	// result, err := rdb.Get(ctx, "foo").Result()

	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(result)

}
