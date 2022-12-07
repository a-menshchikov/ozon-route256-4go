package reports

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/Shopify/sarama"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/config"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/dto/message"
	"gitlab.ozon.dev/almenschhikov/go-course-4/internal/types"
	"go.uber.org/zap"
)

type producer struct {
	topic          string
	saramaProducer sarama.SyncProducer
	logger         *zap.Logger
}

func NewProducer(kafkaCfg config.ReportsKafkaConfig, l *zap.Logger) (*producer, error) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Version = sarama.V2_8_0_0
	saramaCfg.Producer.RequiredAcks = sarama.WaitForAll

	if saramaCfg.Producer.Idempotent {
		saramaCfg.Producer.Retry.Max = 1
		saramaCfg.Net.MaxOpenRequests = 1
	}

	saramaCfg.Producer.Return.Successes = true

	_ = saramaCfg.Producer.Partitioner

	saramaProducer, err := sarama.NewSyncProducer(kafkaCfg.Brokers, saramaCfg)
	if err != nil {
		return nil, errors.Wrap(err, "starting sarama sync producer")
	}

	return &producer{
		topic:          kafkaCfg.Topic,
		saramaProducer: saramaProducer,
		logger:         l,
	}, nil
}

func (p *producer) Send(ctx context.Context, user *types.User, from time.Time, currency string) error {
	value, err := json.Marshal(message.Report{
		From:     from,
		Currency: currency,
	})
	if err != nil {
		return errors.Wrap(err, "cannot marshal report message")
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(user.String()),
		Value: sarama.ByteEncoder(value),
	}

	if span := opentracing.SpanFromContext(ctx); span != nil {
		buf := new(bytes.Buffer)
		if err := span.Tracer().Inject(span.Context(), opentracing.Binary, buf); err != nil {
			p.logger.Error("failed to inject span ID", zap.Error(err))
		} else {
			msg.Headers = []sarama.RecordHeader{{
				Key:   []byte("x-trace"),
				Value: buf.Bytes(),
			}}
		}
	}

	_, _, err = p.saramaProducer.SendMessage(msg)
	if err != nil {
		return errors.Wrap(err, "failed to send report message to kafka")
	}

	producedCount.WithLabelValues(currency).Inc()

	return nil
}

func (p *producer) Close() {
	if err := p.saramaProducer.Close(); err != nil {
		p.logger.Error("failed to close sarama producer", zap.Error(err))
	}
}
