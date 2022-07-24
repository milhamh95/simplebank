# Build stage
FROM golang:1.18 AS builder
WORKDIR /app

# first dot, copy everything from the current folder
# second dot, current working directory inside the image
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -v -o simplebank main.go

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz

# Run stage
FROM alpine
WORKDIR /app
COPY --from=builder /app/simplebank ./simplebank
COPY --from=builder /app/migrate ./migrate
COPY app.env .
COPY start.sh .
COPY wait-for.sh .
COPY db/migration ./migration

EXPOSE 8080
CMD [ "/app/simplebank" ]
ENTRYPOINT [ "/app/start.sh" ]

