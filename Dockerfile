# Build stage
FROM golang:1.18 AS builder
WORKDIR /app

# first dot, copy everything from the current folder
# second dot, current working directory inside the image
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -v -o simplebank main.go

# Run stage
FROM alpine
WORKDIR /app
COPY --from=builder /app/simplebank ./simplebank
COPY app.env .
COPY start.sh .
COPY wait-for.sh .
COPY db/migration ./migration

EXPOSE 8080
CMD [ "/app/simplebank" ]
ENTRYPOINT [ "/app/start.sh" ]

