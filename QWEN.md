# Go MMA (Modular Monolith Architecture) Project

## Project Overview

This project demonstrates how to build a **modular monolith** application in Go. It implements a simple e-commerce system with customer and order management features, following best practices for modular architecture, separation of concerns, and maintainability.

### Key Features

1. **Customer Management**
   - Create new customers with email and initial credit
   - Email validation and duplicate checking
   - Welcome email notification upon creation

2. **Order Management**
   - Create orders for existing customers
   - Credit validation and deduction
   - Cancel orders and credit refund

### Architecture

The project follows a **modular monolith** architecture with the following principles:

- **Modular Structure**: Code is organized by business capabilities (customer, order, notification modules)
- **Layered Architecture**: Each module follows CQRS (Command Query Responsibility Segregation) with clear separation between domain, application, and infrastructure layers
- **Event-Driven Design**: Uses domain events and integration events for loose coupling
- **Data Isolation**: Each module has its own database schema
- **Dependency Inversion**: Modules depend on contracts/interfaces rather than concrete implementations

### Technology Stack

- **Language**: Go 1.24+
- **Web Framework**: Fiber v3
- **Database**: PostgreSQL
- **Database Migration**: migrate/migrate
- **Logging**: Zap with OpenTelemetry integration
- **Observability**: Grafana stack (Loki, Tempo, Prometheus) with OpenTelemetry
- **Containerization**: Docker with multi-stage build
- **Configuration**: Environment variables

## Project Structure

```bash
.
├── config/                 # Configuration files
├── migrations/             # Database migration scripts
├── src/
│   ├── app/                # Main application entry point
│   │   ├── application/    # Application layer (HTTP server, middleware)
│   │   ├── build/          # Build information
│   │   ├── cmd/api/        # Main entry point
│   │   ├── config/         # Configuration loading
│   │   └── docs/           # API documentation (Swagger)
│   ├── modules/            # Business modules
│   │   ├── customer/       # Customer management module
│   │   ├── notification/   # Notification service module
│   │   └── order/          # Order management module
│   └── shared/             # Shared libraries
│       ├── common/         # Common utilities (logging, errors, etc.)
│       ├── contract/       # Module contracts/interfaces
│       └── messaging/      # Integration event definitions
├── docker-compose.yml      # Base docker-compose configuration
├── docker-compose.dev.yml  # Development environment overrides
├── docker-compose.prod.yml # Production environment configuration
├── Dockerfile              # Multi-stage Docker build
├── Makefile               # Build and development commands
└── README.md              # Detailed project documentation
```

## Building and Running

### Prerequisites

- Go 1.24+
- Docker and Docker Compose
- PostgreSQL client (for direct database access)

### Development Setup

1. **Environment Configuration**:

   ```bash
   cp .env.example .env
   # Edit .env to match your environment
   ```

2. **Start Development Database**:

   ```bash
   make devup
   ```

3. **Run Database Migrations**:

   ```bash
   make mgu
   ```

4. **Run Application**:

   ```bash
   make run
   ```

### Production Deployment

1. **Build Docker Image**:

   ```bash
   BUILD_VERSION=1.0.0 make image
   ```

2. **Start Production Environment**:

   ```bash
   make produp
   ```

### Key Commands

| Command | Description |
|---------|-------------|
| `make run` | Run application in development mode |
| `make build` | Build application binary |
| `make image` | Build Docker image |
| `make devup` | Start development environment |
| `make devdown` | Stop development environment |
| `make produp` | Start production environment |
| `make proddown` | Stop production environment |
| `make mgu` | Run database migrations up |
| `make mgd` | Run database migrations down |
| `make doc` | Generate API documentation |

## API Endpoints

### Customer Management

- **POST /api/v1/customers**
  - Create a new customer
  - Request body: `{ "email": "string", "credit": "integer" }`
  - Response: `{ "id": "integer" }`

### Order Management

- **POST /api/v1/orders**
  - Create a new order
  - Request body: `{ "customer_id": "integer", "order_total": "integer" }`
  - Response: `{ "id": "integer" }`

- **DELETE /api/v1/orders/{orderID}**
  - Cancel an existing order
  - Response: 204 No Content

## Security Features

### XSS Protection

The application includes built-in XSS (Cross-Site Scripting) protection through input sanitization using two approaches:

