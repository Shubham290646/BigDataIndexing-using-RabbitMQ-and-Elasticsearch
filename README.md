# info7255-bigdata-app
# BigData Indexing using RabbitMQ and Elasticsearch

A RESTful API system for healthcare plan management with advanced features including CRUD operations, JSON Schema validation, Redis key-value storage, Elasticsearch indexing, and message queueing with RabbitMQ.

## Features

- REST API supporting structured JSON data
- Complete CRUD operations with merge support and cascaded delete
- JSON Schema validation
- Key-value storage using Redis
- Parent-child indexing and search using Elasticsearch
- Message queueing with RabbitMQ for asynchronous operations
- JWT-based authentication

## Architecture

The system follows a microservice architecture:
1. **API Layer**: Handles HTTP requests and responses
2. **Service Layer**: Contains business logic and orchestrates operations
3. **Storage Layer**: Redis for primary data storage
4. **Search Layer**: Elasticsearch for advanced search capabilities
5. **Message Queue**: RabbitMQ for asynchronous operations

## Prerequisites

- Go 1.18+
- Redis server
- Elasticsearch 8.x
- RabbitMQ

## Setup and Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/Shubham290646/BigDataIndexing-using-RabbitMQ-and-Elasticsearch.git
   cd BigDataIndexing-using-RabbitMQ-and-Elasticsearch

Install dependencies
bashgo mod download

Start required services
bash# Start Redis
redis-server

# Start Elasticsearch
elasticsearch

# Start RabbitMQ
rabbitmq-server

Run the application
bashgo run main.go

Run the consumer (in a separate terminal)
bashgo run consumer.go


API Endpoints
MethodEndpointDescriptionPOST/v1/planCreate a new medical planGET/v1/plan/Retrieve a plan by IDPATCH/v1/plan/Update specific fields of a planPUT/v1/planReplace an entire planDELETE/v1/plan/Delete a plan and all associated resourcesPOST/v1/searchSearch for plans using Elasticsearch
Authentication
The API uses JWT token-based authentication:

Get a token using the appropriate authentication endpoint
Include the token in the Authorization header for subsequent requests

Example Usage
Create a Plan
bashcurl -X POST http://localhost:8080/v1/plan \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer [YOUR_TOKEN]" \
  -d '{
    "planCostShares": {
      "deductible": 2000,
      "_org": "example.com",
      "copay": 23,
      "objectId": "1234vxc2324sdf-5014",
      "objectType": "membercostshare"
    },
    "linkedPlanServices": [
      {
        "linkedService": {
          "_org": "example.com",
          "objectId": "1234520xvc30asdf-502",
          "objectType": "service",
          "name": "Yearly physical"
        },
        "planserviceCostShares": {
          "deductible": 10,
          "_org": "example.com",
          "copay": 0,
          "objectId": "1234512xvc1314asdfs-503",
          "objectType": "membercostshare"
        },
        "_org": "example.com",
        "objectId": "27283xvx9asdff-504",
        "objectType": "planservice"
      }
    ],
    "_org": "example.com",
    "objectId": "12xvxc345ssdsds-508",
    "objectType": "plan",
    "planType": "inNetwork",
    "creationDate": "12-12-2017"
  }'
Search for Plans
bashcurl -X POST http://localhost:8080/v1/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer [YOUR_TOKEN]" \
  -d '{
    "key": "copay",
    "value": "23"
  }'
Advanced Elasticsearch Queries
The system supports various advanced Elasticsearch queries for parent-child relationships:
Get All Plans with a Specific Copay Value
GET /plans/_search
{
  "query": {
    "has_child": {
      "type": "planCostShares",
      "query": {
        "range": {
          "copay": {
            "gte": 2000
          }
        }
      }
    }
  }
}
