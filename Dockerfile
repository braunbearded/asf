FROM debian:bookworm

# Install dependencies
RUN apt-get update && apt-get install -y \
  curl \
  wget \
  git \
  apt-transport-https \
  ca-certificates \
  gnupg \
  lsb-release \
  software-properties-common \
  fzf \
  sudo \
  && rm -rf /var/lib/apt/lists/*

# -----------------------------------------
# Create non-root user (UID=1000, GID=1000)
# -----------------------------------------
RUN groupadd -g 1000 developer && \
  useradd -m -u 1000 -g 1000 developer && \
  echo "developer ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

USER root

# -----------------------------------------
# Install Azure CLI
# -----------------------------------------
RUN curl -sL https://aka.ms/InstallAzureCLIDeb | bash

# -----------------------------------------
# Install Go
# -----------------------------------------
ENV GO_VERSION=1.23.0
RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
  tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
  rm go${GO_VERSION}.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"

# Switch to the non-root user
USER developer

# Working directory
WORKDIR /workspace

