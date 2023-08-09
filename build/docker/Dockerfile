FROM rucciva/golang:1.14.2-alpine AS builder

WORKDIR /usr/src/giteaty
COPY . .
RUN make generate && make build


FROM alpine:3.18.3
LABEL maintainer="rucciva@gmail.com"

RUN apk add --no-cache ca-certificates
COPY --from=builder /usr/src/giteaty/bin/giteaty /usr/local/bin/giteaty
EXPOSE 389
ENTRYPOINT ["/usr/local/bin/giteaty"]

