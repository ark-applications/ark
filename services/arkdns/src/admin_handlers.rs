use std::sync::Arc;

use axum::{extract::{Path, State}, http::StatusCode, response::IntoResponse, Json};
use native_db::Database;
use serde::Deserialize;
use serde_json::json;

use crate::models;

#[derive(thiserror::Error, Debug)]
pub enum AdminError {}

pub async fn get_deployment_records(State(db): State<Arc<Database<'_>>>, Path((stack_name, deployment_name)): Path<(String, String)>) -> Result<impl IntoResponse, Json<serde_json::Value>> {
    let r = db.r_transaction().expect("failed to create tx");

    let records: Vec<models::Record> = r
        .scan()
        .secondary(models::RecordKey::stack_id).expect("key error").range(stack_name..)
        .collect();

    let response_value: Vec<serde_json::Value> = records
        .into_iter()
        .map(|row| {
            json!(row)
        })
        .collect();
    Ok(Json(json!({ "records": response_value })))
}

pub async fn delete_deployment() -> impl IntoResponse {
    Json(json!({ "msg": "deleting deployment" }))
}

#[derive(Deserialize)]
pub struct RecordUpsertParams {
    pub deployment_name: String,
    pub stack_id: String,
    pub app_name: String,
    pub value: String,
}

pub async fn upsert_record(State(db): State<Arc<Database<'_>>>, Json(record_params): Json<RecordUpsertParams>) -> Result<impl IntoResponse, (StatusCode, Json<serde_json::Value>)> {
    let rw = db.rw_transaction().unwrap();

    rw.insert(models::Record { 
        guid: 1, 
        stack_id: record_params.stack_id, 
        deployment_name: record_params.deployment_name,
        app_name: record_params.app_name,
        record_type: "A".to_string(),
        domain_name: "".to_string(),
        value: record_params.value,
    }).expect("could not insert record");

    rw.commit().expect("failed to commit record upsert");
    let r = db.r_transaction().unwrap();

    let record: models::Record = r.get().primary(1_u32).expect("could not find record").unwrap();

    Ok((StatusCode::OK, Json(record)))
}

pub async fn delete_record_by_address() -> impl IntoResponse {
    Json(json!({ "msg": "deleting record by address" }))
}
