FROM hashicorp/terraform:1.13.5 AS terraform
FROM registry.k8s.io/kubectl:v1.35.0 AS kubectl
FROM public.ecr.aws/aws-cli/aws-cli:2.27.41 AS awscli

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends bash ca-certificates git jq \
    && rm -rf /var/lib/apt/lists/*
COPY --from=terraform /bin/terraform /usr/local/bin/terraform
COPY --from=ghcr.io/opentofu/opentofu:1.13.9-minimal /usr/local/bin/tofu /usr/local/bin/tofu
COPY --from=kubectl /bin/kubectl /usr/local/bin/kubectl
COPY --from=awscli /usr/local/aws-cli /usr/local/aws-cli
RUN ln -s /usr/local/aws-cli/v2/current/bin/aws /usr/local/bin/aws
ENV TF_IN_AUTOMATION=true TF_INPUT=false HOME=/tmp AWS_EC2_METADATA_DISABLED=true
USER 10001:10001
ENTRYPOINT ["bash"]
