# Go RealWorld Conduit Backend



A Go implementation of the [RealWorld](https://github.com/gothinkster/realworld) backend specification (Medium clone).

Built with: Go, [gorilla/mux](https://github.com/gorilla/mux), [pgx/v5](https://github.com/jackc/pgx), [log/slog](https://pkg.go.dev/log/slog), [Paseto](https://github.com/o1egl/paseto), [go-playground/validator](https://github.com/go-playground/validator), [goose](https://github.com/pressly/goose).

## Features

Implements a subset of the RealWorld API spec, including:

*   User registration & login (Paseto tokens)
*   Get/Update current user
*   Get user profiles
*   Follow/Unfollow users
*   Create articles
*   List articles (filter by author/tag)
*   (Add other implemented features)

## Requirements

*   Go (1.21+ recommended)
*   PostgreSQL
*   [Goose CLI](https://github.com/pressly/goose#install)
*   Docker (optional)

## Configuration

The application requires the following environment variables:

*   `DB_URL`: PostgreSQL connection URL.
    *   Example: `postgres://user:password@host:port/database_name?sslmode=disable`
*   `JWT_SECRET`: A **32-byte** secret key for Paseto token encryption.
    *   Example (for testing only, **use a secure key!**): `ThisIsASecureSecretKeyOf32Bytes!`

## Running Locally

1.  **Clone:**
    ```bash
    git clone https://github.com/YOUR_ACCOUNT/YOUR_REPO.git # Update URL!
    cd YOUR_REPO
    ```
2.  **Set Environment Variables:**
    ```bash
    export DB_URL="your_db_connection_url"
    export JWT_SECRET="your_32_byte_secret_key"
    ```
3.  **Run Migrations:**
    ```bash
    cd db/migrations
    goose postgres "$DB_URL" up
    cd ../..
    ```
4.  **Build:**
    ```bash
    go build -o conduit-backend .
    ```
5.  **Run:**
    ```bash
    ./conduit-backend
    ```
    The server starts on port `:8080` by default.

## Running with Docker

1.  **Build Image:**
    ```bash
    docker build -t conduit-backend:latest .
    ```
2.  **Run Container:**
    ```bash
    docker run -p 8080:8080 \
      -e DB_URL="your_db_connection_url_accessible_from_docker" \
      -e JWT_SECRET="your_32_byte_secret_key" \
      --name conduit-app \
      conduit-backend:latest  
    ```

## Testing

Requires `DB_URL` and `JWT_SECRET` to be set.

```bash
go test ./...

##NOTE:
** This project was developed based on a public homework assignment. The integration test suite (app_test.go) was provided by the instructor. The application logic and implementation are by the repository author. **
