package server

import (
	"context"

	"AIGateway/internal/service"
	"AIGateway/internal/transport/grpc/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//Реализует pb.EnrichmentServiceServer,
//поэтому обязательно встраиваем Unimplemented
type EnrichmentServer struct {
	pb.UnimplementedEnrichmentServiceServer
	Enrichments service.Enrichments
}

func NewEnrichmentServer(enrichments service.Enrichments) *EnrichmentServer {
	return &EnrichmentServer{Enrichments: enrichments}
}

func (s *EnrichmentServer) GetEnrichment(ctx context.Context, req *pb.GetEnrichmentRequest) (*pb.GetEnrichmentResponse, error) {
	//В proto id int64, а в модели обычный int,
	//поэтому приводим тип перед вызовом сервиса
	e, err := s.Enrichments.GetById(ctx, int(req.Id))
	if err != nil {
		//В gRPC нет HTTP-статусов,
		//поэтому ошибку оборачиваем в status.Error с нужным кодом
		return nil, status.Error(codes.NotFound, err.Error())
	}

	//Переносим поля модели в ответ вручную,
	//CreatedAt конвертируем через timestamppb.New
	return &pb.GetEnrichmentResponse{
		Id:        int64(e.Id),
		EventId:   int64(e.EventId),
		Response:  e.Response,
		Model:     e.Model,
		CreatedAt: timestamppb.New(e.CreatedAt),
	}, nil
}