package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"AIGateway/internal/aiclient"
	"AIGateway/internal/config"
	"AIGateway/internal/db"
	"AIGateway/internal/repository"
	"AIGateway/internal/service"
	"AIGateway/internal/transport/grpc/pb"
	"AIGateway/internal/transport/grpc/server"
	"AIGateway/internal/transport/rest/handlers"
	"AIGateway/internal/worker"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

//Срок жизни ключа дедупликации в Redis
const dedupTTL = 24 * time.Hour

func main() {
	cfg := config.MustLoad()

	dbConn := db.Connect(cfg.Db)
	defer dbConn.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr})
	defer redisClient.Close()

	aiClient := aiclient.NewClient(cfg.AiClient)

	eventsRepo := repository.NewEvents(dbConn)
	enrichmentsRepo := repository.NewEnrichments(dbConn)

	eventsService := service.NewEventsService(eventsRepo, redisClient, dedupTTL)
	enrichmentsService := service.NewEnrichmentsService(enrichmentsRepo)

	//wg объединяет воркеров в Merge, gs ждет,
	//пока последний результат не запишется в БД при остановке
	wg := &sync.WaitGroup{}
	gs := &sync.WaitGroup{}
	jobs := make(chan worker.Job, cfg.Pipeline.BufferSize)

	pipeline := worker.NewPipeline(jobs, aiClient, wg, enrichmentsRepo, cfg.Pipeline.Count, cfg.Pipeline.BufferSize, gs)

	ctx, cancel := context.WithCancel(context.Background())
	pipeline.Start(ctx)

	eventsHandler := handlers.NewEventsHandler(eventsService, pipeline)

	router := gin.Default()
	router.POST("/events", eventsHandler.Post)
	router.GET("/events/:id", eventsHandler.GetEventById)

	restSrv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Rest.Host, cfg.Rest.Port),
		Handler:      router,
		ReadTimeout:  cfg.Rest.Timeout,
		WriteTimeout: cfg.Rest.Timeout,
	}

	go func() {
		log.Printf("rest server listening on %s", restSrv.Addr)
		if err := restSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("rest server error: %v", err)
		}
	}()

	grpcSrv := grpc.NewServer()
	pb.RegisterEnrichmentServiceServer(grpcSrv, server.NewEnrichmentServer(enrichmentsService))

	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Grpc.Port))
	if err != nil {
		log.Fatalf("grpc listen error: %v", err)
	}

	go func() {
		log.Printf("grpc server listening on %s", grpcLis.Addr())
		if err := grpcSrv.Serve(grpcLis); err != nil {
			log.Fatalf("grpc server error: %v", err)
		}
	}()

	//Ждем SIGINT/SIGTERM,
	//чтобы корректно остановить оба сервера и дождаться воркеров
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Rest.Timeout)
	defer shutdownCancel()

	if err := restSrv.Shutdown(shutdownCtx); err != nil {
		log.Println("rest shutdown err:", err)
	}
	grpcSrv.GracefulStop()

	//Новые Job уже не поступают,
	//закрываем канал, чтобы воркеры доработали очередь и завершились
	close(jobs)
	gs.Wait()

	cancel()
	log.Println("shutdown complete")
}
