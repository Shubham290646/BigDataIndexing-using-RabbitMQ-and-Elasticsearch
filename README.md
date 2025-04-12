


# BigData Indexing using RabbitMQ and Elasticsearch

A RESTful API system for managing structured JSON objects. It supports CRUD operations, JSON Schema validation, Redis-based key-value storage, Elasticsearch indexing, and asynchronous processing with RabbitMQ.

---

## 🚀 Features

- **RESTful API**: Supports POST, PUT, PATCH, GET, and DELETE methods for JSON data.
- **JSON Schema Validation**: Ensures incoming data follows predefined schemas.
- **Redis Integration**: Efficient key-value storage for structured data.
- **Elasticsearch Indexing**: Enables powerful search capabilities.
- **RabbitMQ Queueing**: Handles indexing requests asynchronously.
- **ETag Caching**: Implements ETag headers for cache validation.

---

## 🛠 Tech Stack

- **Language**: Go (Gin-Gonic framework)
- **Cache**: Redis
- **Search Engine**: Elasticsearch
- **Message Broker**: RabbitMQ
- **Containerization**: Docker & Docker Compose

---

## 🔄 Data Flow

1. **Authentication** (if implemented): OAuth 2.0 for secure access.
2. **Validation**: Validates input using JSON Schema.
3. **Creation**: Accepts POST requests to store JSON objects.
4. **Redis Storage**: Persists data as key-value pairs.
5. **Queueing**: Pushes indexing tasks to RabbitMQ.
6. **Indexing**: Consumes messages and indexes into Elasticsearch.
7. **Search**: Data can be queried using Kibana or Elasticsearch APIs.

---

## 📦 Setup Instructions

### Prerequisites

Make sure the following are installed:

- [Go](https://golang.org/doc/install)
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

---

Clone the Repository

```
git clone https://github.com/Shubham290646/BigDataIndexing-using-RabbitMQ-and-Elasticsearch.git
cd BigDataIndexing-using-RabbitMQ-and-Elasticsearch
```

---

### Start Services with Docker Compose

```bash
docker-compose up -d
```

This will start:
- Redis
- Elasticsearch
- Kibana
- RabbitMQ

---

### Run the Go Application

```bash
go run main.go
```

---

### Start the RabbitMQ Consumer

```bash
go run consumer/main.go
```

---

## 🔗 Service Endpoints

- **Elasticsearch**: [http://localhost:9200](http://localhost:9200)
- **Kibana**: [http://localhost:5601](http://localhost:5601)
- **RabbitMQ Admin**: [http://localhost:15672](http://localhost:15672)
  - Username: `guest`
  - Password: `guest`

---

## 📚 API Endpoints

### Plan Management

- `POST /v1/plan`: Create a new plan
- `PUT /v1/plan/{id}`: Update a plan (requires valid ETag)
- `PATCH /v1/plan/{id}`: Partial update (requires valid ETag)
- `GET /v1/plan/{id}`: Retrieve a plan (supports ETag caching)
- `DELETE /v1/plan/{id}`: Delete a plan (requires valid ETag)

---

## 🧪 Testing

Use tools like [Postman](https://www.postman.com/) or `curl` to test the API. Make sure to send appropriate headers:

- `Content-Type: application/json`
- `If-Match` / `If-None-Match` for ETag-based endpoints

---

## 📁 Project Structure

```plaintext
├── consumer/             # RabbitMQ consumer service
├── data/                 # Sample data and JSON schemas
├── database/             # Database connection and initialization
├── elastic/              # Elasticsearch integration
├── handlers/             # HTTP request handlers
├── middleware/           # Custom middleware
├── models/               # Data models and schemas
├── rabbitmq/             # RabbitMQ connection and publisher
├── repositories/         # Data access logic
├── routes/               # API route definitions
├── services/             # Business logic
├── docker-compose.yaml   # Docker Compose setup
├── go.mod                # Go modules config
└── main.go               # Main application entry
```

---

## 👨‍💻 Author

Shubham Mangaonkar  

