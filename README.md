# PathfinderFanWiki - Spell Database
https://pathfinderfanwiki-production.up.railway.app/

A comprehensive database of Pathfinder RPG spells, advanced filtering, and sorting capabilities.

# 🚀 Features
- 📖 Detailed spell information with descriptions

- 🔍 Advanced search and filtering system

- 🗃️ PostgreSQL database backend

- 🐳 Docker containerization

- 📊 Visitor analytics tracking

# 🛠️ Technologies
- Backend: Go (Golang)

- Database: PostgreSQL 15

- Containerization: Docker + Docker Compose

- Templating: HTML templates with custom functions

# ⚙️ Setup & Installation
- Prerequisites
- Docker 20.10+

- Docker Compose 2.0+

- Go 1.23.2+ (optional for local development)

## Quick Start with Docker
- Clone the repository:

```bash
git clone github.com/BirgusL/PathfinderFanWiki.git
```
- Copy the example environment file:

```bash
cp .env.example .env
```
- Edit the .env file with your configuration:

```bash
nano .env
```
- Build and start the containers:

```bash
docker-compose up -d --build
```
- The application will be available at:

```text
http://localhost:8080
```
## Database Migrations
All database migrations are automatically applied on startup through the migrations service. Place your SQL migration files in the migrations/ directory with sequential numbering:

```text
migrations/
  001_initial_schema.sql
  002_spell_data_seed.sql
```
# 📂 Project Structure
```text
pathfinder-wiki/
├── cmd/                  # Main application entrypoint
├── configs/              # Configuration files
├── initdb/               # Schemas creator
├── migrations/           # Database migration files
├── pkg/                  # Core application logic
│   ├── handlers/         # Http hadlers
│   ├── models/           # Data models
│   ├── repositories/     # Database operations
│   └── services/         # Business logic
├── static/               # CSS, JS, images
├── templates/            # HTML templates
├── .env.example          # Environment variables template
├── Dockerfile            # Docker configuration
├── go.mod
├── go.sum
└── docker-compose.yml    # Service orchestration
```
# 🌐 Environment Variables
Configuration is managed through environment variables. See .env.example for all available options:

```ini
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=spells
DB_SSLMODE=disable

# Application
PORT=8080
APP_ENV=development
DEFAULT_LANG=EN
```
# 🐋 Docker Services
##Service	Description	Port
- app	Main application server	8080
- db	PostgreSQL database	5432
- migrations	Automatic database migrations
# 🧑‍💻 Development
Running without Docker
- Start PostgreSQL database:

```bash
docker-compose up -d db
```
- Set up environment variables in .env

- Run the application:

```bash
go run cmd/main.go
```

# 🤝 Contributing
Contributions are welcome! Please follow these steps:

- Fork the repository

- Create your feature branch (git checkout -b feature/AmazingFeature)

- Commit your changes (git commit -m 'Add some AmazingFeature')

- Push to the branch (git push origin feature/AmazingFeature)

- Open a Pull Request

# 📜 License
Distributed under the MIT License. See LICENSE for more information.

# 📧 Contact
Project Maintainer - Kuzmin Anton - kuzmin1a.a@gmail.com
