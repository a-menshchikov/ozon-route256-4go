package reports

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/Shopify/sarama"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/message"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/model/expense/reports/api"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/storage"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type consumer struct {
	kafkaBrokers       []string
	kafkaConsumerGroup string
	kafkaTopic         string
	saramaConfig       *sarama.Config

	grpcAddress    string
	grpcConnection io.Closer
	grpcClient     api.ReporterClient

	storage storage.ExpenseStorage
	rater   model.Rater

	logger *zap.Logger
}

func NewConsumer(kafkaCfg config.ReportsKafkaConfig, grpcCfg config.ReportsGrpcConfig, s storage.ExpenseStorage, r model.Rater, l *zap.Logger) (*consumer, error) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Version = sarama.V2_5_0_0
	saramaCfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaCfg.Consumer.Offsets.Retention = kafkaCfg.Timeout

	switch kafkaCfg.Assignor {
	case "sticky":
		saramaCfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategySticky}
	case "round-robin":
		saramaCfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRoundRobin}
	case "range", "":
		saramaCfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.BalanceStrategyRange}
	default:
		return nil, fmt.Errorf("unrecognized consumer group partition assignor: %s", kafkaCfg.Assignor)
	}

	return &consumer{
		kafkaBrokers:       kafkaCfg.Brokers,
		kafkaConsumerGroup: kafkaCfg.ConsumerGroup,
		kafkaTopic:         kafkaCfg.Topic,
		saramaConfig:       saramaCfg,

		grpcAddress: grpcCfg.ClientAddress,

		storage: s,
		rater:   r,

		logger: l,
	}, nil
}

func (c *consumer) Run(ctx context.Context) error {
	consumerGroup, err := sarama.NewConsumerGroup(c.kafkaBrokers, c.kafkaConsumerGroup, c.saramaConfig)
	if err != nil {
		return errors.Wrap(err, "creating consumer group")
	}

	c.logger.Info("start report consumer")

	if err := consumerGroup.Consume(ctx, []string{c.kafkaTopic}, c); err != nil {
		return errors.Wrap(err, "report consuming")
	}

	return nil
}

func (c *consumer) Setup(sarama.ConsumerGroupSession) error {
	conn, err := grpc.Dial(c.grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return errors.Wrap(err, "cannot connect to gRPC server")
	}

	c.grpcConnection = conn
	c.grpcClient = api.NewReporterClient(conn)

	return nil
}

func (c *consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return c.grpcConnection.Close()
}

func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for m := range claim.Messages() {
		go func(m *sarama.ConsumerMessage) {
			var span opentracing.Span
			if spanCtx, err := extractSpanFromHeaders(m.Headers); err != nil {
				c.logger.Error("cannot extract trace from headers", zap.Error(err))
				span = opentracing.StartSpan("consume")
			} else {
				span = opentracing.GlobalTracer().StartSpan("consume", ext.RPCServerOption(spanCtx))
			}

			ctx := opentracing.ContextWithSpan(session.Context(), span)
			defer span.Finish()

			c.handleMessage(ctx, m)
			session.MarkMessage(m, "")
		}(m)
	}

	return nil
}

func (c *consumer) handleMessage(ctx context.Context, saramaMessage *sarama.ConsumerMessage) {
	userID, err := strconv.ParseInt(string(saramaMessage.Key), 10, 64)
	if err != nil {
		c.logger.Error("cannot parse user ID from report message", zap.Error(err))
		return
	}

	logger := c.logger.With(zap.Int64("userID", userID))
	logger.Debug("handle message")

	report := &api.Report{
		User: &api.User{Id: userID},
	}
	defer func() {
		c.sendReport(ctx, report)
		if report.Success {
			_consumedCount.WithLabelValues("ok").Inc()
		} else {
			_consumedCount.WithLabelValues("error").Inc()
		}
	}()

	if !c.rater.TryAcquireExchange() {
		return
	}
	defer c.rater.ReleaseExchange()

	var reportMessage message.Report
	if err := json.Unmarshal(saramaMessage.Value, &reportMessage); err != nil {
		logger.Warn("cannot unmarshal incoming message", zap.Error(err), zap.ByteString("message", saramaMessage.Value))
		return
	}

	user := types.User(userID)
	expenses, err := c.storage.List(ctx, &user, reportMessage.From)
	if err != nil {
		logger.Error("ExpenseStorage.List", zap.Error(err))
		return
	}

	data := make(map[string]int64)

	for category := range expenses {
		for _, item := range expenses[category] {
			amount, err := c.rater.Exchange(ctx, item.Amount, item.Currency, reportMessage.Currency, item.Date)
			if err != nil {
				logger.Error("cannot exchange currency", zap.Error(err), zap.String("from", item.Currency), zap.String("to", reportMessage.Currency))
				return
			}

			data[category] += amount
		}
	}

	report.Data = data
	report.Success = true
}

func (c *consumer) sendReport(ctx context.Context, report *api.Report) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		traceBuffer := new(bytes.Buffer)
		spanContext := span.Context()
		if err := span.Tracer().Inject(spanContext, opentracing.Binary, traceBuffer); err != nil {
			c.logger.Error("failed to inject span ID", zap.Error(err))
		} else {
			encodedTrace := new(bytes.Buffer)
			encoder := base64.NewEncoder(base64.RawStdEncoding, encodedTrace)
			_, err := encoder.Write(traceBuffer.Bytes())
			if err != nil {
				c.logger.Error("failed to encode trace binary", zap.Error(err))
			} else if err := encoder.Close(); err != nil {
				c.logger.Error("cannot close trace encoder", zap.Error(err))
			} else {
				s := encodedTrace.String()
				ctx = metadata.AppendToOutgoingContext(ctx, "x-trace", s)
			}
		}
	}

	_, err := c.grpcClient.SendReport(ctx, report)
	if err != nil {
		c.logger.Error("cannot send report", zap.Error(err), zap.Int64("userID", report.User.Id))
	}
}

func extractSpanFromHeaders(headers []*sarama.RecordHeader) (opentracing.SpanContext, error) {
	for _, header := range headers {
		switch string(header.Key) {
		case "x-trace":
			return opentracing.GlobalTracer().Extract(opentracing.Binary, bytes.NewReader(header.Value))
		}
	}

	return nil, nil
}
