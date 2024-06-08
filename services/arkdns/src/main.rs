use std::{sync::Arc, time::Duration};

use anyhow::Result;
use axum::{routing::{delete, get, put}, Router};
use clap::Parser;
use dns_handler::DNSHandler;
use native_db::Database;
use once_cell::sync::Lazy;
use options::Options;
use tokio::net::{TcpListener, UdpSocket};
use trust_dns_server::ServerFuture;

mod admin_handlers;
mod dns_handler;
mod models;
mod options;

/// Timeout for TCP connections.
const TCP_TIMEOUT: Duration = Duration::from_secs(10);

static ADMIN_DATABASE_BUILDER: Lazy<native_db::DatabaseBuilder> = Lazy::new(|| {
    let mut builder = native_db::DatabaseBuilder::new();
    builder
        .define::<models::Record>()
        .expect("failed to define model Record");
    builder
});

#[tokio::main]
async fn main() -> Result<()> {
    tracing_subscriber::fmt::init();
    let options = Options::parse();

    let db = ADMIN_DATABASE_BUILDER
        // Create with a file path to persist the database
        .create_in_memory()
        .expect("failed to create database");

    let rw = db
        .rw_transaction()
        .expect("failed to create rw migration transaction");
    rw.migrate::<models::Record>().expect("failed to migrate Record");
    rw.commit().expect("failed to commit migration");

    let shared_db = Arc::new(db);

    let dns_server = build_dns_server(&options, &shared_db).await?;
    let dns_task = tokio::spawn(async move {
        dns_server.block_until_done().await
    });

    // build admin server
    let admin_api = Router::new()
        .route("/v1/up", get(|| async { "Hello from arkdns" }))
        .route("/v1/stacks/:stack_id/deployments/:deployment_name", get(admin_handlers::get_deployment_records))
        .route("/v1/stacks/:stack_id/deployments/:deployment_name", delete(admin_handlers::delete_deployment))
        .route("/v1/stacks/:stack_id/deployments/:deployment_name/record", put(admin_handlers::upsert_record))
        .route("/v1/stacks/:stack_id/deployments/:deployment_name/record/:address", delete(admin_handlers::delete_record_by_address))
        .with_state(shared_db);

    // run admin server
    let admin_task = tokio::spawn(async move {
        let listener = tokio::net::TcpListener::bind(options.admin_addr).await.unwrap();
        axum::serve(listener, admin_api).await
    });

    let _ = dns_task.await?;
    let _ = admin_task.await?;
    Ok(())
}

async fn build_dns_server(options: &Options, _db: &Arc<Database<'static>>) -> Result<ServerFuture<DNSHandler>> {
    let dns_handler = DNSHandler::from_options(&options);

    // create DNS server
    let mut server = ServerFuture::new(dns_handler);

    // register UDP listeners
    for udp in &options.udp {
        server.register_socket(UdpSocket::bind(udp).await?)
    }

    // register TCP listeners
    for tcp in &options.tcp {
        server.register_listener(TcpListener::bind(&tcp).await?, TCP_TIMEOUT);
    }

    Ok(server)
}
