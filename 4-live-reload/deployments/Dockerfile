# WORKBENCH ###################################################
FROM golang:1.14.1 AS workbench
ENV ENVIRONMENT development
WORKDIR /app

RUN apt-get update
RUN apt-get install -y entr

COPY go.mod /app
COPY go.sum /app
RUN go mod download

COPY . /app
RUN go build .

CMD find /app/tilt-example-go /app/web | entr -r -d /app/tilt-example-go


# ARTIFACT ####################################################
FROM alpine AS deploy
ENV ENVIRONMENT production
WORKDIR /app

COPY --from=workbench /app/tilt-example-go /app/tilt-example-go

CMD /app/tilt-example-go
