FROM golang:1.26.1-bookworm

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    make \
    build-essential \
  && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["make", "test"]
