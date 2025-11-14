# EzModel - Visual Database Schema Designer

EzModel empowers developers to visually design database schemas with real-time collaboration, multi-region deployment, and seamless API integration. Think of it as "Figma for databases" - providing an intuitive visual interface for database design teams.

## ✨ Features

- **🎨 Visual Schema Design**: Drag-and-drop interface using @xyflow/svelte
- **👥 Real-time Collaboration**: WebSocket-powered live collaboration with cursors and presence
- **🌎 Multi-Region Deployment**: Deployed across NYC3 and SFO3 for low latency worldwide
- **🗄️ Multi-Database Support**: PostgreSQL, MySQL, SQLite, SQL Server
- **🔄 Auto-save**: Automatic saving of canvas and schema changes
- **🔐 JWT Authentication**: Secure authentication with access and refresh tokens
- **🚀 CI/CD Pipeline**: Automated testing and deployment via GitHub Actions

## 🏗️ Architecture

```
┌─────────────────────────────────────────┐
│   DigitalOcean Spaces CDN (Global)      │
│        Frontend (SvelteKit)             │
└────────────────┬────────────────────────┘
                 │
    ┌────────────┴────────────┐
    │                         │
┌───▼─────────────┐    ┌─────▼────────────┐
│  Region 1 SFO3  │    │  Region 2 NYC3   │
│  Backend x2     │    │  Backend x2      │
│  PostgreSQL     │───→│  Read Replica    │
│  Redis (Upstash)│    │                  │
└─────────────────┘    └──────────────────┘
```

**Latency Performance:**

- SFO3 users: ~5-10ms WebSocket latency
- NYC3 users: ~5-10ms WebSocket latency
- Cross-region sync: ~20-30ms (via Redis Pub/Sub)

## 🚀 Quick Start

### Prerequisites

- Go 1.24.1+
- Node.js 20+
- pnpm
- PostgreSQL 15+
- Redis 7+ (optional, for multi-region)

### Local Development

```bash
docker compose -f docker-compose.dev.yml up -d postgres

# Backend
cd backend
go mod download
go run cmd/api/main.go

# Frontend
cd frontend
pnpm install
pnpm dev
```

### Multi-Region Production Deployment

See [DEPLOYMENT_GUIDE.md](./DEPLOYMENT_GUIDE.md) for complete deployment guide.

```bash
# 1. Setup databases (PostgreSQL + Redis)
# See DIGITALOCEAN_DATABASE_SETUP.md

# 2. Configure GitHub secrets
gh secret set DIGITALOCEAN_ACCESS_TOKEN
gh secret set JWT_SECRET
gh secret set DB_PASSWORD
gh secret set REDIS_PASSWORD
# ... (see DEPLOYMENT_GUIDE.md)

# 3. Push to main - automatic deployment!
git push origin main

# 4. Monitor deployment
gh run watch
```

## 📚 Documentation

### Getting Started

- **[Deployment Guide](./DEPLOYMENT_GUIDE.md)** - Complete deployment walkthrough (45-60 min)
- [Database Setup](./DIGITALOCEAN_DATABASE_SETUP.md) - PostgreSQL & Redis detailed setup
- [Project Documentation](./CLAUDE.md) - Complete codebase overview

### CI/CD & DevOps

- [CI/CD Guide](./.github/CICD_GUIDE.md) - Automated deployment workflows
- [Secrets Setup](./.github/SECRETS_SETUP.md) - GitHub secrets configuration

## 🛠️ Tech Stack

### Backend (Go 1.24.1)

- **Framework**: Chi router
- **Database**: PostgreSQL + GORM
- **WebSocket**: Gorilla WebSocket
- **Cache/Pub-Sub**: Redis (go-redis)
- **Authentication**: JWT (golang-jwt)

### Frontend (SvelteKit + TypeScript)

- **Framework**: SvelteKit (Svelte 5.0)
- **UI**: ShadCN-Svelte + Tailwind CSS
- **Canvas**: @xyflow/svelte
- **HTTP Client**: Axios

### Infrastructure

- **Compute**: DigitalOcean App Platform (multi-region)
- **Database**: DigitalOcean Managed PostgreSQL (primary + replica)
- **Cache**: Upstash Redis (serverless)
- **CDN**: DigitalOcean Spaces with CDN
- **CI/CD**: GitHub Actions

## 💡 Key Features Explained

### Real-time Collaboration

- **WebSocket Hubs** in each region manage local connections
- **Redis Pub/Sub** synchronizes messages across regions
- Users see cursors, table creation, and schema changes instantly
- Sub-10ms latency for local users

### Multi-Region Architecture

- **Region 1 (SFO3)**: Primary database + Redis (Upstash)
- **Region 2 (NYC3)**: Read replica for fast local reads
- **Global CDN**: Frontend served from edge locations
- Automatic failover if replica goes down

### Automated Deployment

- Push to `main` → automatic deployment
- Tests run first (backend + frontend)
- Deploys to both regions in parallel
- Health checks and auto-rollback on failure
- ~15-20 minute total deployment time

## 🧪 Testing

```bash
# Backend tests
cd backend
go test ./...
go test -cover ./...

# Frontend tests
cd frontend
pnpm test
pnpm check

# Build verification
cd backend && go build -o bin/ezmodel cmd/api/main.go
cd frontend && pnpm build
```

## 📊 Performance & Cost

### Performance

- **WebSocket latency**: 5-10ms (within region)
- **Database reads**: 5-10ms (local replica)
- **Database writes**: 50ms (to primary)
- **Frontend load**: <1s (global CDN)

### Cost (Monthly)

| Service                        | Cost    |
| ------------------------------ | ------- |
| App Platform (2 instances)     | $24     |
| PostgreSQL (primary + replica) | $60     |
| Upstash Redis (Fixed Plan)     | $10     |
| Spaces CDN                     | $5      |
| **Total**                      | **$99** |

## 🤝 Contributing

We welcome contributions! Please see our contributing guidelines.

1. Fork the repository
2. Create feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'feat: add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open Pull Request

## 👥 Team

Meet the Pioneers:

- **Frank Dai** - Project Lead
- **Anson Zhong** - Backend Developer
- **Johnson Wang** - Full Stack Developer
- **Eric Weng** - Frontend Developer

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Resources

- [Live Demo](https://ezmodel-frontend.nyc3.cdn.digitaloceanspaces.com) (coming soon)
- [Documentation](./CLAUDE.md)
- [API Docs](./CLAUDE.md#api-endpoints-implemented)
- [GitHub Issues](https://github.com/your-org/ezmodel/issues)

## 🙏 Acknowledgments

- Built with [SvelteKit](https://kit.svelte.dev/)
- UI components from [ShadCN-Svelte](https://www.shadcn-svelte.com/)
- Visual canvas powered by [@xyflow/svelte](https://svelteflow.dev/)
- Deployed on [DigitalOcean](https://www.digitalocean.com/)

---

**Made with ❤️ by the EzModel team**
