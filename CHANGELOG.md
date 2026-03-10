# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-07-16

### Added

- Observability features with OpenTelemetry integration
- Distributed logging capabilities

### Changed

- Production docker-compose configuration

### Fixed

- Database locking mechanisms

## [0.0.9] - 2025-06-26

### Added

- API documentation using Swag for Swagger 2.0 generation

### Fixed

- Docker image build process

## [0.0.8] - 2025-06-12

### Added

- Integration event handling
- Domain event implementation
- Feature-based CQRS pattern for customer and order modules
- Data isolation between modules
- Public API contract definitions for customer module
- Monorepo structure support

## [0.0.7] - 2025-06-11

### Added

- Summary for day 2 of development
- Internal folder organization
- Service registry pattern
- Sub-domain encapsulation
- Modular structure implementation
- Dependency inversion principle

### Changed

- Implemented unit of work pattern with post-commit hooks
- Database transaction handling

## [0.0.6] - 2025-06-10

### Added

- Summary for day 1 of development
- Layered architecture for order module
- Layered architecture for notification module
- Custom error handling with multiple approaches:
  - AppError creation
  - response.JSONError usage
  - ResponseError middleware
- Layered architecture for customer module
- Database connection handling with context
- Database migration support
- Database context implementation

### Changed

- Separated business logic from router for order module
- Separated business logic from router for customer module

## [0.0.5] - 2025-06-09

### Added

- Docker image building capabilities
- Application build process
- Refactoring for Single Responsibility Principle (SRP)
- Makefile for common operations
- Configuration management
- Centralized logging system
- Graceful shutdown handling
- REST client testing capabilities
- Routing implementation
- Basic web server

[Unreleased]: https://github.com/somprasongd/course-go-mma/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/somprasongd/course-go-mma/compare/v0.0.9...v0.1.0
[0.0.9]: https://github.com/somprasongd/course-go-mma/compare/v0.0.8...v0.0.9
[0.0.8]: https://github.com/somprasongd/course-go-mma/compare/v0.0.7...v0.0.8
[0.0.7]: https://github.com/somprasongd/course-go-mma/compare/v0.0.6...v0.0.7
[0.0.6]: https://github.com/somprasongd/course-go-mma/compare/v0.0.5...v0.0.6
[0.0.5]: https://github.com/somprasongd/course-go-mma/releases/tag/v0.0.5
