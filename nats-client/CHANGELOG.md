# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-11-28

### Added

#### JetStream Stream Management
- `CreateStream()` - Create a new JetStream stream
- `UpdateStream()` - Update an existing stream
- `DeleteStream()` - Delete a stream
- `GetStream()` - Get stream information
- `ListStreams()` - List all streams
- `StreamNames()` - Get stream names
- `PurgeStream()` - Purge all messages from a stream

#### JetStream Consumer Management
- `CreateConsumer()` - Create a new consumer
- `UpdateConsumer()` - Update an existing consumer
- `DeleteConsumer()` - Delete a consumer
- `GetConsumer()` - Get consumer information
- `ListConsumers()` - List all consumers for a stream
- `ConsumerNames()` - Get consumer names for a stream

#### JetStream Publishing
- `PublishToStream()` - Publish with acknowledgement
- `PublishToStreamAsync()` - Async publish with future
- `PublishMsgToStream()` - Publish message object to stream

#### JetStream Subscribing
- `PullSubscribe()` - Create pull-based subscription
- `SubscribeToStream()` - Create push subscription to stream
- `QueueSubscribeToStream()` - Create queue subscription to stream
- `ChanSubscribeToStream()` - Create channel subscription to stream

#### JetStream Message Operations
- `GetMsg()` - Get message by sequence number
- `GetLastMsg()` - Get last message for a subject
- `DeleteMsg()` - Delete message by sequence
- `SecureDeleteMsg()` - Securely delete message

#### New Publish Methods
- `PublishMsg()` - Publish a NATS message object
- `PublishRequest()` - Publish with reply subject

#### New Request Methods
- `RequestWithContext()` - Request with context for cancellation/timeout
- `RequestMsgWithContext()` - Request message with context

#### New Subscribe Methods
- `ChanSubscribe()` - Channel-based subscription
- `QueueChanSubscribe()` - Queue group channel subscription

#### Environment Configuration
- `NewFromEnv()` - Create client from environment variables with custom prefix
- `NewFromEnvWithDefaults()` - Create client with default `NATS_` prefix
- Added `env.go` with full environment variable support via `env-util`

#### New Error Types
- `ErrJetStreamNotEnabled` - JetStream not enabled in config
- `ErrJetStreamNotInitialized` - JetStream context not initialized
- `ErrStreamNotFound` - Stream not found
- `ErrConsumerNotFound` - Consumer not found
- `ErrMessageNotFound` - Message not found

#### Error Helper Functions
- `IsJetStreamError()` - Check if error is JetStream related
- `IsNotFoundError()` - Check if error is a not found error

#### Configuration Types
- `StreamConfig` - Stream configuration struct
- `ConsumerConfig` - Consumer configuration struct
- `RetentionPolicy` - LimitsPolicy, InterestPolicy, WorkQueuePolicy
- `StorageType` - FileStorage, MemoryStorage
- `DeliverPolicy` - DeliverAll, DeliverLast, DeliverNew, etc.
- `AckPolicy` - AckNone, AckAll, AckExplicit
- `ReplayPolicy` - ReplayInstant, ReplayOriginal

#### Development
- Added `docker-compose.yml` for local NATS with JetStream
- Added integration tests (`integration_test.go`)
- Added Makefile targets: `docker-up`, `docker-down`, `docker-logs`, `docker-status`
- Added `test-integration` and `test-integration-keep` targets

### Changed
- `Request()` method signature changed from `context.Context` to `time.Duration` for timeout
- Improved test coverage from 31.9% to 70.0% (unit) / 81.7% (with integration)

### Fixed
- Request method signature now matches README documentation

## [0.0.2] - 2024-XX-XX

### Added
- Initial JetStream context support via `JetStream()` method
- Basic pub/sub functionality
- Request/reply pattern support
- Queue group subscriptions
- TLS/SSL support
- Functional options pattern
- Connection monitoring

## [0.0.1] - 2024-XX-XX

### Added
- Initial release
- Basic NATS client wrapper
- Configuration management
- Error handling

[Unreleased]: https://github.com/isimtekin/go-packages/compare/nats-client/v0.0.2...HEAD
[0.0.2]: https://github.com/isimtekin/go-packages/compare/nats-client/v0.0.1...nats-client/v0.0.2
[0.0.1]: https://github.com/isimtekin/go-packages/releases/tag/nats-client/v0.0.1
