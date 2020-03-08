# Build
FROM golang:alpine as builder

# create nobody user
RUN echo "nobody:x:1:1:nobody:/:" >> /etc/passwd
RUN echo "nobody:x:1:" >> /etc/group

# fetch tz data
RUN apk --no-cache add tzdata

# fetch ca certs
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

# build app
WORKDIR /app-nix-64
ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -o rate-api


# Final
FROM scratch as final

# copy user info
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# copy tz info
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=America/Chicago

# copy ca info
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# copy binary
COPY --from=builder /app-nix-64/rate-api /bin/rate-api

USER nobody
ENTRYPOINT [ "rate-api" ]