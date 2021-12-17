# Build Go API Server
FROM golang:1.17-alpine AS go_builder
RUN go version
ARG BUILD_VERSION
ARG SOURCE_VERSION
ADD . /app
WORKDIR /app
RUN go build -ldflags="-X 'nt-folly-xmaxx-comp/internal/pkg/build.Version=$BUILD_VERSION' -X 'nt-folly-xmaxx-comp/internal/pkg/build.BuildHash=$SOURCE_VERSION'" -o /main cmd/serve/main.go

# Final stage build, this will be the container
# that we will deploy to production
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=go_builder /main ./

# Execute Main Server
RUN adduser -D follyteam
USER follyteam
CMD ./main service --api_addr ":$PORT" --prod
