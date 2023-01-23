# builder image
FROM golang:alpine3.16 as builder
RUN mkdir /build
ADD *.go /build/
ADD go.mod /build/
WORKDIR /build
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -o word-sequences .


# final image
FROM alpine:3.16
COPY --from=builder /build/word-sequences .

# executable
ENTRYPOINT [ "./word-sequences" ]