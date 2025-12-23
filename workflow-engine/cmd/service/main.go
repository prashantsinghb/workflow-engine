package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	service "github.com/prashantsinghb/workflow-engine/api/service"
	"github.com/prashantsinghb/workflow-engine/pkg/config"
	"github.com/prashantsinghb/workflow-engine/pkg/execution/postgres"
	moduleregistry "github.com/prashantsinghb/workflow-engine/pkg/module/registry"
	pb "github.com/prashantsinghb/workflow-engine/pkg/proto"
	"github.com/prashantsinghb/workflow-engine/pkg/server"
	"github.com/prashantsinghb/workflow-engine/pkg/workflow/executor"
	wfregistry "github.com/prashantsinghb/workflow-engine/pkg/workflow/registry"
)

func main() {
	/* -------------------- CONFIG -------------------- */

	cfg := config.Load()

	/* -------------------- DB -------------------- */

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	/* -------------------- STORES -------------------- */

	store := postgres.New(db)
	executionStore := store.Executions()
	nodeStore := store.Nodes()
	eventStore := store.Events()
	workflowStore := wfregistry.NewPostgresWorkflowStore(db)

	/* -------------------- MODULE REGISTRY -------------------- */

	modulePgRegistry := moduleregistry.NewPostgresRegistry(db)
	moduleRegistry := moduleregistry.NewModuleRegistry(modulePgRegistry)

	/* -------------------- gRPC SERVER -------------------- */

	grpcServer := grpc.NewServer()

	stepRegistry := wfregistry.NewLocalStepRegistry()
	pbServer := server.NewStepRegistryServer(stepRegistry, moduleRegistry)
	pb.RegisterStepRegistryServer(grpcServer, pbServer)

	// Register executors with module registry
	grpcExecutor := executor.NewRemoteExecutor(moduleRegistry)
	httpExecutor := executor.NewRemoteExecutor(moduleRegistry)
	executor.Register("grpc", grpcExecutor)
	executor.Register("http", httpExecutor)

	workflowSvc := server.NewWorkflowService(
		executionStore,
		nodeStore,
		eventStore,
		workflowStore,
		moduleRegistry,
	)

	service.RegisterWorkflowServiceServer(grpcServer, workflowSvc)

	moduleSvc := &server.ModuleServer{
		Registry: *moduleRegistry,
	}
	service.RegisterModuleServiceServer(grpcServer, moduleSvc)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		log.Println("gRPC server started on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	/* -------------------- HTTP (gRPC-Gateway) -------------------- */

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	gatewayMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	if err := service.RegisterWorkflowServiceHandlerFromEndpoint(
		ctx,
		gatewayMux,
		"localhost:50051",
		opts,
	); err != nil {
		log.Fatalf("failed to register workflow gateway: %v", err)
	}

	if err := service.RegisterModuleServiceHandlerFromEndpoint(
		ctx,
		gatewayMux,
		"localhost:50051",
		opts,
	); err != nil {
		log.Fatalf("failed to register module gateway: %v", err)
	}

	/* -------------------- HTTP Router (Chi) -------------------- */

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// CORS middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Register timeline endpoint
	timelineServer := server.NewExecutionTimelineServer(store)
	router.Get("/v1/projects/{projectId}/executions/{executionId}/timeline", timelineServer.GetExecutionTimeline)

	// Mount grpc-gateway for all other routes
	router.Mount("/", gatewayMux)

	httpServer := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	go func() {
		log.Println("HTTP server started on :8081")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	/* -------------------- SHUTDOWN -------------------- */

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig

	log.Println("shutting down servers...")

	grpcServer.GracefulStop()
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	httpServer.Shutdown(ctxTimeout)
}
