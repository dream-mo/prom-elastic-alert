# Contributing
Welcome to elastic-alert repository!

## Development Environment

You can contribute to this project from a Windows, macOS or Linux machine. The first step to contributing is ensuring you can run the demo successfully from your local machine.

On all platforms, the minimum requirements are:

- Docker
- Docker Compose v2.0.0+

### Clone Repo
- Clone the elastic-alert repository:

```bash
git clone https://github.com/openinsight-proj/elastic-alert.git
```

### Open Folder

- Navigate to the cloned folder:

```bash
cd elastic-alert/
```

### Run Docker Compose

- Start the demo:

```bash
 cd example
docker compose up -d
```

### Verify the Alerts

Once the images are built and containers are started you can access:

- Alertmanager: [http://localhost:9093/](http://localhost:9093/)
- Exposed Metrics: [http://localhost:9003/metrics](http://localhost:9003/metrics)


## Create Your First Pull Request

TODO