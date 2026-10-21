# Go-APIs
Developing APIs Bookstore Management System using Go framework.

## Dependencies
For developing the project, we used the following dependencies:
1. MongoDB database for storing data.
2. Gorilla mux for routing the incoming calls.
3. Keycloak (open source) as the identity provider for user credentials, signup/signin and role based authorization.
4. `github.com/golang-jwt/jwt/v5` for offline validation of Keycloak access tokens.

## Installation
1. Clone the repository:
    ```sh
    git clone https://github.com/yourusername/Go-Learning.git
    ```
2. Navigate to the project directory:
    ```sh
    cd BookManagementSystem
    ```
3. Install the dependencies:
    ```sh
    go mod tidy
    ```

## Usage
1. Set environment variables:
    - `MONGODB_URI` (required)
    - `MONGODB_DB_NAME` (optional, default: `test`)
    - `MONGODB_COLLECTION` (optional, default: `bookstore`)
    - `PORT` (optional, default: `8080`)

    Keycloak settings:
    - `KEYCLOAK_BASE_URL` (optional, default: `http://localhost:8081`) - public URL, must match the `iss` claim
    - `KEYCLOAK_INTERNAL_BASE_URL` (optional, defaults to `KEYCLOAK_BASE_URL`) - URL the API uses for outbound calls
    - `KEYCLOAK_REALM` (optional, default: `bookstore`)
    - `KEYCLOAK_CLIENT_ID` (optional, default: `bookstore-api`)
    - `KEYCLOAK_CLIENT_SECRET` (only for confidential clients)
    - `KEYCLOAK_ADMIN_USERNAME` (optional, default: `admin`)
    - `KEYCLOAK_ADMIN_PASSWORD` (**required for signup** - used against the `master` realm Admin API)
    - `KEYCLOAK_DEFAULT_USER_ROLE` (optional, default: `book-user`) - realm role granted on signup
    - `KEYCLOAK_EXPECTED_AUDIENCE` (optional) - enforce `aud`/`azp` when set
    - `BOOKSTORE_READER_ROLE` (optional, default: `book-user`)
    - `BOOKSTORE_ADMIN_ROLE` (optional, default: `book-admin`)

    Never commit real secrets - pass them through your shell, `.env` file or secret manager.

    PowerShell example:
    ```powershell
    $env:MONGODB_URI="mongodb+srv://admin:<password>@cluster0.urh68bm.mongodb.net/"
    $env:MONGODB_DB_NAME="test"
    $env:MONGODB_COLLECTION="bookstore"
    $env:KEYCLOAK_BASE_URL="http://localhost:8081"
    $env:KEYCLOAK_ADMIN_PASSWORD="<keycloak-admin-password>"
    ```

2. Start Keycloak and MongoDB (free/open source Keycloak distribution):
    ```powershell
    $env:KEYCLOAK_ADMIN_PASSWORD="<keycloak-admin-password>"
    docker compose up -d mongo keycloak
    ```
    This imports `keycloak/realm-export.json`, which creates:
    - realm `bookstore`
    - public client `bookstore-api` with direct access grants enabled
    - realm roles `book-user` and `book-admin`
    - a seed administrator `bookadmin` (change the password immediately)

    The Keycloak admin console is available at `http://localhost:8081`.

3. Start the server:
    ```sh
    go run main.go
    ```
4. Open the API documentation in your browser:
    - Swagger UI: `http://localhost:8080/swagger/` (use the **Authorize** button to paste your access token)
    - OpenAPI spec: `http://localhost:8080/openapi.yaml`
5. Use the following API endpoints to interact with the system:

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