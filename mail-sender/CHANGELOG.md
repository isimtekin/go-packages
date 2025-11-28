# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2025-11-28

### Added
- **SenderManager**: New component for managing multiple email providers simultaneously
- **SMTP Provider**: Standard SMTP support with TLS/SSL encryption and authentication
- **AWS SES Provider**: Amazon Simple Email Service integration with credential chain support
- Provider selection at runtime via `manager.Send(ctx, "provider-name", message)`

### Changed
- Updated test coverage to 85.8%

## [0.1.1] - 2024-01-15

### Fixed
- Minor bug fixes and improvements

## [0.1.0] - 2024-01-10

### Added
- Initial release
- SendGrid provider support
- Async/event-based email sending with worker pools
- HTML and plain text template rendering
- Event handlers (OnSuccess, OnFailure, OnRetry)
- Automatic retry logic with configurable attempts
- Real-time statistics tracking
- Graceful shutdown support
- Environment variable configuration
- Functional options pattern
