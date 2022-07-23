# Build stage
FROM golang:1.18 AS builder
WORKDIR /app

# first dot, copy everything from the current folder
# second dot, current working directory inside the image
COPY . .

RUN go build -race -o main main.go

# Run stage
FROM alpine
WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8080
CMD ["/app/main"]
