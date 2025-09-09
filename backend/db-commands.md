Local MongoDB setup (dev)
Docker (recommended):
docker run -d --name mcp-mongo -p 27017:27017 -e MONGO_INITDB_ROOT_USERNAME=dev -e MONGO_INITDB_ROOT_PASSWORD=dev mongo:7
export MONGO_URI='mongodb://dev:dev@localhost:27017/?authSource=admin'
export MONGO_DB='mcp'
Native (Ubuntu):
sudo apt-get update && sudo apt-get install -y mongodb-org
sudo systemctl enable --now mongod
Default URI: mongodb://localhost:27017
Test:
mongosh "$MONGO_URI" --eval 'db.runCommand({ping:1})'
MongoDB on Kubernetes (prod-ready)
Use Bitnami MongoDB Helm chart (replica set + auth):
helm repo add bitnami https://charts.bitnami.com/bitnami && helm repo update
helm install mongo bitnami/mongodb --namespace mcp --create-namespace \
--set auth.enabled=true \
--set auth.rootUser=admin \
--set auth.rootPassword=replace-me \
--set architecture=replicaset
Connection string (internal):
mongodb://admin:replace-me@mongo-mongodb-0.mongo-mongodb-headless.mcp.svc.cluster.local:27017,mongo-mongodb-1.mongo-mongodb-headless.mcp.svc.cluster.local:27017/?replicaSet=rs0
Set env in backend deployment:
MONGO_URI (above)
MONGO_DB=mcp
JWT_SECRET=secret
GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, OAUTH_REDIRECT_URL