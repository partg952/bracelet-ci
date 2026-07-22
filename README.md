<div align="center">

# BraceletCI

### A distributed, event-driven Continuous Integration platform built in Go

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go&logoColor=white)
![Redis](https://img.shields.io/badge/Redis-Queue-DC382D?style=flat-square&logo=redis&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Database-4169E1?style=flat-square&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Builds-2496ED?style=flat-square&logo=docker&logoColor=white)

</div>

BraceletCI is a self-hosted CI system designed with scalability and service isolation in mind. Instead of a monolithic architecture, BraceletCI separates responsibilities into dedicated services that communicate through events and APIs.

## ✨ Features

- Event-driven architecture
- Distributed worker execution
- Docker-based isolated builds
- GitHub webhook integration
- Real-time pipeline updates
- Redis-backed job queue
- PostgreSQL persistence
- Service-oriented architecture
- Designed for horizontal scaling

---

## 🏗️ Architecture

<div align="center">
  <img
    src="docs/assets/braceletci-architecture.png"
    alt="BraceletCI system architecture"
    width="100%"
  />
  <br />
  <sub>BraceletCI request, queue, worker, and database flow</sub>
</div>

<br />

```text
                    GitHub
                       │
                 Webhook Event
                       │
                       ▼
                BraceletCI API
                       │
             Create Job Event
                       │
                       ▼
                   DB Service
                       │
             Persist Job Record
                       │
                 Push Job ID
                       │
                    Redis Queue
                       │
              ┌────────┴────────┐
              ▼                 ▼
         Worker #1         Worker #2
              │                 │
      Clone Repository   Clone Repository
              │                 │
       Build Docker Image Build Docker Image
              │                 │
       Execute Pipeline  Execute Pipeline
              │                 │
       Send Events via HTTP
              │
              ▼
           DB Service
              │
      Update Database
              │
      Server Side Events 
              │
              ▼
          Dashboard
```

---

## 🧩 Services

### BraceletCI API Service

The `bracelet-ci` service is the main API entry point. It detects repository
pushes through GitHub webhooks, creates a CI job, and queues that job for a
worker.

Responsible for:

- Receiving GitHub push events at `POST /webhook`
- Reading the repository clone URL and pushed commit SHA
- Creating a job through the DB Service
- Queueing the job ID in Redis for a worker

---

### Worker Service

Responsible for:

- Polling Redis
- Cloning repositories
- Building Docker images
- Executing pipelines
- Sending execution events

Workers are completely stateless and can be horizontally scaled.

---

### DB Service

Responsible for:

- Database operations
- Event processing
- Relational queries
- WebSocket updates
- Job metadata
- Project metadata
- User metadata

The DB Service is the single owner of the PostgreSQL database.

---

## 🛠️ Tech Stack

| Component | Technology |
| :--- | :--- |
| Language | **Go** |
| Queue | **Redis** |
| Database | **PostgreSQL** |
| Containers | **Docker** |
| Realtime | **SSE** |
| Version Control | GitHub |
| API | REST |

---

## 📁 Project Structure

```text
bracelet-cicd/
├── cmd/
│   ├── bracelet-ci/            # Main API and GitHub push webhook
│   │   └── main.go
│   ├── bracelet-DB-service/    # Database service
│   │   └── main.go
│   └── bracelet-worker/        # CI worker service
│       └── main.go
│
├── internal/
│   ├── bracelet-DB-service/
│   │   ├── db/                 # PostgreSQL connection and queries
│   │   ├── db-actions/         # User, project, and job operations
│   │   ├── models/             # Database entity models
│   │   └── parser/             # Incoming event parser
│   │
│   └── bracelet-worker/
│       ├── docker-executors/   # Docker image build and execution
│       ├── repository/         # Git repository cloning
│       ├── testrunner/         # Pipeline test execution
│       └── worker/             # Redis queue consumer
│
├── .env                        # Local environment variables
├── go.mod                      # Go module definition
├── go.sum                      # Dependency checksums
└── README.md
```

---

## 🔄 Pipeline Flow

1. User registers a GitHub repository.
2. GitHub sends the push event to the BraceletCI API webhook.
3. BraceletCI API creates a job through the DB Service.
4. DB Service stores the job.
5. Job ID is pushed into Redis.
6. A worker picks up the job.
7. Repository is cloned.
8. Docker image is built.
9. Pipeline commands are executed.
10. Worker streams logs and status updates.
11. DB Service updates PostgreSQL.
12. Dashboard receives real-time updates.

---

## ⚡ Event-Driven Design

BraceletCI uses an event-driven write model.

Examples of events include:

- Create User
- Create Project
- Create Job

Writes are processed through events, while relational reads are served synchronously by the DB Service.

---

## 📈 Scalability

BraceletCI is designed so that every component can be scaled independently.

- Multiple BraceletCI API instances
- Multiple workers
- Shared Redis queue
- Shared PostgreSQL database
- Stateless workers
- Independent DB Service

---

## 🗺️ Roadmap

- [x] Distributed workers
- [x] Docker execution
- [x] Event-driven DB Service
- [x] Real-time dashboard
- [ ] GitHub App authentication
- [ ] Pipeline artifacts
- [ ] Pipeline caching
- [ ] Parallel job execution
- [ ] Secrets management
- [ ] Multi-stage pipelines
- [ ] Pipeline templates
- [ ] Kubernetes executor

---

## 💡 Inspiration

BraceletCI draws inspiration from modern CI/CD systems such as:

- GitHub Actions
- GitLab CI
- Buildkite
- Drone CI

while exploring a service-oriented, event-driven architecture implemented entirely in Go.

---

## 📄 License

MIT
