FROM mcr.microsoft.com/playwright:v1.47.2-focal AS builder

RUN apt-get update && apt-get install -y \
    curl \
    ca-certificates \
    wget \
    tar \
    && rm -rf /var/lib/apt/lists/*

RUN wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz && \
    rm go1.21.5.linux-amd64.tar.gz

ENV PATH=$PATH:/usr/local/go/bin


WORKDIR /home/runner/work

COPY . .

RUN go mod download

RUN go install github.com/playwright-community/playwright-go/cmd/playwright@latest

RUN CGO_ENABLED=0 GOOS=linux go build -o test-runner cmd/runRegressionAutomation/runRegressionAutomationMain.go

ENTRYPOINT [ "./test-runner" ]















#FROM golang:alpine AS builder
#WORKDIR /home/runner/work

#COPY . .

#RUN go mod download

#RUN CGO_ENABLED=0 GOOS=linux go build -o test-runner cmd/runRegressionAutomation/runRegressionAutomationMain.go

#RUN go install github.com/playwright-community/playwright-go/cmd/playwright@latest

#FROM mcr.microsoft.com/playwright:v1.47.2-focal

#RUN apt-get update && apt-get install -y \
#    curl \
#    ca-certificates \
#    && rm -rf /var/lib/apt/lists/*

#WORKDIR /home/runner/work
#COPY --from=builder /home/runner/work .
#COPY --from=builder /go/bin/playwright /usr/local/bin/playwright

#RUN playwright install --with-deps



#ENTRYPOINT ["./test-runner"]

