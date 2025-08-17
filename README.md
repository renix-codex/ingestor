# Ingestor Service

A simple data ingestion service that fetches posts from a public API, transforms them, and stores them in Postgres.  

---

## **Run Locally**

### **Option A — Docker Compose (Postgres + App)**

Make sure `docker-compose.yml` is at the project root.  
If your `Dockerfile` is in `docker/`, your compose service should look like:

```yaml
services:
  ingestor:
    build:
      context: .                   # repo root
      dockerfile: docker/Dockerfile
    environment:
      HTTP_LISTEN_ADDR: ":8080"
      HTTP_TIMEOUT: "8s"
      SOURCE_URL: "https://jsonplaceholder.typicode.com/posts"
      SOURCE_NAME: "placeholder_api"
      PG_HOST: postgres
      PG_PORT: 5432
      PG_USER: app
      PG_PASSWORD: app
      PG_DATABASE: ingestor
      PG_SSLMODE: disable
    ports: ["8080:8080"]

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: app
      POSTGRES_PASSWORD: app
      POSTGRES_DB: ingestor
    ports: ["5432:5432"]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U app -d ingestor"]
      interval: 5s
      timeout: 3s
      retries: 20
```

From the project root:
```
docker compose up --build
```

Verify:
```
curl http://localhost:8080/healthz
curl "http://localhost:8080/posts?userId=1"
```

Schema creation: the app creates the posts table/indexes on startup (no manual migration needed).

## **Deploy to Cloud**

The deployment pattern is the same everywhere:

1. Build & push your container to a registry.
2. Provision managed Postgres.
3. Deploy the container (App service / Container service / Cloud Run) with the DB env vars.
4. Store secrets in the cloud secret manager where possible.

### **Azure — Container Apps + Azure Database for PostgreSQL**

```
# vars
RG=ingestor-rg
LOC=eastus
ACR=ingestoracr$RANDOM
APP=ingestor-app
PG=ingestor-pg

# resource group & registry
az group create -n $RG -l $LOC
az acr create -n $ACR -g $RG --sku Basic
az acr login -n $ACR

# build & push
docker tag ingestor:alpine $ACR.azurecr.io/ingestor:latest
docker push $ACR.azurecr.io/ingestor:latest

# postgres (public for demo; secure appropriately in real env)
az postgres flexible-server create \
  --resource-group $RG --name $PG --location $LOC \
  --admin-user app --admin-password appPassword123! \
  --public-access 0.0.0.0 --version 16 \
  --storage-size 32 --tier Burstable --sku-name Standard_B1ms

# get PG host
PGHOST=$(az postgres flexible-server show -g $RG -n $PG --query "fullyQualifiedDomainName" -o tsv)

# container apps env
az containerapp env create -g $RG -n ${APP}-env -l $LOC

# deploy app
az containerapp create -g $RG -n $APP --environment ${APP}-env \
  --image $ACR.azurecr.io/ingestor:latest \
  --registry-server $ACR.azurecr.io \
  --target-port 8080 --ingress external \
  --env-vars \
    HTTP_LISTEN_ADDR=":8080" HTTP_TIMEOUT="8s" \
    SOURCE_URL="https://jsonplaceholder.typicode.com/posts" SOURCE_NAME="placeholder_api" \
    PG_HOST=$PGHOST PG_PORT=5432 PG_USER=app PG_PASSWORD=appPassword123! PG_DATABASE=ingestor PG_SSLMODE=require
```

Grab the URL shown by the command and hit /healthz.

Production: put secrets in Azure Key Vault, restrict DB network, or use private networking.
