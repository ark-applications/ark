use clap::Parser;
use std::net::SocketAddr;

#[derive(Parser,Clone,Debug)]
pub struct Options {
    /// UDP socket to listen on.
    #[clap(long, short, default_value = "0.0.0.0:1053", env = "ARKDNS_UDP")]
    pub udp: Vec<SocketAddr>,

    /// TCP socket to listen on.
    #[clap(long, short, env = "ARKDNS_TCP")]
    pub tcp: Vec<SocketAddr>,

    /// Socket address the admin API should listen on.
    #[clap(long, short, default_value = "0.0.0.0:4500", env = "ARKDNS_ADMIN_API_ADDR")]
    pub admin_addr: SocketAddr,

    /// Domain name
    #[clap(long, short, default_value = "internal", env = "ARKDNS_DOMAIN")]
    pub domain: String,
}
