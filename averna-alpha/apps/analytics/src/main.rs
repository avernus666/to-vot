use axum::{routing::get, Json, Router};
use serde::Serialize;
use std::{env, net::SocketAddr};

#[derive(Serialize)]
struct HealthResponse {
    status: &'static str,
    service: &'static str,
}

async fn health() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "ok",
        service: "averna-analytics",
    })
}

#[tokio::main]
async fn main() {
    let address: SocketAddr = env::var("ANALYTICS_ADDR")
        .unwrap_or_else(|_| "127.0.0.1:8090".to_owned())
        .parse()
        .expect("ANALYTICS_ADDR must be a valid socket address");

    let app = Router::new().route("/health", get(health));
    let listener = tokio::net::TcpListener::bind(address)
        .await
        .expect("failed to bind analytics listener");

    println!("analytics service listening on {address}");
    axum::serve(listener, app)
        .await
        .expect("analytics service failed");
}
