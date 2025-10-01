package box

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/thefabric-io/fieldslog"
	"github.com/thefabric-io/transactional"
)

type fifoSubscriber struct {
	transactional      transactional.Transactional
	name               string
	box                Box
	handlers           map[string]Handler
	handlersLock       sync.RWMutex
	processedCount     int
	batchSize          int
	waitTime           time.Duration
	waitTimeIfMessages time.Duration
}

func NewFIFOSubscriber(tx transactional.Transactional, name string, box Box, batchSize int, waitTime time.Duration, waitTimeIfMessages time.Duration) Subscriber {
	return &fifoSubscriber{
		box:                box,
		name:               name,
		transactional:      tx,
		handlers:           make(map[string]Handler),
		batchSize:          batchSize,
		waitTime:           waitTime,
		waitTimeIfMessages: waitTimeIfMessages,
	}
}

func (es *fifoSubscriber) DefaultLogFields() map[string]any {
	_, file, _, _ := runtime.Caller(0)

	return map[string]any{
		"metadata": map[string]any{
			"file": file,
		},
		"subject": map[string]any{
			"type":           "fifoSubscriber",
			"implements":     "Subscriber",
			"name":           es.name,
			"batchSize":      es.batchSize,
			"waitTime":       es.waitTime,
			"processedCount": es.processedCount,
		},
		"dependencies": map[string]any{
			"transactional": es.transactional.DefaultLogFields(),
			"box":           es.box.DefaultLogFields(),
		},
	}
}

func (es *fifoSubscriber) Start(ctx context.Context) error {
	fieldslog.Info(es, "fifoSubscriber started")

	initTx, err := es.transactional.BeginTransaction(ctx, transactional.DefaultWriteTransactionOptions())
	if err != nil {
		fieldslog.Error(es, "beginning transaction failed", err)

		return err
	}

	if err := es.box.Init(ctx, initTx); err != nil {
		fieldslog.Error(es, "error while initializing box", err)

		if rbErr := initTx.Rollback(); rbErr != nil {
			fieldslog.Error(es, "error rolling back init transaction", rbErr)
		}

		return err
	}

	if err := initTx.Commit(); err != nil {
		fieldslog.Error(es, "error committing transaction when initialized box", err)

		if rbErr := initTx.Rollback(); rbErr != nil {
			fieldslog.Error(es, "error rolling back init transaction after commit failure", rbErr)
		}

		return err
	}

	for {
		// Exit if context is canceled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		tx, err := es.transactional.BeginTransaction(ctx, transactional.DefaultWriteTransactionOptions())
		if err != nil {
			fieldslog.Error(es, "error beginning transaction", err)

			es.wait(ctx, err.Error(), es.waitTime)

			continue
		}

		messageRetrieved, err := es.processMessages(ctx, tx)
		if err != nil {
			fieldslog.Error(es, "error processing messages", err)

			if rbErr := tx.Rollback(); rbErr != nil {
				fieldslog.Error(es, "error rolling back transaction after processing failure", rbErr)
			}

			es.wait(ctx, err.Error(), es.waitTime)

			continue
		}

		if err := tx.Commit(); err != nil {
			fieldslog.Error(es, "error committing transaction", err)

			if rbErr := tx.Rollback(); rbErr != nil {
				fieldslog.Error(es, "error rolling back transaction after commit failure", rbErr)
			}

			es.wait(ctx, err.Error(), es.waitTime)

			continue
		}

		// If no messages are retrieved, wait for the specified duration
		if messageRetrieved == 0 {
			es.wait(ctx, "", es.waitTime)
		} else {
			es.wait(ctx, "", es.waitTimeIfMessages)
		}
	}
}

func (es *fifoSubscriber) wait(ctx context.Context, cause string, d time.Duration) {
	if cause != "" {
		fieldslog.Info(es, cause)
	}

	select {
	case <-ctx.Done():
		return
	case <-time.After(d):
		return
	}
}

func (es *fifoSubscriber) processMessages(ctx context.Context, tx transactional.Transaction) (int, error) {
	messages, err := es.box.Retrieve(ctx, tx, es.name, es.messageTypesProcessing(), es.batchSize, es.waitTime)
	if err != nil {
		return 0, err
	}

	messageRetrieved := len(messages)

	var lastAckedMessageOffset, lastConsumedMessageOffset int

	for _, msg := range messages {
		es.processedCount++

		handler, exists := es.handlers[msg.Type()]
		if exists {
			lastConsumedMessageOffset = msg.Offset()

			if err := handler.HandleEvent(ctx, tx, msg); err != nil {
				return messageRetrieved, err
			}

			lastAckedMessageOffset = msg.Offset()
		}
	}

	if messageRetrieved > 0 {
		if err := es.box.UpdateConsumer(ctx, tx, es.name, lastAckedMessageOffset, lastConsumedMessageOffset); err != nil {
			fieldslog.Error(es, "error updating consumer", err)

			return messageRetrieved, nil
		}
	}

	return messageRetrieved, nil

}

func (es *fifoSubscriber) RegisterHandler(messageType string, handler Handler) {
	es.handlersLock.Lock()
	defer es.handlersLock.Unlock()

	es.handlers[messageType] = handler
}

func (es *fifoSubscriber) UnregisterHandler(messageType string) error {
	es.handlersLock.Lock()
	defer es.handlersLock.Unlock()

	delete(es.handlers, messageType)

	return nil
}

func (es *fifoSubscriber) Status() SubscriberStatus {
	return SubscriberStatus{
		ProcessedMessageCount: es.processedCount,
	}
}

func (es *fifoSubscriber) messageTypesProcessing() []string {
	keys := make([]string, 0, len(es.handlers))
	for k := range es.handlers {
		keys = append(keys, k)
	}

	return keys
}
