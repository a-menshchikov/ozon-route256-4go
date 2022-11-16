package reports

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"strings"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model/expense/reports/api"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type listener struct {
	api.UnimplementedReporterServer

	subscribers map[types.User]chan types.Report

	listener net.Listener
	logger   *zap.Logger
}

func NewListener(grpcCfg config.ReportsGrpcConfig, l *zap.Logger) (*listener, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcCfg.ServerPort))
	if err != nil {
		return nil, errors.Wrap(err, "starting gRPC server")
	}

	l.Debug("listen report gRPC connection")

	return &listener{
		subscribers: make(map[types.User]chan types.Report),
		listener:    lis,
		logger:      l,
	}, nil
}

func (l *listener) Close() {
	if err := l.listener.Close(); err != nil {
		l.logger.Error("cannot close gRPC server connection", zap.Error(err))
	}
}

func (l *listener) Run() {
	server := grpc.NewServer(grpc.UnaryInterceptor(metricsInterceptor))
	api.RegisterReporterServer(server, l)

	l.logger.Info("start report gRPC server")

	go func() {
		if err := server.Serve(l.listener); err != nil && err != grpc.ErrServerStopped {
			l.logger.Error("failed to serve gRPC connection", zap.Error(err))
		}
	}()
}

func (l *listener) SendReport(ctx context.Context, in *api.Report) (*emptypb.Empty, error) {
	if err := in.Validate(); err != nil {
		l.logger.Warn("invalid report", zap.Error(err))
		return &emptypb.Empty{}, nil
	}

	var span opentracing.Span
	if md, ok := metadata.FromIncomingContext(ctx); !ok {
		span = opentracing.StartSpan("handle report")
	} else {
		reader := strings.NewReader(md["x-trace"][0])
		decoder := base64.NewDecoder(base64.RawStdEncoding, reader)
		if spanCtx, err := opentracing.GlobalTracer().Extract(opentracing.Binary, decoder); err != nil {
			l.logger.Error("cannot extract trace from metadata", zap.Error(err), zap.Any("metadata", md))
			span = opentracing.StartSpan("handle report")
		} else {
			span = opentracing.GlobalTracer().StartSpan("handle report", ext.RPCServerOption(spanCtx))
		}
	}
	defer span.Finish()

	logger := l.logger.With(zap.Int64("userID", in.User.Id))
	logger.Debug("handle report call")

	user := types.User(in.User.Id)
	if subscriber, ok := l.subscribers[user]; ok {
		subscriber <- types.Report{
			Data:    in.Data,
			Success: in.Success,
			Error:   in.Error,
		}
		close(subscriber)
	} else {
		logger.Warn("no subscriber for report")
	}

	return &emptypb.Empty{}, nil
}

func (l *listener) Subscribe(user *types.User) <-chan types.Report {
	l.subscribers[*user] = make(chan types.Report)

	return l.subscribers[*user]
}

func (l *listener) Unsubscribe(user *types.User) {
	delete(l.subscribers, *user)
}
