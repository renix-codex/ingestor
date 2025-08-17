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


## **API Endpoints**

### **GET** /healthz

Purpose: Liveness probe.
Response: 200 OK with body ok.

### **GET** /posts

Return ingested posts.

***Query parameters***

userId (optional, int): If provided, returns posts for that user only.

***Responses***

200 OK

```
{
  "items": [
    {
      "userId": 1,
      "id": 1,
      "title": "sunt aut facere repellat provident occaecati",
      "body": "quia et suscipit...",
      "ingested_at": "2025-08-17T02:03:04Z",
      "source": "placeholder_api"
    }
  ]
}
```
400 Bad Request — invalid query (e.g., userId not an integer)

500 Internal Server Error — storage/read failure

***Examples***

All recent (paginated):
```
curl "http://localhost:8080/posts?limit=20&offset=0"
```

Only for a user:
```
curl "http://localhost:8080/posts?userId=1"
```

Note: All HTTP calls are routed through the API layer (internal/api) which delegates to the ingest service.

## **Transformation Logic**

Input (from upstream API):
```
{
  "userId": <int>,
  "id": <int>,
  "title": <string>,
  "body": <string>
}
```

### Enrichment (ingest.Enrich)

Adds:
ingested_at: set at processing time, UTC (time.Now().UTC() via injected clock).

source: static string from config/env (SOURCE_NAME, e.g., "placeholder_api").

Passes through the original fields unchanged (userId, id, title, body).

### Determinism & idempotency

Storage uses upsert on (user_id, id); reruns won’t create duplicates.

On re-ingest, ingested_at is updated to the new run time (documented behavior).

If you need “first_seen_at” semantics, add a separate column and only set it on insert.

### Validation & error handling

Non-2xx upstream → error (no writes).
Invalid JSON → error (no writes).
Timeouts → error (configurable HTTP_TIMEOUT).

## Database Schema (PostgreSQL)

Table is auto-created on startup.
```
CREATE TABLE IF NOT EXISTS posts (
  user_id     INT            NOT NULL,
  id          INT            NOT NULL,
  title       TEXT           NOT NULL,
  body        TEXT           NOT NULL,
  ingested_at TIMESTAMPTZ    NOT NULL,   -- always UTC
  source      TEXT           NOT NULL,
  doc         JSONB          NOT NULL,   -- full enriched record for easy retrieval
  PRIMARY KEY (user_id, id)
);

-- Hot filters & search helpers
CREATE INDEX IF NOT EXISTS idx_posts_user_id  ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_doc_gin ON posts USING GIN (doc);
```

### Storage strategy

Primary key: (user_id, id) ensures idempotent upserts.

Typed columns: fast filters/sorts by user_id, id, ingested_at, etc.

doc JSONB: exact copy of the enriched payload for simple reads and ad-hoc queries.

### Write path (upsert)
```
INSERT INTO posts (user_id,id,title,body,ingested_at,source,doc)
VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (user_id,id) DO UPDATE SET
  title=EXCLUDED.title,
  body=EXCLUDED.body,
  ingested_at=EXCLUDED.ingested_at,
  source=EXCLUDED.source,
  doc=EXCLUDED.doc;
```

### Common queries

By user:
```
SELECT doc
FROM posts
WHERE user_id = $1
ORDER BY id ASC;
```

Recent (pagination):
```
SELECT doc
FROM posts
ORDER BY ingested_at DESC, id DESC
LIMIT $1 OFFSET $2;
```

### Notes

All timestamps are stored as TIMESTAMPTZ and must be UTC.

No manual migrations are required; schema is bootstrapped at service start.

For very large volumes, consider additional indexes (e.g., on ingested_at) and/or partitioning by time.
