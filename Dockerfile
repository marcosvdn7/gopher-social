FROM golang:1.24.5 as builder
WORKDIR /app
COPY . .
RUN CGP_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api cmd/api/*.go

FROM scratch
WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/api .
EXPOSE 8080
CMD ["./api"]
