# Go-APIs
Developing APIs Bookstore Management System using Go framework.

## Dependencies
For developing the project, we used the following dependencies:
1. MongoDB database for storing data.
2. Gorilla mux for routing the incoming calls.

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

    PowerShell example:
    ```powershell
    $env:MONGODB_URI="mongodb+srv://<username>:<password>@<cluster-url>/?retryWrites=true&w=majority"
    $env:MONGODB_DB_NAME="test"
    $env:MONGODB_COLLECTION="bookstore"
    ```

1. Start the server:
    ```sh
    go run main.go
    ```
2. Use the following API endpoints to interact with the system:

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
- **Invalid ID format**: Returns a `400 Bad Request` error if the ID format is invalid.
- **Book not found**: Returns a `404 Not Found` error if the book is not found in the database.
