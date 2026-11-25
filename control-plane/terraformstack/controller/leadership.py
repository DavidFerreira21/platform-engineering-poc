"""Eleição opcional do reconciliador usando uma Lease do Kubernetes."""

import logging
from datetime import datetime, timezone
from time import monotonic, sleep

from kubernetes import client
from kubernetes.client.exceptions import ApiException
from urllib3.exceptions import HTTPError

LEASE_NAME = "terraformstack-controller"
LEASE_SECONDS = 30
RENEW_DEADLINE = 15
RETRY_SECONDS = 2
REQUEST_TIMEOUT = (2, 2)


class Leadership:
    def __init__(self, api, namespace, identity):
        self.api = api
        self.namespace = namespace
        self.identity = identity
        self.observed_version = None
        self.observed_at = 0
        self.deadline = None

    def acquire_or_renew(self):
        started = monotonic()
        try:
            lease = self.api.read_namespaced_lease(
                LEASE_NAME, self.namespace, _request_timeout=REQUEST_TIMEOUT
            )
            version = lease.metadata.resource_version
            now = monotonic()
            if version != self.observed_version:
                self.observed_version = version
                self.observed_at = now
            spec = lease.spec or client.V1LeaseSpec()
            holder = spec.holder_identity
            if self.deadline is not None and holder != self.identity:
                self.deadline = 0
                return False
            duration = spec.lease_duration_seconds or LEASE_SECONDS
            if holder and holder != self.identity and now - self.observed_at < duration:
                return False
            if self.deadline is not None and now >= self.deadline:
                return False

            if holder != self.identity:
                spec.acquire_time = datetime.now(timezone.utc)
                spec.lease_transitions = (spec.lease_transitions or 0) + 1
            spec.holder_identity = self.identity
            spec.lease_duration_seconds = LEASE_SECONDS
            spec.renew_time = datetime.now(timezone.utc)
            lease.spec = spec
            # O PUT leva o resourceVersion do GET e falha se outro candidato venceu.
            self.api.replace_namespaced_lease(
                LEASE_NAME, self.namespace, lease, _request_timeout=REQUEST_TIMEOUT
            )
        except (ApiException, HTTPError) as error:
            if getattr(error, "status", None) != 409:
                logging.warning("Falha ao atualizar Lease: %s", getattr(error, "status", type(error).__name__))
            return False
        now = monotonic()
        if now >= started + RENEW_DEADLINE or (self.deadline is not None and now >= self.deadline):
            return False
        self.deadline = started + RENEW_DEADLINE
        return True

    def acquire(self):
        logging.info("Aguardando liderança em %s/%s", self.namespace, LEASE_NAME)
        while not self.acquire_or_renew():
            sleep(RETRY_SECONDS)
        logging.info("Liderança adquirida pelo Pod %s", self.identity)

    def renew(self, wake):
        while self.active():
            sleep(RETRY_SECONDS)
            if self.active():
                self.acquire_or_renew()
        wake.set()

    def active(self):
        return self.deadline is not None and monotonic() < self.deadline

    def check(self):
        if not self.active():
            logging.error("Prazo de renovação da Lease expirou. Encerrando o reconciliador.")
            raise SystemExit(1)
