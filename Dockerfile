# build stage
FROM golang:1.23-alpine3.22 AS builder
WORKDIR /app
COPY . .
RUN go build -o bank main.go 

# run stage
FROM alpine:3.22
WORKDIR /app 
COPY --from=builder /app/bank .

EXPOSE 8080
CMD [ "/app/bank" ]