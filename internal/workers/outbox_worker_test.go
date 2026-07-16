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
	claimedEvents []domain.OutboxEvent
	claimError    error

	claimCalled   bool
	claimWorkerID string
	claimLimit    int64

	releaseCalled      bool
	releaseStaleBefore time.Time
	releaseError       error

	markPublishedCalled bool
	markPublishedID     string
	markPublishedError  error

	recordFailureCalled        bool
	recordFailureID            string
	recordFailureReason        string
	recordFailureNextAttemptAt time.Time
	recordFailureError         error

	markFailedCalled bool
	markFailedID     string
	markFailedReason string
	markFailedError  error
}

func (r *fakeOutboxRepository) ClaimPending(
	ctx context.Context,
	workerID string,
	limit int64,
) ([]domain.OutboxEvent, error) {
	r.claimCalled = true
	r.claimWorkerID = workerID
	r.claimLimit = limit

	return r.claimedEvents, r.claimError
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
	nextAttemptAt time.Time,
) error {
	r.recordFailureCalled = true
	r.recordFailureID = id
	r.recordFailureReason = reason
	r.recordFailureNextAttemptAt = nextAttemptAt

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

func (r *fakeOutboxRepository) ReleaseStaleLocks(
	ctx context.Context,
	staleBefore time.Time,
) error {
	r.releaseCalled = true
	r.releaseStaleBefore = staleBefore

	return r.releaseError
}

func (r *fakeOutboxRepository) CreateIndexes(
	ctx context.Context,
) error {
	return nil
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
		slog.NewTextHandler(io.Discard, nil),
	)
}

