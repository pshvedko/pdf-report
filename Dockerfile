FROM golang AS build
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN go build  .
FROM ubuntu
WORKDIR /app
COPY --from=build /app/pdf-report .
COPY template template
COPY fonts fonts
USER nobody
ENTRYPOINT ["/app/pdf-report"]
