# Mongodb Docker Compose for testing

WARNING: Do not in production, this is not secure! 

`docker-compose.yml` for testing exposes a mongo port on `27017` (localhost). Credentials are: `root:rootpassword` 

```bash
docker compose --env-file=debug.env up -d
sudo docker volume rm mongodb_data_container //to clean up volume
```