func newTestWorker(
	repository *fakeOutboxRepository,
	producer *fakeCustomerEventProducer,
) *OutboxWorker {
	return NewOutboxWorker(
		repository,
		producer,
		newTestLogger(),
		time.Second,
		"worker-test",
		30*time.Second,
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

	now := time.Now().UTC()

	return domain.OutboxEvent{
		ID:            id,
		EventID:       event.EventID,
		AggregateID:   event.ID,
		AggregateType: "customer",
		EventType:     event.EventType,
		Payload:       payload,
		Status:        domain.OutboxProcessing,
		Attempts:      attempts,
		CreatedAt:     now,
		LockedAt:      &now,
		LockedBy:      "worker-test",
	}
}

func TestOutboxWorker_ProcessBatch_ClaimsAndPublishesEvent(
	t *testing.T,
) {
	event := createOutboxEvent(t, "outbox-1", 0)

	repository := &fakeOutboxRepository{
		claimedEvents: []domain.OutboxEvent{event},
	}
	producer := &fakeCustomerEventProducer{}
	worker := newTestWorker(repository, producer)

	worker.processBatch(context.Background())

	if !repository.releaseCalled {
		t.Fatal("expected ReleaseStaleLocks to be called")
	}

	if !repository.claimCalled {
		t.Fatal("expected ClaimPending to be called")
	}

	if repository.claimWorkerID != "worker-test" {
		t.Fatalf(
			"expected worker ID %q, got %q",
			"worker-test",
			repository.claimWorkerID,
		)
	}

	if repository.claimLimit != defaultBatchSize {
		t.Fatalf(
			"expected limit %d, got %d",
			defaultBatchSize,
			repository.claimLimit,
		)
	}

	if !producer.publishCalled {
		t.Fatal("expected producer to be called")
	}

	if !repository.markPublishedCalled {
		t.Fatal("expected MarkPublished to be called")
	}

	if repository.markPublishedID != event.ID {
		t.Fatalf(
			"expected published ID %q, got %q",
			event.ID,
			repository.markPublishedID,
		)
	}
}

func TestOutboxWorker_ProcessBatch_ClaimFailureDoesNotPublish(
	t *testing.T,
) {
	repository := &fakeOutboxRepository{
		claimError: errors.New("database unavailable"),
	}
	producer := &fakeCustomerEventProducer{}
	worker := newTestWorker(repository, producer)

	worker.processBatch(context.Background())

	if !repository.claimCalled {
		t.Fatal("expected ClaimPending to be called")
	}

	if producer.publishCalled {
		t.Fatal("expected producer not to be called")
	}
}

func TestOutboxWorker_ProcessBatch_ReleaseFailureStillClaims(
	t *testing.T,
) {
	repository := &fakeOutboxRepository{
		releaseError: errors.New("release failed"),
	}
	producer := &fakeCustomerEventProducer{}
	worker := newTestWorker(repository, producer)

	worker.processBatch(context.Background())

	if !repository.releaseCalled {
		t.Fatal("expected ReleaseStaleLocks to be called")
	}

	if !repository.claimCalled {
		t.Fatal("expected ClaimPending to still be called")
	}
}

func TestOutboxWorker_ProcessEvent_InvalidPayloadMarksFailed(
	t *testing.T,
) {
	repository := &fakeOutboxRepository{}
	producer := &fakeCustomerEventProducer{}
	worker := newTestWorker(repository, producer)

	event := domain.OutboxEvent{
		ID:      "outbox-1",
		EventID: "event-1",
		Payload: []byte("invalid-json"),
	}

	worker.processEvent(context.Background(), event)

	if !repository.markFailedCalled {
		t.Fatal("expected MarkFailed to be called")
	}

	if producer.publishCalled {
		t.Fatal("expected producer not to be called")
	}
}

func TestOutboxWorker_ProcessEvent_PublishFailureSchedulesRetry(
	t *testing.T,
) {
	expectedError := errors.New("kafka unavailable")

	repository := &fakeOutboxRepository{}
	producer := &fakeCustomerEventProducer{
		publishError: expectedError,
	}
	worker := newTestWorker(repository, producer)
	event := createOutboxEvent(t, "outbox-1", 0)

	before := time.Now().UTC()
	worker.processEvent(context.Background(), event)

	if !repository.recordFailureCalled {
		t.Fatal("expected RecordFailure to be called")
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

	expectedMinimum := before.Add(baseRetryDelay)
	if repository.recordFailureNextAttemptAt.Before(expectedMinimum) {
		t.Fatalf(
			"expected retry at or after %v, got %v",
			expectedMinimum,
			repository.recordFailureNextAttemptAt,
		)
	}

	if repository.markFailedCalled {
		t.Fatal("expected MarkFailed not to be called")
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
	worker := newTestWorker(repository, producer)
	event := createOutboxEvent(
		t,
		"outbox-1",
		maxPublishAttempts-1,
	)

	worker.processEvent(context.Background(), event)

	if !repository.markFailedCalled {
		t.Fatal("expected MarkFailed to be called")
	}

	if repository.recordFailureCalled {
		t.Fatal("expected RecordFailure not to be called")
	}
}

func TestOutboxWorker_ProcessEvent_MarkPublishedFailureDoesNotRetryPublish(
	t *testing.T,
) {
	repository := &fakeOutboxRepository{
		markPublishedError: errors.New("database update failed"),
	}
	producer := &fakeCustomerEventProducer{}
	worker := newTestWorker(repository, producer)
	event := createOutboxEvent(t, "outbox-1", 0)

	worker.processEvent(context.Background(), event)

	if producer.callCount != 1 {
		t.Fatalf(
			"expected producer call count 1, got %d",
			producer.callCount,
		)
	}

	if !repository.markPublishedCalled {
		t.Fatal("expected MarkPublished to be called")
	}

	if repository.recordFailureCalled {
		t.Fatal("expected RecordFailure not to be called")
	}
}

func TestRetryDelay_UsesExponentialBackoffAndCap(
	t *testing.T,
) {
	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{
			name:     "first failure",
			attempt:  0,
			expected: 5 * time.Second,
		},
		{
			name:     "second failure",
			attempt:  1,
			expected: 10 * time.Second,
		},
		{
			name:     "third failure",
			attempt:  2,
			expected: 20 * time.Second,
		},
		{
			name:     "capped",
			attempt:  20,
			expected: maxRetryDelay,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := retryDelay(test.attempt)

			if actual != test.expected {
				t.Fatalf(
					"expected %v, got %v",
					test.expected,
					actual,
				)
			}
		})
	}
}
