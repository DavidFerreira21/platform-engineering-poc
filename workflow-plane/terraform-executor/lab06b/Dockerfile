FROM openpolicyagent/opa:1.11.0-static AS opa
FROM hashicorp/terraform:1.13.5 AS terraform
FROM python:3.13-slim
RUN apt-get update && apt-get install -y --no-install-recommends git ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=opa /opa /usr/local/bin/opa
COPY --from=terraform /bin/terraform /usr/local/bin/terraform
COPY control-plane/terraformstack/controller/requirements.txt /app/
RUN pip install --no-cache-dir -r /app/requirements.txt
COPY control-plane/terraformstack/controller/real.py /app/
COPY workflow-plane/terraform-executor/lab06b/executor.py workflow-plane/terraform-executor/lab06b/policy.rego /app/
ENV TF_IN_AUTOMATION=true TF_INPUT=false HOME=/tmp
USER 10001:10001
ENTRYPOINT ["python", "/app/executor.py"]
