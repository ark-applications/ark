use std::sync::Arc;

use axum::{extract::State, http::StatusCode, response::IntoResponse, Json};
use native_db::Database;
use serde_json::{json, Value};

use crate::models;

#[derive(thiserror::Error, Debug)]
pub enum AdminError {}

pub async fn get_deployment_records(State(db): State<Arc<Database>>) -> impl IntoResponse {
    Json(json!({ "msg": "getting deployment records" }))
}

pub async fn create_deployment() -> Result<impl IntoResponse, (StatusCode, Json<serde_json::Value>)> {
    let json_response = json!({ "msg": "creating deployment" });
    Ok((StatusCode::CREATED, Json(json_response)))
}

pub async fn delete_deployment() -> impl IntoResponse {
    Json(json!({ "msg": "deleting deployment" }))
}

pub async fn upsert_record() -> impl IntoResponse {
    Json(json!({ "msg": "upsert record" }))
}

pub async fn delete_record_by_address() -> impl IntoResponse {
    Json(json!({ "msg": "deleting record by address" }))
}
