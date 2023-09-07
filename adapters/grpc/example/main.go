// Example below shows how to add Stanza fault tolerance guards
// to a simple net/http service.

package main

import (
	"context"
	"encoding/json"
	"net/http"

	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	quotev1 "github.com/StanzaSystems/sdk-go/adapters/grpc/example/gen/quote/v1"
	"github.com/StanzaSystems/sdk-go/stanza"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var (
	wg           sync.WaitGroup
	env          string
	debugLogging bool
	port         int
	srv          = grpc.NewServer()
	healthSrv    = health.NewServer()
	logOpts      = []logging.Option{logging.WithLogOnEvents(logging.FinishCall)}

	name      = "fiber-example" // TODO: add new service (this isn't "fiber")
	release   = "1.0.0"
	stanzaOpt = stanza.ClientOptions{
		// APIKey:   "my-api-key", // set here or in an STANZA_API_KEY environment variable
		Name:        name,
		Release:     release,
		Environment: env,

		// optionally prefetch Guard configs
		Guard: []string{"RootGuard", "ZenQuotes"},
	}
)

// For decoding ZenQuotes (https://zenquotes.io) JSON
var zq []struct {
	Q string
	A string
}

// Implement QuoteService gRPC server API
type QuoteServer struct {
	quotev1.UnimplementedQuoteServiceServer
	log *zap.Logger
}

func (qs *QuoteServer) GetQuote(ctx context.Context, req *quotev1.GetQuoteRequest) (*quotev1.GetQuoteResponse, error) {
	url := "https://zenquotes.io/api/random"
	resp, err := stanza.HttpGet(
		stanza.GuardRequest{
			Context: ctx,
			Name:    "ZenQuotes",
			URL:     url,
		})
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			json.NewDecoder(resp.Body).Decode(&zq)
			return &quotev1.GetQuoteResponse{
				Quote:  zq[0].Q,
				Author: zq[0].A,
				Source: url,
			}, nil
		}
	}
	return nil, status.Error(codes.Unavailable, "Service Unavailable")

	// // Name the Stanza Guard which protects this workflow
	// stz := stanza.Guard(ctx, "ZenQuotes")

	// // Check for and log any returned error messages
	// if stz.Error() != nil {
	// 	qs.log.Error("ZenQuotes", zap.Error(stz.Error()))
	// }

	// // 🚫 Stanza Guard has *blocked* this workflow, log the reason and return 429 status
	// if stz.Blocked() {
	// 	qs.log.Info(stz.BlockMessage(), zap.String("reason", stz.BlockReason()))
	// 	return nil, status.Error(codes.ResourceExhausted, stz.BlockMessage())
	// }

	// // ✅ Stanza Guard has *allowed* this workflow, business logic goes here.
	// url := "https://zenquotes.io/api/random"
	// if resp, _ := http.Get(url); resp != nil {
	// 	defer resp.Body.Close()

	// 	if resp.StatusCode == http.StatusOK {
	// 		json.NewDecoder(resp.Body).Decode(&zq)

	// 		// 😭 Sad path, rate limited by ZenQuotes
	// 		if zq[0].A == "zenquotes.io" {
	// 			// TODO: Add a secondary quote source here
	// 			return nil, status.Error(codes.ResourceExhausted, zq[0].Q)
	// 		}

	// 		// 🎉 Happy path, our "business logic" succeeded
	// 		stz.End(stz.Success)
	// 		return &quotev1.GetQuoteResponse{
	// 			Quote:  zq[0].Q,
	// 			Author: zq[0].A,
	// 			Source: url,
	// 		}, nil
	// 	}
	// }

	// // 😭 Sad path, our "business logic" failed
	// stz.End(stz.Failure)
	// return nil, status.Error(codes.Unavailable, "Service Unavailable")
}

func main() {
	// Parse command line flags
	flag.StringVar(&env, "environment", "dev", "Environment: for example, dev, staging, qa (default dev)")
	flag.IntVar(&port, "port", 3000, "Port to listen/accept requests on")
	flag.BoolVar(&debugLogging, "debug", true, "Debugging on/off")
	flag.Parse()

	// Create an interruptible context to use for graceful server shutdowns
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Configure Zap structured logger
	logger := newZapLogger(env, debugLogging)
	defer logger.Sync()

	// Init Stanza fault tolerance library
	stanzaExit, stanzaInitErr := stanza.Init(ctx, stanzaOpt)
	defer stanzaExit()
	if stanzaInitErr != nil {
		fmt.Printf("\n%s\n\n", stanzaInitErr.Error())
		os.Exit(-1)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		// stanza.NewGrpcGuard(ctx, "ZenQuotes")
		// guardOpts :=

		// Create new gRPC server with logging and Stanza Guard middleware
		srv = grpc.NewServer(
			grpc.ConnectionTimeout(5*time.Second),
			grpc.KeepaliveParams(keepalive.ServerParameters{MaxConnectionAge: 2 * time.Minute}),
			grpc.ChainStreamInterceptor(
				// stanza.StreamServerInterceptor(stanza.Guard(ctx, "ZenQuotes")),
				logging.StreamServerInterceptor(zapInterceptor(logger), logOpts...),
			),
			grpc.ChainUnaryInterceptor(
				stanza.UnaryServerInterceptor("RootGuard"),
				logging.UnaryServerInterceptor(zapInterceptor(logger), logOpts...),
				recovery.UnaryServerInterceptor(recoveryInterceptor(logger)),
			),
		)

		// Register gRPC services with server
		quotev1.RegisterQuoteServiceServer(srv, &QuoteServer{log: logger})
		grpc_health_v1.RegisterHealthServer(srv, healthSrv)
		reflection.Register(srv)

		// Start our example gRPC service
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			logger.Error("net.Listen", zap.Error(err))
			os.Exit(-1)
		}
		if err := srv.Serve(lis); err != nil {
			logger.Error("srv.Serve", zap.Error(err))
		}
	}()

	// GRACEFUL SHUTDOWN
	// - watches for a "Done" signal to the context we setup at the start
	// - triggered by os.Interrupt, syscall.SIGINT, or syscall.SIGTERM
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		srv.GracefulStop()
	}()

	wg.Wait()
	os.Exit(2)
}
