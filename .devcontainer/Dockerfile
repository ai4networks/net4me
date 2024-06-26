FROM ghcr.io/willfantom/devcontainer-go:latest

RUN apt-get update -q && \
  apt-get install --no-install-recommends -yqq \
  curl \
  hping3 \
  iperf3 \
  iproute2 \
  iptables \
  iputils-ping \
  net-tools \
  openvswitch-common \
  openvswitch-switch \
  openvswitch-testcontroller \
  traceroute && \
  rm -rf /var/lib/apt/lists/*
WORKDIR /
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh
WORKDIR /workspace
ENTRYPOINT [ "bash", "/entrypoint.sh" ]
