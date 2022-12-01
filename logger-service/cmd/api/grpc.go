package main

import (
	"context"
	"fmt"
	"github.com/dorukbulut/log-service/data"
	"github.com/dorukbulut/log-service/logs"
	"google.golang.org/grpc"
	"log"
	"net"
)

type LogServer struct {
	logs.UnimplementedLogServiceServer
	Models data.Models
}

func (l *LogServer) WriteLog(ctx context.Context, req *logs.LogRequest) (*logs.LogResponse, error) {
	input := req.GetEntry()

	// write the log
	logEntry := data.LogEntry{
		Name: input.Name,
		Data: input.Data,
	}

	err := l.Models.LogEntry.Insert(logEntry)
	if err != nil {
		res := &logs.LogResponse{Result: "failed"}
		return res, err
	}

	//return response
	res := &logs.LogResponse{Result: "logged via gRPC"}
	return res, nil
}

func (app *Config) gRPCListen() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", GRPC))

	if err != nil {
		log.Fatalf("Failed to listen for gRPC : %v", err)
	}

	srv := grpc.NewServer()
	logs.RegisterLogServiceServer(srv, &LogServer{Models: app.Models})

	log.Printf("gRPC Server started on port %s", GRPC)

	if err := srv.Serve(lis); err != nil {
		log.Fatalf("Failed to listen for gRPC : %v", err)
	}
}
