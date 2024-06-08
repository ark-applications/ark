use native_db::*;
use native_model::{native_model, Model};
use serde::{Deserialize, Serialize};

#[derive(Clone, Debug, Serialize, Deserialize)]
#[native_model(id = 1, version = 1)]
#[native_db]
pub struct Record {
    /// A unique identifier for a DNS record
    #[primary_key]
    pub guid: u32,
    /// The stack this record belongs to
    #[secondary_key]
    pub stack_id: String,
    /// The deployment this record belongs to
    pub deployment_name: String,
    /// The app this record belongs to
    pub app_name: String,
    /// The type of DNS record this is (currently only A records supported)
    pub record_type: String,
    /// The domain name this record resolves
    pub domain_name: String,
    /// The value of the record, in practice the IP address of the A record
    pub value: String,
}
