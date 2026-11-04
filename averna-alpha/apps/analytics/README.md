# AVERNA Analytics

Rust service boundary for analytics workloads. Current scope is a health endpoint only.
The next implementation adds a Redis Streams consumer, versioned event validation,
deduplication by event ID, retry/dead-letter handling, and PostgreSQL-backed aggregates.
Core business rules remain in the Go application.
