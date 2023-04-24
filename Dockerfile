FROM public.ecr.aws/lambda/provided:al2 AS build
ENV CGO_ENABLED=0
# Get rid of the extension warning
RUN mkdir -p /opt/extensions
RUN yum -y install golang
RUN go env -w GOPROXY=direct
# cache dependencies
ADD go.mod go.sum ./
RUN go mod download
COPY . ${LAMBDA_TASK_ROOT}
RUN env GOOS=linux GOARCH=amd64 go build -o=/main

# copy artifacts to a clean image
FROM public.ecr.aws/lambda/provided:al2
COPY --from=build /main /main
RUN yum install git -y
RUN git clone https://github.com/seedspirit/lambda-cicd-go.git
RUN cp lambda-cicd-go/main.go /var/task/
ENTRYPOINT ["./main.go"]
