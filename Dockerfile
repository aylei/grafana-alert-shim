FROM golang AS builder

WORKDIR /build
# go mod cache
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build

FROM scratch
COPY --from=builder /build/shim /shim

ENTRYPOINT ["/shim"]
