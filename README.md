# Distributed Docker

A distributed container orchestration system built in Go that manages Docker containers across multiple nodes using etcd for coordination, with automatic load balancing via Traefik and monitoring through Prometheus.

## Features

- **Distributed Scheduling**: Automatically schedules containers on available nodes based on resource capacity
- **Load Balancing**: Integrated Traefik for automatic domain-based routing and SSL termination
- **Monitoring**: Prometheus metrics collection for system resources and container health
- **Cluster Management**: etcd-based coordination for node discovery and state management
- **REST API**: HTTP API for container lifecycle management
- **Resource Awareness**: Intelligent node scoring based on CPU, memory, and disk availability

## Architecture

The system consists of multiple components:

- **Control Plane**: Manages cluster state, scheduling decisions, and API endpoints
- **Worker Nodes**: Run containers and report resource metrics
- **etcd**: Distributed key-value store for coordination
- **Traefik**: Reverse proxy and load balancer with dynamic configuration
- **Prometheus**: Metrics collection and monitoring
- **Grafana**: Dashboard for visualization (optional)

## Prerequisites

- Go 1.25+
- Docker and Docker Compose
- Linux environment with Docker socket access

## Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/shakirmengrani/distributed_docker.git
   cd distributed_docker
   ```

2. **Configure the system**
   Edit `config.yml` to set your node configuration:
   ```yaml
   prefix: "master"
   address: "127.0.0.1:8000"
   control_plane: "127.0.0.1:8000"
   etcd: "http://etcd:2379"
   ```

3. **Start the cluster**
   ```bash
   docker-compose up -d
   ```

4. **Build and run the application**
   ```bash
   go build -o dist-docker
   ./dist-docker
   ```

## API Endpoints

### Health Check
- `GET /health` - Service health status

### Node Information
- `GET /info` - Node resource information and capacity status

### Container Management
- `POST /container` - Create a new container
- `GET /container?name=<name>` - Get container details
- `POST /container/remove` - Remove a container
- `GET /container/list` - List all containers

### Domain Management
- `POST /container/connect` - Connect domain to existing container

### Cluster Management
- `POST /member` - Add a new cluster member
- `GET /member/list` - List all cluster members

### Metrics
- `GET /metrics` - Prometheus metrics endpoint

## Container Creation

Create a container with domain routing:

```json
POST /container
{
  "name": "my-app",
  "domain": ["my-app.example.com"],
  "image": "nginx:latest",
  "port": 80,
  "environments": ["ENV=production"],
  "volumes": {"/data": {}},
  "workingDir": "/app",
  "cmd": ["nginx", "-g", "daemon off;"]
}
```

## Monitoring

Access monitoring interfaces:
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Traefik Dashboard**: http://localhost:8081

## Configuration

### config.yml
```yaml
prefix: "master"                    # Node identifier
address: "127.0.0.1:8000"          # Node listen address
control_plane: "127.0.0.1:8000"    # Control plane address
etcd: "http://etcd:2379"           # etcd endpoint
```

### Docker Compose Services

- **app**: Go application container
- **etcd**: Distributed key-value store
- **prometheus**: Metrics collection
- **grafana**: Visualization dashboard
- **traefik**: Reverse proxy and load balancer

## Development

### Building
```bash
go build -o dist-docker main.go
```

### Testing
```bash
go test ./...
```

### Docker Development
```bash
docker-compose up --build
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.