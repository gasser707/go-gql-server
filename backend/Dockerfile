FROM golang:1.16-alpine as builder

WORKDIR /app

# COPY go.mod, go.sum and download the dependencies
COPY go.* ./
RUN go mod download

# COPY All things inside the project and build
COPY . .
RUN go build -o /app/build .

FROM alpine:latest
COPY --from=builder /app/build /app/build

ENTRYPOINT [ "/app/build" ]