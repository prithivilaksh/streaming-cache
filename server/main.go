package main

import (
	"context"
	"io"
	"net"
	"os"
	"strconv"

	pb "github.com/prithivilaksh/streaming-cache/proto/cache"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func InitRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Username: "default",
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	err := rdb.Ping(context.Background()).Err()

	if err != nil {
		panic(err)
	}

	return rdb
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

	pipe := s.rdb.TxPipeline()

	cmd := pipe.HGetAll(ctx, req.Tkr)
	_, err := pipe.Exec(ctx)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "redis error: %v", err)
	}

	res, err := cmd.Result()
	if err == redis.Nil {
		return nil, status.Errorf(codes.NotFound, "key not found")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "redis error: %v", err)
	}

	tkr := res["tkr"]
	timestamp, _ := strconv.ParseInt(res["timestamp"], 10, 64)
	price, _ := strconv.ParseFloat(res["price"], 64)
	volume, _ := strconv.ParseInt(res["volume"], 10, 64)

	tkrData := &pb.TkrData{
		Tkr:       tkr,
		Timestamp: timestamp,
		Price:     price,
		Volume:    volume,
	}

	return tkrData, nil
}

func (s *cacheServer) GetStream(req *pb.Tkr, stream pb.Cache_GetStreamServer) error {

	ctx := stream.Context()
	chnl := s.rdb.Subscribe(ctx, req.Tkr).Channel()

	getAndSend := func() error {
		res, err := s.Get(ctx, req)
		if err != nil {
			return err
		}
		if err := stream.Send(res); err != nil {
			return err
		}
		return nil
	}

	if err := getAndSend(); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done(): // client cancelled or disconnected
			return ctx.Err()

		case _, ok := <-chnl:
			if !ok { // subscription closed
				return nil
			}
			// only send if actual Redis update comes in
			if err := getAndSend(); err != nil {
				return err
			}
		}
	}
}

func (s *cacheServer) Set(ctx context.Context, req *pb.TkrData) (*pb.Ack, error) {

	pipe := s.rdb.TxPipeline()

	val := map[string]any{
		"tkr":       req.Tkr,
		"timestamp": req.Timestamp,
		"price":     req.Price,
		"volume":    req.Volume,
	}

	cmd1 := pipe.HSet(ctx, req.Tkr, val)
	cmd2 := pipe.Publish(ctx, req.Tkr, "updated")
	_, err := pipe.Exec(ctx)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "redis error: %v", err)
	}

	if cmd1.Err() != nil {
		return nil, status.Errorf(codes.Internal, "redis error: %v", cmd1.Err())
	}

	if cmd2.Err() != nil {
		return nil, status.Errorf(codes.Internal, "redis error: %v", cmd2.Err())
	}

	return &pb.Ack{Success: true}, nil
}

func (s *cacheServer) SetStream(stream pb.Cache_SetStreamServer) error {

	ctx := stream.Context()

	for {
		tkrData, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if _, err := s.Set(ctx, tkrData); err != nil {
			return err
		}
		if err := stream.Send(&pb.Ack{Success: true}); err != nil {
			return err
		}
	}
}

func main() {

	addr := os.Getenv("SERVER_ADDRESS")
	if addr == "" {
		addr = "localhost:50051" // default
	}
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()
	rdb := InitRedis()
	pb.RegisterCacheServer(grpcServer, NewCacheServer(rdb))
	grpcServer.Serve(lis)

}
