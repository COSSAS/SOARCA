
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
| SOARCA_ALLOWED_ORIGINS     | `*`                              | Set allowed origins for cross-origin requests. Default is `*`.              |
| GIN_MODE                   | `release`                        | Set the GIN mode. Default is `release`.                                     |
| DATABASE                   | `false`                          | Set if you want to run with an external database. Default is `false`.       |
| MONGODB_URI                | `mongodb://localhost:27017`      | Set the MongoDB URI. Default is `mongodb://localhost:27017`.                |
| DATABASE_NAME              | `soarca`                         | Set the MongoDB database name when using Docker. Default is `soarca`.       |
| DB_USERNAME                | `root`                           | Set the MongoDB database user when using Docker. Default is `root`.         |
| DB_PASSWORD                | `rootpassword`                   | Set the MongoDB database user password when using Docker. **Change this in production!** Default is `rootpassword`. |
| PLAYBOOK_API_LOG_LEVEL     | `trace`                          | Set the log level for the playbook API. Default is `trace`.                 |
| MAX_REPORTERS              | `5`                              | Set the maximum number of downstream reporters. Default is `5`.             |
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
More information on setting up authentication can be found [here](/docs/installation-configuration/authentication.md).
{{% /alert %}}
**Is your feature request related to a problem? Please describe.**
A clear and concise description of what the problem is. Ex. I'm always frustrated when [...]

**Describe the solution you'd like**
A clear and concise description of what you want to happen.

**Describe alternatives you've considered**
A clear and concise description of any alternative solutions or features you've considered.

**Additional context**
Add any other context or screenshots about the feature request here.
| Variable               | Content                                    | Description                                                                                 |
|------------------------|---------------------------------------------|---------------------------------------------------------------------------------------------|
| AUTH_ENABLED           | `false`                                    | Enable authentication. Default is `false`.                                                  |
| OIDC_ISSUER            | `https://localhost:9443/application/u/test/` | The OIDC issuer URL.                                                                         |
| OIDC_CLIENT_ID         | `SOME_CLIENT_ID`                           | Set the OIDC client ID.                                                                     |
| OIDC_CLIENT_SECRET     | `SOME_CLIENT_SECRET`                       | Set the OIDC client secret.                                                                 |
| OIDC_REDIRECT_URL      | `http://localhost:8081/auth/soarca_gui/callback` | Set the OIDC redirect URL.                                                                 |
| COOKIE_SECRET_KEY      | `SOME_COOKIE_SECRET`                       | Optional: Secret key for cookies. Generate using `openssl rand -base64 32` or `head -c 32 /dev/urandom | base64`. |
| OIDC_SKIP_TLS_VERIFY   | `true`                                     | Set whether to skip TLS verification. Default is `true`.                                    |
| AUTH_GROUP             | `soarca_admin`                             | Specify the group users must belong to for authentication against SOARCA.                  |
