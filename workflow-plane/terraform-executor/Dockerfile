# Build a partir da raiz do repositório:
# docker build -f workflow-plane/terraform-executor/Dockerfile -t terraformstack-executor:lab08 .
FROM public.ecr.aws/aws-cli/aws-cli:latest AS awscli

FROM python:3.12-slim
ARG TERRAFORM_VERSION=1.13.5
ARG TARGETARCH=amd64

RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl unzip \
    && rm -rf /var/lib/apt/lists/*
COPY --from=awscli /usr/local/aws-cli /usr/local/aws-cli
RUN ln -s /usr/local/aws-cli/v2/current/bin/aws /usr/local/bin/aws \
    && curl --fail --location --silent --show-error \
      "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${TARGETARCH}.zip" \
      -o /tmp/terraform.zip \
    && unzip -q /tmp/terraform.zip terraform -d /usr/local/bin \
    && rm /tmp/terraform.zip \
    && terraform version && aws --version

RUN useradd --uid 10001 --create-home --shell /usr/sbin/nologin executor
COPY foundation/modules/configuration /opt/foundation/modules/configuration
COPY foundation/config /opt/foundation/config
COPY foundation/lab08/validation-stack /opt/foundation/lab08/validation-stack
COPY workflow-plane/terraform-executor/verify.py /opt/terraform-executor/verify.py
RUN mkdir -p /workspace && chown executor:executor /workspace
USER 10001
WORKDIR /workspace
ENV HOME=/home/executor \
    AWS_EC2_METADATA_DISABLED=true \
    TF_IN_AUTOMATION=true
ENTRYPOINT ["python3", "/opt/terraform-executor/verify.py"]
