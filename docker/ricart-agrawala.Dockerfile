
#############################################################################################################################
FROM golang:1.19-alpine AS builder
WORKDIR $GOPATH/src/mutualexclusion-project/

COPY . .

RUN go mod download

# Build the binary.
RUN CGO_ENABLED=0 go build -o /main main/main.go

################################################################################################################################
# STEP 2 build a small image
################################################################################################################################
FROM golang:1.19-alpine

ENV NODEID "0"

# Copy our static executable.
COPY --from=builder /main /main

CMD /main --alg=ricartagrawala --node-id=$NODEID