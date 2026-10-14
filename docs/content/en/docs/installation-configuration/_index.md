
---

description: Everything you need to install and configure SOARCA
title: Advanced Installation and Configuration
categories: [documentation, configuration]
tags: [extension, security]
weight: 4
---

After completing the [Getting Started](/docs/getting-started/_index.md) setup for SOARCA, you may find that certain advanced configurations or customizations are necessary to optimize SOARCA for your specific use cases, for example integrating with The Hive. This section provides in-depth guidance on additional steps you can take to enhance, secure, and integrate SOARCA with your infrastructure, ensuring it meets your unique operational needs.

### Configuring SOARCA

| Variable                   | Content                          | Description                                                                 |
|----------------------------|-----------------------------------|-----------------------------------------------------------------------------|
| PORT                       | `8080`                           | Set the exposed port of SOARCA. Default is `8080`.                          |
| ENABLE_TLS                 | `false`                          | Enable TLS for secure communication. Default is `false`.                    |
| CERT_FILE                  | `"/certs/server.crt"`            | Path to the TLS certificate file. Default is `"/certs/server.crt"`.         |
| CERT_KEY_FILE              | `"/certs/server.key"`            | Path to the TLS certificate key file. Default is `"/certs/server.key"`.     |
| MAX_EXECUTIONS             | `1000`                           | The number of historical executions saved, including the current one. Default is `10`. |
| SOARCA_ALLOWED_ORIGINS     | `*`                              | Set allowed origins for cross-origin requests. Default is `*`.              |
| GIN_MODE                   | `release`                        | Set the GIN mode. Default is `release`.                                     |
| DATABASE                   | `false`                          | Set if you want to run with an external database. Default is `false`.       |
| MONGODB_URI                | `mongodb://localhost:27017`      | Set the MongoDB URI. Default is `mongodb://localhost:27017`.                |
| DATABASE_NAME              | `soarca`                         | Set the MongoDB database name when using Docker. Default is `soarca`.       |
| DB_USERNAME                | `root`                           | Set the MongoDB database user when using Docker. Default is `root`.         |
| DB_PASSWORD                | `rootpassword`                   | Set the MongoDB database user password when using Docker. **Change this in production!** Default is `rootpassword`. |
| PLAYBOOK_API_LOG_LEVEL     | `trace`                          | Set the log level for the playbook API. Default is `trace`.                 |
| MAX_REPORTERS              | `10`                              | Set the maximum number of downstream reporters. Default is `10`.             |
| LOG_GLOBAL_LEVEL           | `info`                           | One of the specified log levels. Default is `info`.                         |
| LOG_MODE                   | `development`                    | Set the logging mode. If `production`, `LOG_GLOBAL_LEVEL` is used for all modules. Default is `development`. |
| LOG_FILE_PATH              | `""`                             | Path to the logfile for all logging. Default is `""` (empty string).        |
| LOG_FORMAT                 | `json`                           | Set the logging format. Either `text` or `json`. Default is `json`.         |
| ENABLE_FINS                | `false`                          | Enable FINS in SOARCA. Default is `false`.                                  |
| MQTT_BROKER                | `localhost`                      | The broker address for SOARCA to connect to for communication with FINS. Default is `localhost`. |
| MQTT_PORT                  | `1883`                           | The port for the MQTT broker. Default is `1883`.                            |
| HTTP_SKIP_CERT_VALIDATION  | `false`                          | Set whether to skip certificate validation for HTTP connections. Default is `false`. |
| VALIDATION_SCHEMA_URL      | `""`                             | Set a custom validation schema to validate playbooks. Default is `""` to use the internal schema. **Note:** Changing this can heavily impact performance. |

-----

### Integrations

#### The Hive

| Variable             | Content                          | Description                                             |
|----------------------|-----------------------------------|---------------------------------------------------------|
| THEHIVE_ACTIVATE     | `false`                          | Enable integration with The Hive. Default is `false`.   |
| THEHIVE_API_TOKEN    | `your_token`                     | Set the API token for The Hive integration.             |
| THEHIVE_API_BASE_URL | `http://your.thehive.instance/api/v1/` | Set the base URL for The Hive API. Default is `""`.      |

-----

### Authentication

{{% alert title="Note" color="primary" %}}
More information on setting up authentication can be found [here](/docs/installation-configuration/authentication/).
{{% /alert %}}
| Variable | Content | Description |
|---|---|---|
| AUTH_ENABLED | `false` | Require a valid credential on every API request. Default is `false`. |
| OIDC_ISSUER | `https://authentik.example.com/application/o/soarca/` | Issuer URL of the OIDC provider (authentik). |
| OIDC_CLIENT_ID | `SOME_CLIENT_ID` | Client ID of the OIDC provider; the expected token audience. |
| OIDC_SKIP_TLS_VERIFY | `false` | Skip certificate checks towards the OIDC provider. Default is `false`. |
| AUTH_GROUP | `soarca_admin` | OIDC group whose members get every permission. Default is `soarca_admin`. |
| AUTH_ENTRA_TENANT_IDS | `<tenant-id>` | Comma-separated Entra ID tenants whose tokens SOARCA accepts. |
| AUTH_ENTRA_API_CLIENT_ID | `<client-id>` | Client ID of the SOARCA API app registration in Entra ID. |
| AUTH_ENTRA_API_APP_ID_URI | `api://soarca` | Application ID URI of that registration. Optional. |
| AUTH_ENTRA_ALLOWED_CLIENT_IDS | `<client-id>,<client-id>` | Applications allowed to call SOARCA. Empty allows all in the tenant. |
| AUTH_ENTRA_REQUIRED_USER_SCOPE | `access_as_user` | Delegated scope user tokens must carry. Default is `access_as_user`. |
| AUTH_ENTRA_ACCEPT_V1 | `false` | Also accept Entra ID v1 tokens. Default is `false`. |
| AUTH_ENTRA_AUTHORITY | `https://login.microsoftonline.com` | Entra ID login host; change only for national clouds. |
| AUTH_API_KEYS_FILE | `/etc/soarca/api-keys.json` | JSON file with hashed API keys. Empty disables API keys. |

-----

#### Key management system
{{% alert title="Note" color="primary" %}}
More information on the key management system can be found [here](/docs/soarca-extensions/key-management-system).
{{% /alert %}}
| Variable               | Content                                    | Description                                                                                 |
|------------------------|---------------------------------------------|---------------------------------------------------------------------------------------------|
| ENABLE_SSH_KMS         | `false`                                    | Enable the key management system for SSH|
| SSH_KMS_DIR            | `deployments/docker/testing/ssh-kms-test/ssh-keystore/`| Set the folder where SSH keys for the key management system are stored|

