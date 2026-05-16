#!/bin/bash

set -a
source .infra.env
set +a

echo "Waiting for Kafka to come online..."

BROKER=${KAFKA_BROKERS%%,*}

cub kafka-ready -b "$BROKER" 1 20

for TOPIC in $KAFKA_TOPICS; do
    echo "Creating topic: $TOPIC"
    kafka-topics \
      --bootstrap-server "$KAFKA_BROKERS" \
      --topic "$TOPIC" \
      --replication-factor "$KAFKA_REPLICATION_FACTOR" \
      --partitions "$KAFKA_PARTITIONS"\
      --create --if-not-exists
done

echo "Kafka topics created successfully."

