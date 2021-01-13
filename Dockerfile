FROM golang:1.15-alpine AS builder

RUN apk update && apk upgrade && apk --no-cache add curl
RUN apk add build-base

# Copy local source
COPY . $GOPATH/src/github.com/alexsniffin/nr-problem.git/
WORKDIR $GOPATH/src/github.com/alexsniffin/nr-problem.git/

# Run quality check
RUN go test -tags musl -v

# Build binary
RUN GOOS=linux go build -a -o /app/nr-problem .


############# Build the image #############
FROM alpine:3

RUN apk update && apk upgrade

WORKDIR /app/
COPY --from=builder /app/nr-problem .

ENTRYPOINT ["./nr-problem"]