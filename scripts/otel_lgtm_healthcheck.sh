#!/bin/sh
set -e

# Check Grafana API
curl -f http://localhost:3000/api/health || exit 1

# Check Prometheus API
curl -f http://localhost:9090/-/ready || exit 1

# Check Loki API
curl -f http://localhost:3100/ready || exit 1

# Check Tempo API
curl -f http://localhost:3200/ready || exit 1

exit 0
