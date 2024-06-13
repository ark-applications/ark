use std::{
    net::Ipv4Addr, str::FromStr, sync::{atomic::{AtomicU64, Ordering}, Arc}
};

use r2d2::{Pool, PooledConnection};
use r2d2_sqlite::SqliteConnectionManager;
use trust_dns_server::{
    authority::MessageResponseBuilder, proto::{op::{Header, MessageType, OpCode, ResponseCode}, rr::{rdata::{self, TXT}, LowerName, Name, RData, Record}}, server::{Request, RequestHandler, ResponseHandler, ResponseInfo}
};

use crate::options::Options;

#[derive(thiserror::Error, Debug)]
pub enum Error {
    #[error("Invalid OpCode {0:}")]
    InvalidOpCode(OpCode),
    #[error("Invalid MessageType {0:}")]
    InvalidMessageType(MessageType),
    #[error("Invalid Zone {0:}")]
    InvalidZone(LowerName),
    #[error("Invalid Record Type {0:}")]
    InvalidRecordType(String),
    #[error("IO error: {0:}")]
    Io(#[from] std::io::Error),
}

/// DNS Request Handler
#[derive(Clone, Debug)]
pub struct DNSHandler {
    db: Pool<SqliteConnectionManager>,
    pub root_zone: LowerName,
    pub counter: Arc<AtomicU64>,
    pub counter_zone: LowerName,
    pub myip_zone: LowerName,
    pub hello_zone: LowerName,
}

fn get_matching_record(db_conn: &PooledConnection<SqliteConnectionManager>, match_on: String, query_name: LowerName) -> Result<(String, String), Error> {
    let mut stmt = db_conn.prepare("SELECT record_type, value FROM dns_records WHERE domain_name LIKE ?;").unwrap();
    let db_record = match stmt.query_row([match_on], |r| {
        Ok((r.get::<usize, String>(0).unwrap(), r.get::<usize, String>(1).unwrap()))
    }) {
        Ok(db_record) => db_record,
        Err(error) => match error {
            rusqlite::Error::QueryReturnedNoRows => {
                return Err(Error::InvalidZone(query_name))
            }
            error => panic!("could not handle sql error on lookup: {:?}", error)
        }
    };
    Ok(db_record)
}

impl DNSHandler {
    /// Create a new handler from the command-line options.
    pub fn from_options(options: &Options, db: &Pool<SqliteConnectionManager>) -> Self {
        let domain = &options.domain;
        DNSHandler {
            db: db.clone(),
            root_zone: LowerName::from(Name::from_str(domain).unwrap()),
            counter: Arc::new(AtomicU64::new(0)),
            counter_zone: LowerName::from(Name::from_str(&format!("counter.{domain}")).unwrap()),
            myip_zone: LowerName::from(Name::from_str(&format!("myip.{domain}")).unwrap()),
            hello_zone: LowerName::from(Name::from_str(&format!("hello.{domain}")).unwrap()),
        }
    }

    /// Handle requests for counter.{domain}
    async fn do_handle_request_count<R: ResponseHandler>(
        &self,
        request: &Request,
        mut responder: R,
    ) -> Result<ResponseInfo, Error> {
        let counter = self.counter.fetch_add(1, Ordering::SeqCst);

        let builder = MessageResponseBuilder::from_message_request(request);
        let mut header = Header::response_from_request(request.header());
        header.set_authoritative(true);

        let rdata = RData::TXT(TXT::new(vec![counter.to_string()]));
        let records = vec![Record::from_rdata(request.query().name().into(), 60, rdata)];

        let response = builder.build(header, records.iter(), &[], &[], &[]);
        Ok(responder.send_response(response).await?)
    }

    async fn do_handle_deployment_request<R: ResponseHandler>(
        &self,
        request: &Request,
        mut responder: R,
    ) -> Result<ResponseInfo, Error> {
        let qname = request.query().name();

        let matching_str = str::replace(&qname.to_string(), &format!(".{}", self.root_zone.to_string()), "");
        let db_conn = self.db.get().unwrap();
        let db_record = match get_matching_record(&db_conn, matching_str, qname.clone()) {
            Ok(db_record) => db_record,
            Err(err) => return Err(err),
        };

        let builder = MessageResponseBuilder::from_message_request(request);
        let mut header = Header::response_from_request(request.header());
        header.set_authoritative(true);

        let rdata = match db_record.0 {
            record_type if record_type == "A" => {
                let ip_addr_octets = db_record.1.parse::<Ipv4Addr>().unwrap().octets();
                RData::A(rdata::A::new(ip_addr_octets[0], ip_addr_octets[1], ip_addr_octets[2], ip_addr_octets[3]))
            }
            record_type => {
                return Err(Error::InvalidRecordType(record_type))
            }
        };

        let records = vec![Record::from_rdata(qname.into(), 60, rdata)];

        let response = builder.build(header, records.iter(), &[], &[], &[]);
        Ok(responder.send_response(response).await?)
    }

    async fn do_handle_request<R: ResponseHandler>(
        &self,
        request: &Request,
        response: R,
    ) -> Result<ResponseInfo, Error> {
        // make sure the request is a query
        if request.op_code() != OpCode::Query {
            return Err(Error::InvalidOpCode(request.op_code()));
        }

        // make sure the message type is a query
        if request.message_type() != MessageType::Query {
            return Err(Error::InvalidMessageType(request.message_type()));
        }

        match request.query().name() {
            name if self.counter_zone.zone_of(name) => {
                self.do_handle_request_count(request, response).await
            }
            name if self.root_zone.zone_of(name) => {
                self.do_handle_deployment_request(request, response).await
            }
            name => Err(Error::InvalidZone(name.clone())),
        }
    }
}

#[async_trait::async_trait]
impl RequestHandler for DNSHandler {
    async fn handle_request<R: ResponseHandler>(
        &self,
        request: &Request,
        response: R,
    ) -> ResponseInfo {
        match self.do_handle_request(request, response).await {
            Ok(info) => info,
            Err(error) => {
                print!("Error in RequestHandler: {error}");
                let mut header = Header::new();
                header.set_response_code(ResponseCode::ServFail);
                header.into()
            }
        }
    }
}