1. **Struct Tag-Based Sanitization**: Uses reflection to sanitize struct fields based on tags
2. **Direct Function Calls**: Explicit sanitization functions for specific use cases

#### Struct Tag-Based Sanitization

The preferred approach uses struct tags to declare sanitization rules:

```go
type CreateCustomerRequest struct {
    Email  string `json:"email" sanitize:"email"`
    Credit int    `json:"credit"`
}
```

Available sanitization tags:

- `email`: Sanitizes email addresses by removing HTML tags
- `xss`: Removes common XSS attack vectors
- `trim`: Trims whitespace from the beginning and end
- `lower`: Converts to lowercase

The sanitization is applied in the HTTP endpoint:

```go
// Sanitize inputs using struct tags to prevent XSS attacks
sanitizer.SanitizeStruct(&req)
```

This approach provides:

- **Declarative**: Clear intent about which fields need sanitization
- **Reusable**: Can apply different sanitization rules per field
- **Centralized**: One sanitizer function can handle all tagged structs
- **Maintainable**: Easy to add new sanitization rules

#### Direct Function Calls

For more specific sanitization needs, direct function calls are available:

```go
cleanEmail := sanitizer.SanitizeEmail(dirtyEmail)
cleanText := sanitizer.SanitizeInput(dirtyText)
```

## Development Practices

### Modular Design

Each module (customer, order, notification) follows these principles:

1. **Feature-Based Structure**: Code is organized by features rather than technical layers
2. **CQRS Pattern**: Commands (write operations) and Queries (read operations) are separated
3. **Mediator Pattern**: Request/Response handling through a mediator
4. **Domain Events**: Business events that occur within the domain
5. **Integration Events**: Events that cross module boundaries

### Testing

The project includes comprehensive testing practices:

1. **Unit Tests**: For individual components and functions
2. **Integration Tests**: For module interactions and database operations
3. **Contract Tests**: For API endpoint validation

### Observability

The application includes built-in observability features:

1. **Structured Logging**: JSON-formatted logs with trace correlation
2. **Distributed Tracing**: OpenTelemetry integration for request tracing
3. **Metrics Collection**: Application and system metrics
4. **Grafana Stack**: Pre-configured monitoring with Loki, Tempo, and Prometheus

## Database Schema

The application uses PostgreSQL with schema-based data isolation:

- **customer schema**: Contains customer-related tables
- **sales schema**: Contains order-related tables

### Key Tables

1. **customer.customers**
   - `id`: BIGINT (Primary Key)
   - `email`: TEXT (Unique)
   - `credit`: INTEGER
   - `created_at`: TIMESTAMP
   - `updated_at`: TIMESTAMP

2. **sales.orders**
   - `id`: BIGINT (Primary Key)
   - `customer_id`: BIGINT (Foreign Key reference)
   - `order_total`: INTEGER
   - `created_at`: TIMESTAMP
   - `canceled_at`: TIMESTAMP (Nullable)

## Error Handling

The application uses a structured error handling system:

1. **Custom Error Types**: Defined error categories for different scenarios
2. **HTTP Status Mapping**: Appropriate HTTP status codes for each error type
3. **Error Response Format**: Consistent JSON error responses

## Extending the Application

To add new features or modules:

1. **Create New Module**: Follow the existing module structure
2. **Define Contracts**: Create interfaces for cross-module communication
3. **Implement Features**: Use CQRS pattern with commands and queries
4. **Add Domain Events**: For business event notifications
5. **Register Module**: Add to the main application initialization
6. **Update Documentation**: Add API documentation if needed

## Best Practices Demonstrated

1. **Clean Architecture**: Separation of concerns with clear boundaries
2. **SOLID Principles**: Especially Single Responsibility and Dependency Inversion
3. **Domain-Driven Design**: Rich domain models and ubiquitous language
4. **Event-Driven Architecture**: Loose coupling through events
5. **Observability**: Built-in monitoring and tracing
6. **Configuration Management**: Environment-based configuration
7. **Graceful Shutdown**: Proper resource cleanup on application termination
8. **Database Transactions**: Consistent data operations with rollback support
9. **Security Considerations**: Input sanitization and proper error handling without information leakage

## References

This project is based on the comprehensive guide "แนวทางการสร้างโปรเจกต์แบบ Modular Monolith" which provides detailed explanations of each architectural decision and implementation pattern.
