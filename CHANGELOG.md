# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-01-03

### Added
- Initial release of the Nexmonyx Go SDK
- Complete API client implementation with multiple authentication methods:
  - JWT token authentication
  - API key/secret authentication
  - Server UUID/secret authentication
  - Monitoring key authentication
- Core service implementations:
  - Organizations management
  - Servers management
  - Users management
  - Metrics collection and querying
  - Hardware inventory management
  - Monitoring and probes
  - Alerts management
  - Systemd services monitoring
  - Billing and subscription management
  - Settings management
  - API keys management
  - Health status monitoring
- Comprehensive error handling with typed errors
- Request retry logic with exponential backoff
- Pagination support for list operations
- Time range queries and aggregations
- Example implementations for common use cases
- Full test coverage for all services
- Detailed documentation and usage examples

### Features
- Type-safe Go interfaces for all API endpoints
- Context support for cancellation and timeouts
- Custom time parsing for flexible date formats
- Batch operations support
- Webhook and notification management
- Integration with third-party services
- Comprehensive hardware inventory tracking
- Real-time metrics and monitoring capabilities

### Documentation
- Complete API documentation
- Usage examples for all major features
- CLAUDE.md file for AI assistance
- Integration guides for common scenarios

[1.0.0]: https://github.com/nexmonyx/go-sdk/releases/tag/v1.0.0