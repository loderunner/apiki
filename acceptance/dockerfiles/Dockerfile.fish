FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    curl \
    fish \
    tar \
    gzip \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /home/testuser
RUN useradd -m -s /usr/bin/fish testuser && \
    mkdir -p /home/testuser/.config/fish && \
    touch /home/testuser/.config/fish/config.fish && \
    chown -R testuser:testuser /home/testuser

ENV SHELL=/usr/bin/fish

USER testuser

CMD ["sleep", "infinity"]
