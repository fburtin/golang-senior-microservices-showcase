## Development Branches

The project follows a feature branch workflow.

### Main Branch

The `main` branch contains the stable implementation:

- REST API
- Clean Architecture
- MongoDB Repository
- MongoDB Transactions
- MongoDB Replica Set
- Apache Kafka Producer
- Apache Kafka Consumer
- Swagger Documentation
- Unit Tests

### Transactional Outbox

The Transactional Outbox implementation is currently under active development in the `feature/transactional-outbox` branch.

Current progress:

- ✅ MongoDB Replica Set
- ✅ MongoDB Transactions
- ✅ Kafka Producer
- ✅ Kafka Consumer
- ✅ Atomic persistence of Customer and Outbox Event
- ✅ Background Outbox Publisher
- ✅ Retry Policy
- 🚧 Idempotent Event Publishing

The branch demonstrates the incremental implementation of the Transactional Outbox Pattern, a common approach used in distributed systems to guarantee reliable event publication while maintaining database consistency.
