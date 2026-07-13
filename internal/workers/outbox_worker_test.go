package workers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/fburtin/golang-senior-microservices-showcase/internal/domain"
	"github.com/fburtin/golang-senior-microservices-showcase/internal/messaging"
)

type fakeOutboxRepository struct {
	pendingEvents []domain.OutboxEvent
	findError     error

	findPendingCalled bool
	findPendingLimit  int64

	markPublishedCalled bool
	markPublishedID     string
	markPublishedError  error

	recordFailureCalled bool
	recordFailureID     string
	recordFailureReason string
	recordFailureError  error

	markFailedCalled bool
	markFailedID     string
	markFailedReason string
	markFailedError  error

	createIndexesCalled bool
	createIndexesError  error
}

func (r *fakeOutboxRepository) FindPending(
	ctx context.Context,
	limit int64,
) ([]domain.OutboxEvent, error) {
	r.findPendingCalled = true
	r.findPendingLimit = limit

	return r.pendingEvents, r.findError
}

func (r *fakeOutboxRepository) MarkPublished(
	ctx context.Context,
	id string,
) error {
	r.markPublishedCalled = true
	r.markPublishedID = id

	return r.markPublishedError
}

func (r *fakeOutboxRepository) RecordFailure(
	ctx context.Context,
	id string,
	reason string,
) error {
	r.recordFailureCalled = true
	r.recordFailureID = id
	r.recordFailureReason = reason

	return r.recordFailureError
}

func (r *fakeOutboxRepository) MarkFailed(
	ctx context.Context,
	id string,
	reason string,
) error {
	r.markFailedCalled = true
	r.markFailedID = id
	r.markFailedReason = reason

	return r.markFailedError
}

func (r *fakeOutboxRepository) CreateIndexes(
	ctx context.Context,
) error {
	r.createIndexesCalled = true

	return r.createIndexesError
}

type fakeCustomerEventProducer struct {
	publishCalled bool
	published     messaging.CustomerCreatedEvent
	publishError  error
	callCount     int
}

func (p *fakeCustomerEventProducer) PublishCustomerCreated(
	ctx context.Context,
	event messaging.CustomerCreatedEvent,
) error {
	p.publishCalled = true
	p.published = event
	p.callCount++

	return p.publishError
}

func newTestLogger() *slog.Logger {
	return slog.New(
		slog.NewTextHandler(
			io.Discard,
			nil,
		),
	)
}

