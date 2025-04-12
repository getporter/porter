---
title: Migrate Existing MongoDB Data to New Version
description: How to migrate existing data stored in Docker
weight: 8
aliases:
  - /upgrade-mongo-docker
---

{{< callout type="info" >}}
  Only relevant when upgrading from a version before Porter v1.3.0 to Porter v1.3.0
{{< /callout >}}

In Porter v1.3.0, the MongoDB Docker Storage Plugin was updated to use MongoDB v8.0 instead of v4.0. This change was long overdue, and unfortunately, it is not possible to automatically migrate the data across many versions. Going forward, MongoDB will be updated regularly with automatic migration.

## Backup Data

To ensure no data is lost during migration, create a clone of the data using the following script:

```bash
docker volume create porter-mongodb-docker-plugin-data-clone
docker run --rm -v porter-mongodb-docker-plugin-data:/from -v porter-mongodb-docker-plugin-data-clone:/to alpine ash -c "cd /from && cp -av . /to"
```

## Upgrade Process

To perform the upgrade, run the following script:

```bash
#!/bin/bash

# Function to wait for MongoDB to be ready
wait_for_mongodb() {
  local container_name=$1
  local client=$2  # 'mongo' for older versions, 'mongosh' for newer versions
  
  echo "Waiting for MongoDB to be ready in $container_name..."
  until docker exec $container_name $client --eval "db.adminCommand('ping')" > /dev/null 2>&1; do
    echo "MongoDB is not ready yet... waiting"
    sleep 2
  done
  echo "MongoDB is ready!"
}

# Ensure the existing MongoDB container is stopped and removed
echo "Stopping and removing existing MongoDB container..."
docker stop porter-mongodb-docker-plugin 2>/dev/null || true
docker rm porter-mongodb-docker-plugin 2>/dev/null || true

# Upgrade from 4.0 to 4.2
echo "Upgrading from MongoDB 4.0 to 4.2..."
docker pull mongo:4.2-bionic
docker run --name porter-migrate --rm --mount source=porter-mongodb-docker-plugin-data,destination=/data/db -d mongo:4.2-bionic
wait_for_mongodb "porter-migrate" "mongo"
docker exec -it porter-migrate mongo --eval "db.adminCommand( { setFeatureCompatibilityVersion: '4.2' } )"
docker exec -it porter-migrate mongo --eval "db.adminCommand( { shutdown: 1 } )"

# Upgrade from 4.2 to 4.4
echo "Upgrading from MongoDB 4.2 to 4.4..."
docker pull mongo:4.4-focal
docker run --name porter-migrate --rm --mount source=porter-mongodb-docker-plugin-data,destination=/data/db -d mongo:4.4-focal
wait_for_mongodb "porter-migrate" "mongo"
docker exec -it porter-migrate mongo --eval "db.adminCommand( { setFeatureCompatibilityVersion: '4.4' } )"
docker exec -it porter-migrate mongo --eval "db.adminCommand( { shutdown: 1 } )"

# Upgrade from 4.4 to 5.0
echo "Upgrading from MongoDB 4.4 to 5.0..."
docker pull mongo:5.0-focal
docker run --name porter-migrate --rm --mount source=porter-mongodb-docker-plugin-data,destination=/data/db -d mongo:5.0-focal
wait_for_mongodb "porter-migrate" "mongo"
docker exec -it porter-migrate mongo --eval "db.adminCommand( { setFeatureCompatibilityVersion: '5.0' } )"
docker exec -it porter-migrate mongo --eval "db.adminCommand( { shutdown: 1 } )"

# Upgrade from 5.0 to 6.0
echo "Upgrading from MongoDB 5.0 to 6.0..."
docker pull mongo:6.0-jammy
docker run --name porter-migrate --rm --mount source=porter-mongodb-docker-plugin-data,destination=/data/db -d mongo:6.0-jammy
wait_for_mongodb "porter-migrate" "mongosh"
docker exec -it porter-migrate mongosh --eval "db.adminCommand( { setFeatureCompatibilityVersion: '6.0' } )"
docker exec -it porter-migrate mongosh --eval "db.adminCommand( { shutdown: 1 } )"

# Upgrade from 6.0 to 7.0
echo "Upgrading from MongoDB 6.0 to 7.0..."
docker pull mongo:7.0-jammy
docker run --name porter-migrate --rm --mount source=porter-mongodb-docker-plugin-data,destination=/data/db -d mongo:7.0-jammy
wait_for_mongodb "porter-migrate" "mongosh"
docker exec -it porter-migrate mongosh --eval "db.adminCommand( { setFeatureCompatibilityVersion: '7.0', confirm: true } )"
docker exec -it porter-migrate mongosh --eval "db.adminCommand( { shutdown: 1 } )"

# Final upgrade to 8.0
echo "Finalizing upgrade to MongoDB 8.0..."
docker pull mongo:8.0-noble
docker run --name porter-migrate --rm --mount source=porter-mongodb-docker-plugin-data,destination=/data/db -d mongo:8.0-noble
wait_for_mongodb "porter-migrate" "mongosh"
docker exec -it porter-migrate mongosh --eval "db.adminCommand( { setFeatureCompatibilityVersion: '8.0', confirm: true } )"
docker exec -it porter-migrate mongosh --eval "db.adminCommand( { shutdown: 1 } )"

echo "MongoDB upgrade completed successfully!"
```

After running these commands, verify the upgrade was successful by running:

```bash
porter installation list
```

If the command succeeds, the upgrade has been completed successfully.

## Rollback Procedure

If you need to rollback, follow these steps:

1. Restore the data from the backup:
```bash
docker volume rm porter-mongodb-docker-plugin-data
docker volume create porter-mongodb-docker-plugin-data
docker run --rm -v porter-mongodb-docker-plugin-data-clone:/from -v porter-mongodb-docker-plugin-data:/to alpine ash -c "cd /from && cp -av . /to"
```

2. Install a version of Porter lower than v1.3.0, for example v1.2.1:
```bash
export VERSION="v1.2.1"
curl -L https://cdn.porter.sh/$VERSION/install-linux.sh | bash
```
