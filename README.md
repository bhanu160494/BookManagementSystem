# Go-APIs
Develope Restful APIs for Bookstore Management System using Go framework.

## Dependencies
In this project, we used the following dependencies:
1. MongoDB database for storing data.
2. Gorilla mux for routing the incoming calls.
3. Keycloak (open source) as the identity provider for user credentials, signup/signin and role based authorization.
4. `github.com/golang-jwt/jwt/v5` for offline validation of Keycloak access tokens.

## Installation and setup

### Prerequisites

Install the following before starting the application:

- Go 1.21 or later
- Docker Desktop with Docker Compose
- PowerShell, Bash, or another shell for setting environment variables

### Environment variables

For the local Docker workflow, set this required variable before starting Compose:

- `KEYCLOAK_ADMIN_PASSWORD` - required by Keycloak and the signup flow

When running the Go API directly on the host, also set:

- `MONGODB_URI` - required by the database package; use `mongodb://localhost:27017`

The following variables are optional and have the listed defaults:

| Variable | Default | Description |
| --- | --- | --- |
| `MONGODB_DB_NAME` | `test` | MongoDB database name |
| `MONGODB_COLLECTION` | `bookstore` | MongoDB collection name |
| `PORT` | `8080` | API port |
| `KEYCLOAK_BASE_URL` | `http://localhost:8081` | Public Keycloak URL; must match the token issuer |
| `KEYCLOAK_INTERNAL_BASE_URL` | `KEYCLOAK_BASE_URL` | URL used by the API for Keycloak requests |
| `KEYCLOAK_REALM` | `bookstore` | Keycloak realm |
| `KEYCLOAK_CLIENT_ID` | `bookstore-api` | Keycloak client ID |
| `KEYCLOAK_CLIENT_SECRET` | empty | Secret for a confidential client, if used |
| `KEYCLOAK_ADMIN_REALM` | `master` | Realm used for Keycloak Admin API login |
| `KEYCLOAK_ADMIN_CLIENT_ID` | `admin-cli` | Admin API client ID |
| `KEYCLOAK_ADMIN_USERNAME` | `admin` | Keycloak administrator username |
| `KEYCLOAK_DEFAULT_USER_ROLE` | `book-user` | Role granted to newly signed-up users |
| `KEYCLOAK_EXPECTED_AUDIENCE` | empty | Optional audience/authorized-party validation |
| `BOOKSTORE_READER_ROLE` | `book-user` | Role allowed to read books |
| `BOOKSTORE_ADMIN_ROLE` | `book-admin` | Role allowed to modify books |

Never commit real secrets. Set them in your shell, a local `.env` file, or a secret manager.

### Run the application with local Docker services

Run these steps from the `BookManagementSystem` directory.

1. Download Go dependencies:
    ```powershell
    go mod tidy
    ```

2. Set the Keycloak administrator password:
    ```powershell
    $env:KEYCLOAK_ADMIN_PASSWORD="<keycloak-admin-password>"
    ```

3. Start MongoDB and Keycloak:
    ```powershell
    docker compose up -d mongo keycloak
    ```
    MongoDB runs from the `mongo:7` image at `mongodb://localhost:27017`. Its data is
    persisted in the Docker volume `mongo-data`.

    Keycloak imports `keycloak/realm-export.json`, which creates realm `bookstore`, client
    `bookstore-api`, roles `book-user` and `book-admin`, and the seed administrator
    `bookadmin` with the password defined in that realm export. Change that password after
    the first login. The admin console is available at `http://localhost:8081`.

4. Configure MongoDB for the host-run Go API:
    ```powershell
    $env:MONGODB_URI="mongodb://localhost:27017"
    ```
    The defaults for `MONGODB_DB_NAME`, `MONGODB_COLLECTION`, `PORT`, and the remaining
    Keycloak variables can be used without setting them explicitly.

5. Start the API:
    ```powershell
    go run main.go
    ```

6. Verify the API documentation:
    - Swagger UI: `http://localhost:8080/swagger/` (use the **Authorize** button to paste your access token)
    - OpenAPI spec: `http://localhost:8080/openapi.yaml`

### Run the complete application in Docker

To run MongoDB, Keycloak, and the API as Docker services instead of starting the Go API on
the host:

1. Set the Keycloak administrator password:
    ```powershell
    $env:KEYCLOAK_ADMIN_PASSWORD="<keycloak-admin-password>"
    ```
2. Build and start all services:
    ```powershell
    docker compose up -d --build
    ```
3. Open the API at `http://localhost:8080`. The containerized API connects to MongoDB at
   `mongodb://mongo:27017` and Keycloak at `http://keycloak:8080`.

To stop the services while keeping MongoDB data:

```powershell
docker compose down
```

To stop the services and delete the local MongoDB data:

```powershell
docker compose down -v
```

## Usage

Use the following API endpoints to interact with the system:

## Authentication and authorization
All book endpoints require an `Authorization: Bearer <access_token>` header. Tokens are issued by
Keycloak and validated offline against the realm JWKS (signature, issuer and expiry are all checked).

| Role | Allowed operations |
| --- | --- |
| `book-user` | `GET /books`, `GET /book/{id}`, `GET /auth/me` |
| `book-admin` | everything above plus `POST /book`, `PUT /book/{id}`, `DELETE /book/{id}` |

### Auth APIs
1. **POST /auth/signup**: Creates the user in Keycloak and grants the default `book-user` role.
    ```json
    {
        "username": "alice",
        "email": "alice@example.com",
        "firstName": "Alice",
        "lastName": "Doe",
        "password": "<at least 8 characters>"
    }
    ```
2. **POST /auth/signin**: Returns `access_token` / `refresh_token` from Keycloak.
    ```json
    { "username": "alice", "password": "<password>" }
    ```
3. **POST /auth/refresh**: Exchanges `{ "refreshToken": "..." }` for a fresh access token.
4. **POST /auth/logout**: Revokes `{ "refreshToken": "..." }` and ends the Keycloak session.
5. **GET /auth/me**: Returns the id, username, email and roles of the caller.

Example:
```powershell
$token = (Invoke-RestMethod -Method Post -Uri http://localhost:8080/auth/signin `
    -ContentType 'application/json' `
    -Body '{"username":"bookadmin","password":"<password>"}').access_token

Invoke-RestMethod -Uri http://localhost:8080/books -Headers @{ Authorization = "Bearer $token" }
```

### APIs demonstration:
1. **GET /books**: Returns a complete set of books with their detailed information.
2. **GET /book/{id}**: Returns details of a particular book.
3. **POST /book**: Stores the book details using the payload below.
    - Payload:
    ```json
    {
        "name": "<BookName>",
        "writer": "WriterName",
        "availability": true/false
    }
    ```
4. **PUT /book/{id}**: Updates the book details.
5. **DELETE /book/{id}**: Deletes the book details from the database.

## Error Handling
- **Missing/invalid token**: Returns a `401 Unauthorized` error.
- **Insufficient role**: Returns a `403 Forbidden` error.
- **Invalid ID format**: Returns a `400 Bad Request` error if the ID format is invalid.
- **Book not found**: Returns a `404 Not Found` error if the book is not found in the database.