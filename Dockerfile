FROM golang:1.18-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY . .

RUN go build -o bin/main main.go

RUN sh ./scripts/prepare_env.sh

FROM alpine:3.13 as production

WORKDIR /bin

RUN apk add --no-cache bash

COPY --from=builder /app/bin/main main

COPY --from=builder /app/dao/sqldb/migrations/ app/dao/sqldb/migrations/

COPY --from=builder /app/.repo.env .repo.env

CMD ["/bin/main"]