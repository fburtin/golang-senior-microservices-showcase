package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
)

type fakeProcessedEventRepository struct {
	reserved      bool
	tryStartCalls int
	completed     int
	removed       int
}

func (r *fakeProcessedEventRepository) TryStart(
	context.Context,
	string,
	string,
) (bool, error) {
	r.tryStartCalls++
	return r.reserved, nil
}

func (r *fakeProcessedEventRepository) MarkCompleted(context.Context, string) error {
	r.completed++
	return nil
}

func (r *fakeProcessedEventRepository) Remove(context.Context, string) error {
	r.removed++
	return nil
}

func (r *fakeProcessedEventRepository) CreateIndexes(context.Context) error {
	return nil
}

type fakeCustomerCreatedHandler struct {
	calls int
	err   error
}

func (h *fakeCustomerCreatedHandler) HandleCustomerCreated(
	context.Context,
	CustomerCreatedEvent,
) error {
	h.calls++
	return h.err
}

func testConsumer(
	inbox *fakeProcessedEventRepository,
	handler *fakeCustomerCreatedHandler,
) *CustomerConsumer {
	return &CustomerConsumer{
		inbox:   inbox,
		handler: handler,
		logger:  slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func testMessage(t *testing.T) kafka.Message {
	t.Helper()

	event := CustomerCreatedEvent{
		EventID:   "event-1",
		EventType: "customer.created",
		ID:        "customer-1",
		FirstName: "Francisco",
		LastName:  "Burtin",
		Email:     "francisco@example.com",
		CreatedAt: time.Now().UTC(),
	}

	value, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}

	return kafka.Message{
		Key:   []byte(event.EventID),
		Value: value,
	}
}

func TestConsumer_DuplicateEventDoesNotExecuteHandler(t *testing.T) {
	inbox := &fakeProcessedEventRepository{reserved: false}
	handler := &fakeCustomerCreatedHandler{}
	consumer := testConsumer(inbox, handler)

	if err := consumer.processMessage(context.Background(), testMessage(t)); err != nil {
		t.Fatal(err)
	}

	if handler.calls != 0 {
		t.Fatalf("expected no handler calls, got %d", handler.calls)
	}
	if inbox.completed != 0 {
		t.Fatalf("expected no completion update, got %d", inbox.completed)
	}
}

func TestConsumer_NewEventExecutesHandlerOnce(t *testing.T) {
	inbox := &fakeProcessedEventRepository{reserved: true}
	handler := &fakeCustomerCreatedHandler{}
	consumer := testConsumer(inbox, handler)

	if err := consumer.processMessage(context.Background(), testMessage(t)); err != nil {
		t.Fatal(err)
	}

	if handler.calls != 1 {
		t.Fatalf("expected one handler call, got %d", handler.calls)
	}
	if inbox.completed != 1 {
		t.Fatalf("expected one completion update, got %d", inbox.completed)
	}
}

func TestConsumer_HandlerFailureReleasesReservation(t *testing.T) {
	inbox := &fakeProcessedEventRepository{reserved: true}
	handler := &fakeCustomerCreatedHandler{err: errors.New("business failure")}
	consumer := testConsumer(inbox, handler)

	if err := consumer.processMessage(context.Background(), testMessage(t)); err == nil {
		t.Fatal("expected error")
	}

	if inbox.removed != 1 {
		t.Fatalf("expected reservation removal, got %d", inbox.removed)
	}
	if inbox.completed != 0 {
		t.Fatalf("expected no completion, got %d", inbox.completed)
	}
}
