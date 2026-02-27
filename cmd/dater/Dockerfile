FROM golang:1.25-alpine AS build

WORKDIR /go/src

COPY go.mod ./
RUN go mod download
COPY . .
RUN go install -v ./...

FROM golang:1.25-alpine
COPY --from=build /go/bin/dater /go/dater

CMD ["dater"]