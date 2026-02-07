# User Management Service

A robust RESTful API service for managing users, built with Go and Gin. This service provides secure authentication, user management capabilities, and efficient pagination.

## Getting Started

### Prerequisites
- Go 1.25 
- PostgreSQL
- Redis

### Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd user-management
   ```

2. Set up configuration (ensure your database is running and configured).

3. Run the migrations:
   ```bash
   go run scripts/migrate.go
   ```

4. Start the server:
   ```bash
   go run cmd/server/main.go
   ```

The server will start on port `8082` (default).

## Running with Docker

Alternatively, you can run the entire stack using Docker Compose:

```bash
docker-compose up --build
```

This will start the application, PostgreSQL, and Redis. The API will be available at `http://localhost:8082`.


## API Documentation

### Authentication
All protected endpoints require a Bearer token in the `Authorization` header.
```
Authorization: Bearer <your_access_token>
```

### Base URL
`http://localhost:8082/api/v1`

### Endpoints

#### Health Check
- **GET** `/health`
- Checks if the service is running.

#### Authentication

##### 1. Sign Up
- **POST** `/auth/signup`
- Register a new user.
- **Body**:
  ```json
  {
    "email": "user@example.com",
    "password": "securepassword123"
  }
  ```
- **Response** (201 Created):
  ```json
  {
    "message": "User registered successfully",
    "data": { ...user_details }
  }
  ```

##### 2. Sign In
- **POST** `/auth/signin`
- Authenticate a user and receive tokens.
- **Body**:
  ```json
  {
    "email": "user@example.com",
    "password": "securepassword123"
  }
  ```
- **Response** (200 OK):
  ```json
  {
    "access_token": "eyJhbG...",
    "refresh_token": "eyJhbG..."
  }
  ```

#### User Management
*Requires Authentication*

##### 1. List Users
- **GET** `/users`
- Retrieve a paginated list of users.
- **Query Parameters**:
  - `limit`: Number of results (default: 20, max: 100)
  - `last_id`: Cursor for pagination (ID of the last user from previous page)
  - `email`: Search users by email (partial match)
- **Example**: `GET /users?limit=10&email=test`
- **Response**:
  ```json
  {
    "users": [
      {
        "id": "uuid-string",
        "email": "user@example.com",
        "created_at": "timestamp"
      }
    ],
    "pagination": {
      "limit": 10,
      "next_cursor": "uuid-string-of-last-item"
    }
  }
  ```

##### 2. Get User Details
- **GET** `/users/:id`
- Get extended details for a specific user.
- **Response**:
  ```json
  {
    "id": "uuid-string",
    "email": "user@example.com",
    "created_at": "timestamp",
    "updated_at": "timestamp"
  }
  ```

##### 3. Update User
- **PUT** `/users/:id`
- Update user information.
- **Body** (all fields optional):
  ```json
  {
    "email": "newemail@example.com",
    "password": "newpassword123"
  }
  ```
- **Response** (200 OK):
  Updated user object.
