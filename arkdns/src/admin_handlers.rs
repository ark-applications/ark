
use axum::{extract::{Path, State}, http::StatusCode, response::IntoResponse, Json};
use r2d2::Pool;
use r2d2_sqlite::SqliteConnectionManager;
use serde::Deserialize;
use serde_json::json;

use crate::models;

#[derive(thiserror::Error, Debug)]
pub enum AdminError {}

pub async fn get_deployment_records(State(db): State<Pool<SqliteConnectionManager>>, Path((stack_id, deployment_name)): Path<(String, String)>) -> Result<impl IntoResponse, Json<serde_json::Value>> {
    let db_conn = db.get().unwrap();

    let mut stmt = db_conn.prepare("SELECT * FROM dns_records WHERE stack_id = ? AND deployment_name = ?;").unwrap();
    let records_iter = stmt.query_map([stack_id, deployment_name], |r| {
        Ok(models::Record{
            guid: r.get(0)?,
            stack_id: r.get(1)?,
            deployment_name: r.get(2)?,
            app_name: r.get(3)?,
            record_type: r.get(4)?,
            domain_name: r.get(5)?,
            value: r.get(6)?,
        })
    }).unwrap();

    let mut matching_records: Vec<models::Record> = Vec::new();
    for r in records_iter {
        matching_records.push(r.unwrap());
    }
    Ok(Json(json!({ "records": matching_records })))
}

pub async fn delete_deployment(State(db): State<Pool<SqliteConnectionManager>>, Path((stack_id, deployment_name)): Path<(String, String)>) -> impl IntoResponse {
    let db_conn = db.get().unwrap();

    db_conn.execute(
        "DELETE FROM dns_records WHERE stack_id = ? AND deployment_name = ?", 
        [stack_id, deployment_name]
    ).expect("could not delete deployment");

    Json(json!({ "msg": "deployment deleted" }))
}

#[derive(Deserialize)]
pub struct RecordUpsertParams {
    pub app_name: String,
    pub value: String,
}

pub async fn upsert_record(State(db): State<Pool<SqliteConnectionManager>>, Path((stack_id, deployment_name)): Path<(String, String)>, Json(record_params): Json<RecordUpsertParams>) -> Result<impl IntoResponse, (StatusCode, Json<serde_json::Value>)> {
    let db_conn = db.get().unwrap();

    let domain_name = format!("{app_name}.{deployment_name}.{stack_id}.", app_name = record_params.app_name);

    let created_record = db_conn.query_row(r#"
            INSERT INTO dns_records (
                stack_id,
                deployment_name,
                app_name,
                record_type,
                domain_name,
                value
            ) VALUES (?, ?, ?, ?, ?, ?) RETURNING guid, stack_id, deployment_name, app_name, record_type, domain_name, value;
        "#, 
        &[
            &stack_id, 
            &deployment_name, 
            &record_params.app_name, 
            &"A".to_string(),
            &domain_name.to_string(),
            &record_params.value,
        ], |r| {
            Ok(models::Record {
                guid: r.get(0)?,
                stack_id: r.get(1)?,
                deployment_name: r.get(2)?,
                app_name: r.get(3)?,
                record_type: r.get(4)?,
                domain_name: r.get(5)?,
                value: r.get(6)?,
            })
        }).unwrap();

    Ok((StatusCode::OK, Json(created_record)))
}
