use std::time::Duration;

use anyhow::Result;
use axum::{routing::{delete, get, put}, Router};
use clap::Parser;
use dns_handler::DNSHandler;
use options::Options;
use r2d2_sqlite::SqliteConnectionManager;
use tokio::net::{TcpListener, UdpSocket};
use trust_dns_server::ServerFuture;

mod admin_handlers;
mod dns_handler;
mod models;
mod options;

/// Timeout for TCP connections.
const TCP_TIMEOUT: Duration = Duration::from_secs(10);

#[tokio::main]
async fn main() -> Result<()> {
    tracing_subscriber::fmt::init();
    let options = Options::parse();

    let db_manager = SqliteConnectionManager::memory();
    let db_pool = r2d2::Pool::new(db_manager).unwrap();
    initialize_db(&db_pool).await?;

    let dns_server = build_dns_server(&options, &db_pool).await?;
    let dns_task = tokio::spawn(async move {
        dns_server.block_until_done().await
    });

    // build admin server
    let admin_api = Router::new()
        .route("/v1/up", get(|| async { "Hello from arkdns" }))
        .route("/v1/stacks/:stack_id/deployments/:deployment_name", get(admin_handlers::get_deployment_records))
        .route("/v1/stacks/:stack_id/deployments/:deployment_name", delete(admin_handlers::delete_deployment))
        .route("/v1/stacks/:stack_id/deployments/:deployment_name/record", put(admin_handlers::upsert_record))
        .with_state(db_pool);

    // run admin server
    let admin_task = tokio::spawn(async move {
        let listener = tokio::net::TcpListener::bind(options.admin_addr).await.unwrap();
        axum::serve(listener, admin_api).await
    });

    let _ = dns_task.await?;
    let _ = admin_task.await?;
    Ok(())
}

async fn initialize_db(db: &r2d2::Pool<SqliteConnectionManager>) -> Result<()> {
    let table_migration = r#"
        CREATE TABLE IF NOT EXISTS dns_records (
            guid INTEGER PRIMARY KEY,
            stack_id TEXT,
            deployment_name TEXT,
            app_name TEXT,
            record_type TEXT,
            domain_name TEXT,
            value TEXT
        );
    "#;

    db.get()
        .unwrap()
        .execute(table_migration, ())
        .unwrap();

    Ok(())
}

async fn build_dns_server(options: &Options, db: &r2d2::Pool<SqliteConnectionManager>) -> Result<ServerFuture<DNSHandler>> {
    let dns_handler = DNSHandler::from_options(&options, db);

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