func createOutboxEvent(
	t *testing.T,
	id string,
	attempts int,
) domain.OutboxEvent {
	t.Helper()

	event := messaging.CustomerCreatedEvent{
		EventID:   "event-1",
		EventType: "customer.created",
		ID:        "customer-1",
		FirstName: "Francisco",
		LastName:  "Burtin",
		Email:     "francisco@example.com",
		CreatedAt: time.Now().UTC(),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	return domain.OutboxEvent{
		ID:            id,
		EventID:       event.EventID,
		AggregateID:   event.ID,
		AggregateType: "customer",
		EventType:     event.EventType,
		Payload:       payload,
		Status:        domain.OutboxPending,
		Attempts:      attempts,
		CreatedAt:     time.Now().UTC(),
	}
}

func TestOutboxWorker_ProcessBatch_PublishesPendingEvent(
	t *testing.T,
) {
	outboxEvent := createOutboxEvent(
		t,
		"outbox-1",
		0,
	)

	repository := &fakeOutboxRepository{
		pendingEvents: []domain.OutboxEvent{
			outboxEvent,
		},
	}

	producer := &fakeCustomerEventProducer{}

	worker := NewOutboxWorker(
		repository,
		producer,
		newTestLogger(),
		time.Second,
	)

	worker.processBatch(context.Background())

	if !repository.findPendingCalled {
		t.Fatal("expected FindPending to be called")
	}

	if repository.findPendingLimit != defaultBatchSize {
		t.Fatalf(
			"expected limit %d, got %d",
			defaultBatchSize,
			repository.findPendingLimit,
		)
	}

	if !producer.publishCalled {
		t.Fatal(
			"expected PublishCustomerCreated to be called",
		)
	}

	if producer.published.ID != outboxEvent.AggregateID {
		t.Fatalf(
			"expected customer ID %q, got %q",
			outboxEvent.AggregateID,
			producer.published.ID,
		)
	}

	if !repository.markPublishedCalled {
		t.Fatal("expected MarkPublished to be called")
	}

	if repository.markPublishedID != outboxEvent.ID {
		t.Fatalf(
			"expected published ID %q, got %q",
			outboxEvent.ID,
			repository.markPublishedID,
		)
	}

	if repository.recordFailureCalled {
		t.Fatal(
			"expected RecordFailure not to be called",
		)
	}

	if repository.markFailedCalled {
		t.Fatal("expected MarkFailed not to be called")
	}
}

func TestOutboxWorker_ProcessBatch_NoPendingEvents(
	t *testing.T,
) {
	repository := &fakeOutboxRepository{}
	producer := &fakeCustomerEventProducer{}

	worker := NewOutboxWorker(
		repository,
		producer,
		newTestLogger(),
		time.Second,
	)

	worker.processBatch(context.Background())

	if !repository.findPendingCalled {
		t.Fatal("expected FindPending to be called")
	}

	if producer.publishCalled {
		t.Fatal(
			"expected producer not to be called",
		)
	}

	if repository.markPublishedCalled {
		t.Fatal(
			"expected MarkPublished not to be called",
		)
	}
}

func TestOutboxWorker_ProcessBatch_FindPendingFails(
	t *testing.T,
) {
	repository := &fakeOutboxRepository{
		findError: errors.New("database unavailable"),
	}

	producer := &fakeCustomerEventProducer{}

	worker := NewOutboxWorker(
		repository,
		producer,
		newTestLogger(),
		time.Second,
	)

	worker.processBatch(context.Background())

	if !repository.findPendingCalled {
		t.Fatal("expected FindPending to be called")
	}

	if producer.publishCalled {
		t.Fatal(
			"expected producer not to be called",
		)
	}
}

func TestOutboxWorker_ProcessEvent_InvalidPayloadMarksFailed(
	t *testing.T,
) {
	repository := &fakeOutboxRepository{}
	producer := &fakeCustomerEventProducer{}

	worker := NewOutboxWorker(
		repository,
		producer,
		newTestLogger(),
		time.Second,
	)

	event := domain.OutboxEvent{
		ID:      "outbox-1",
		Payload: []byte("invalid-json"),
	}

	worker.processEvent(
		context.Background(),
		event,
	)

	if !repository.markFailedCalled {
		t.Fatal("expected MarkFailed to be called")
	}

	if repository.markFailedID != event.ID {
		t.Fatalf(
			"expected failed ID %q, got %q",
			event.ID,
			repository.markFailedID,
		)
	}

	if repository.markFailedReason == "" {
		t.Fatal(
			"expected failure reason to be recorded",
		)
	}

	if producer.publishCalled {
		t.Fatal(
			"expected producer not to be called",
		)
	}
}

func TestOutboxWorker_ProcessEvent_PublishFailureRecordsFailure(
	t *testing.T,
) {
	expectedError := errors.New("kafka unavailable")

	repository := &fakeOutboxRepository{}
	producer := &fakeCustomerEventProducer{
		publishError: expectedError,
	}

	worker := NewOutboxWorker(
		repository,
		producer,
		newTestLogger(),
		time.Second,
	)

	event := createOutboxEvent(
		t,
		"outbox-1",
		0,
	)

	worker.processEvent(
		context.Background(),
		event,
	)

	if !producer.publishCalled {
		t.Fatal(
			"expected producer to be called",
		)
	}

	if !repository.recordFailureCalled {
		t.Fatal(
			"expected RecordFailure to be called",
		)
	}

	if repository.recordFailureID != event.ID {
		t.Fatalf(
			"expected failure ID %q, got %q",
			event.ID,
			repository.recordFailureID,
		)
	}

	if repository.recordFailureReason != expectedError.Error() {
		t.Fatalf(
			"expected reason %q, got %q",
			expectedError.Error(),
			repository.recordFailureReason,
		)
	}

	if repository.markFailedCalled {
		t.Fatal(
			"expected MarkFailed not to be called before maximum attempts",
		)
	}

	if repository.markPublishedCalled {
		t.Fatal(
			"expected MarkPublished not to be called",
		)
	}
}

func TestOutboxWorker_ProcessEvent_MaxAttemptsMarksFailed(
	t *testing.T,
) {
	expectedError := errors.New("kafka unavailable")

	repository := &fakeOutboxRepository{}
	producer := &fakeCustomerEventProducer{
		publishError: expectedError,
	}

	worker := NewOutboxWorker(
		repository,
		producer,
		newTestLogger(),
		time.Second,
	)

	event := createOutboxEvent(
		t,
		"outbox-1",
		maxPublishAttempts-1,
	)

	worker.processEvent(
		context.Background(),
		event,
	)

	if !repository.recordFailureCalled {
		t.Fatal(
			"expected RecordFailure to be called",
		)
	}

	if !repository.markFailedCalled {
		t.Fatal("expected MarkFailed to be called")
	}

	if repository.markFailedID != event.ID {
		t.Fatalf(
			"expected failed ID %q, got %q",
			event.ID,
			repository.markFailedID,
		)
	}

	if repository.markFailedReason != expectedError.Error() {
		t.Fatalf(
			"expected reason %q, got %q",
			expectedError.Error(),
			repository.markFailedReason,
		)
	}

	if repository.markPublishedCalled {
		t.Fatal(
			"expected MarkPublished not to be called",
		)
	}
}

func TestOutboxWorker_ProcessEvent_MarkPublishedFails(
	t *testing.T,
) {
	repository := &fakeOutboxRepository{
		markPublishedError: errors.New(
			"database update failed",
		),
	}

	producer := &fakeCustomerEventProducer{}

	worker := NewOutboxWorker(
		repository,
		producer,
		newTestLogger(),
		time.Second,
	)

	event := createOutboxEvent(
		t,
		"outbox-1",
		0,
	)

	worker.processEvent(
		context.Background(),
		event,
	)

	if !producer.publishCalled {
		t.Fatal(
			"expected producer to be called",
		)
	}

	if !repository.markPublishedCalled {
		t.Fatal(
			"expected MarkPublished to be called",
		)
	}

	if repository.recordFailureCalled {
		t.Fatal(
			"expected RecordFailure not to be called",
		)
	}

	if repository.markFailedCalled {
		t.Fatal(
			"expected MarkFailed not to be called",
		)
	}
}

func TestOutboxWorker_ProcessBatch_MultipleEvents(
	t *testing.T,
) {
	repository := &fakeOutboxRepository{
		pendingEvents: []domain.OutboxEvent{
			createOutboxEvent(t, "outbox-1", 0),
			createOutboxEvent(t, "outbox-2", 0),
			createOutboxEvent(t, "outbox-3", 0),
		},
	}

	producer := &fakeCustomerEventProducer{}

	worker := NewOutboxWorker(
		repository,
		producer,
		newTestLogger(),
		time.Second,
	)

	worker.processBatch(context.Background())

	if producer.callCount != 3 {
		t.Fatalf(
			"expected 3 publish calls, got %d",
			producer.callCount,
		)
	}
}

func TestOutboxWorker_Run_StopsWhenContextIsCancelled(
	t *testing.T,
) {
	repository := &fakeOutboxRepository{}
	producer := &fakeCustomerEventProducer{}

	worker := NewOutboxWorker(
		repository,
		producer,
		newTestLogger(),
		10*time.Millisecond,
	)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	done := make(chan struct{})

	go func() {
		worker.Run(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal(
			"worker did not stop after context cancellation",
		)
	}

	if !repository.findPendingCalled {
		t.Fatal(
			"expected initial batch to be processed",
		)
	}
}
