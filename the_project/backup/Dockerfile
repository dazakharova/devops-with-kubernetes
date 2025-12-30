FROM debian:12-slim

RUN apt-get update -y && apt-get install -y --no-install-recommends \
    ca-certificates curl gnupg lsb-release python3 \
 && install -d /etc/apt/keyrings \
 && curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc \
    | gpg --dearmor -o /etc/apt/keyrings/postgresql.gpg \
 && echo "deb [signed-by=/etc/apt/keyrings/postgresql.gpg] http://apt.postgresql.org/pub/repos/apt \
    $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list \
 && apt-get update -y && apt-get install -y --no-install-recommends postgresql-client-16 \
 && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://storage.googleapis.com/pub/gsutil.tar.gz \
    | tar -xz -C /opt \
 && ln -s /opt/gsutil/gsutil /usr/local/bin/gsutil

ENTRYPOINT ["/bin/bash", "-lc"]
