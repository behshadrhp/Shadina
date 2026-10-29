# Shadina — English Documentation

This documentation set describes the **verified implementation** of the Shadina repository at commit
`6c14149` (`:zap: add: Authorization Layer`), reviewed on **2026-10-29**. It is written from source code,
SQL migrations, configuration files, and read-only runtime checks against the local development database.
It does not describe features that are not implemented.

## Start here

| Document | What it answers |
|---|---|
| [01. Project overview](01-project-overview.md) | What the system does, what it does not do, and which technologies it uses |
| [02. Architecture](02-architecture.md) | Packages, responsibilities, dependencies, and architectural decisions |
| [03. Project structure](03-project-structure.md) | Every important file and directory, annotated |
| [04. Application flow](04-application-flow.md) | Startup, middleware order, and the request lifecycle |
| [05. Database design](05-database-design.md) | Tables, columns, constraints, indexes, and ERD |
| [06. Database operations](06-database-operations.md) | Migrations, seed, queries, transactions, and database risks |
| [07. API reference](07-api-reference.md) | Every real route, payload, response, status code, and quirk |
| [08. Authentication and authorization](08-authentication-authorization.md) | JWT issuance/validation, roles, and access control |
| [09. Configuration](09-configuration.md) | Environment variables, defaults, secrets, and hard-coded settings |
| [10. Error handling](10-error-handling.md) | Response envelope, internal codes, and error mapping |
| [11. Logging and observability](11-logging-observability.md) | Zap, request logs, log files, and missing telemetry |
| [12. Docker and deployment](12-docker-deployment.md) | What the Compose file actually runs and how to deploy locally |
| [13. Development guide](13-development-guide.md) | Setup, migrations, seed, run, build, and test commands |
| [14. Testing and quality](14-testing-quality.md) | What was tested, what passed, what failed, and what is missing |
| [15. Security review](15-security-review.md) | Verified findings, severity, evidence, and remediation |
| [16. Troubleshooting](16-troubleshooting.md) | Common failures and safe diagnostic steps |
| [17. Glossary](17-glossary.md) | Bilingual terminology |

Persian documentation: [فارسی](../fa/README.md).

## Audit reports

- [Documentation audit](../reports/documentation-audit.md)
- [Architecture review](../reports/architecture-review.md)
- [Database review](../reports/database-review.md)
- [API consistency review](../reports/api-consistency-review.md)
- [Security review](../reports/security-review.md)
- [Documentation coverage](../reports/documentation-coverage.md)

## Diagram sources

All diagrams are stored as Mermaid source files so they can be reviewed and regenerated:

- [System context](diagrams/system-context.mmd)
- [Component architecture](diagrams/architecture.mmd)
- [Package dependency graph](diagrams/package-dependency.mmd)
- [Request lifecycle](diagrams/request-lifecycle.mmd)
- [Authentication flow](diagrams/authentication-flow.mmd)
- [Authorization flow](diagrams/authorization-flow.mmd)
- [Database ERD](diagrams/database-erd.mmd)
- [Migration flow](diagrams/migration-flow.mmd)
- [Deployment](diagrams/deployment.mmd)
- [Error handling](diagrams/error-handling.mmd)
- [Feed query](diagrams/feed-query.mmd)

## Documentation status conventions

| Status | Meaning |
|---|---|
| **Complete** | All implementation evidence for the area was reviewed and documented; known gaps are listed explicitly. |
| **Partial** | The area is documented, but some behavior could not be verified, or code and contract disagree. |
| **Not Implemented** | The feature does not exist in the repository. |
| **Not Verified** | The code or configuration exists, but was not exercised during this review. |
| **Blocked** | Verification could not proceed because a tool, credential, or environment is unavailable. |

## Source-of-truth files

When this documentation and the code disagree, the code wins. The most important files are:

- Entry point: `cmd/api/main.go`
- Router: `cmd/api/router.go`
- Middleware: `cmd/api/middleware.go`
- Handlers: `cmd/api/*.go`
- Auth: `internal/auth/*.go`
- Storage and SQL: `internal/storage/*.go`
- Database bootstrap: `internal/db/*.go`
- Migrations: `cmd/migrate/migrations/*.sql`
- Seed: `cmd/seed/main.go`, `internal/db/seed.go`
- Configuration: `internal/env/env.go`, `.example.env`, `docker-compose.yaml`, `makefile`
- API contract: handler annotations and `docs/`
