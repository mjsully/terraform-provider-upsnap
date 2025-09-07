FROM golang:alpine

ARG TERRAFORM_VERSION=1.13.1

# Terraform install
RUN apk add --no-cache wget unzip libc6-compat
# RUN wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip
# RUN unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
#     mv terraform /usr/local/bin && \
#     rm terraform_${TERRAFORM_VERSION}_linux_amd64.zip

RUN wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_arm64.zip
RUN unzip terraform_${TERRAFORM_VERSION}_linux_arm64.zip && \
    mv terraform /usr/local/bin && \
    rm terraform_${TERRAFORM_VERSION}_linux_arm64.zip