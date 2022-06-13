# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.18.3 AS build

WORKDIR /app

# download the required Go dependencies
COPY go.mod ./
COPY go.sum ./
RUN go mod download
#COPY *.go ./
COPY . ./

RUN CGO_ENABLED=0 go build -o /go-app

##
## Deploy
##
FROM mgoltzsche/podman:rootless



# RUN adduser --disabled-password --gecos '' newuser \
#     && adduser newuser sudo \
#     && echo '%sudo ALL=(ALL:ALL) ALL' >> /etc/sudoers
RUN apk add bash git curl dos2unix
RUN mkdir -p /etc/containers/registries.conf.d
RUN cat > /etc/containers/registries.conf.d/myregistry.conf <<EOF
[[registry]]
location = "docker-registry:5006"
insecure = true

EOF

# RUN chown newuser /newfolder
# USER newuser
WORKDIR /application

COPY --from=build /go-app /application/bat

EXPOSE 8080


ENTRYPOINT ["/application/bat"]