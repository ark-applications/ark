# arkdns

A custom nameserver to direct traffic to deployments.

Serves records of the following format:

{app_name}.{deployment_name}.{stack_root_domain}

Each stack gets its own root domain.

The `app_name` is optional, the stack will define a `root_app` that {deployment_name}.{stack_root_domain} should go to.

TODO: Add support for CNAME records. internal queries can use a CNAME to get a container name from app name

## Admin API

TODO: Update this

**GET** `/v1/stacks/{stack_id}/deployments/{deployment_name}` - Returns DNS records for the deployment
**POST** `/v1/stacks/{stack_id}/deployments/{deployment_name}` - Creates a new deployment
**PUT** `/v1/stacks/{stack_id}/deployments/{deployment_name}/record` - Upserts a record
**DELETE** `/v1/stacks/{stack_id}/deployments/{deployment_name}/record/{address}` - Deletes all records with the specified address as their value
**DELETE** `/v1/stacks/{stack_id}/deployments/{deployment_name}` - Deletes all records for a deployment

## References

- https://github.com/EmilHernvall/dnsguide
- https://dev.to/xfbs/writing-a-dns-server-in-rust-1gpn
