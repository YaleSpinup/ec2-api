# ec2-api

This API provides simple restful API access to a service.

## Endpoints

```
GET /v1/ec2/ping
GET /v1/ec2/version
GET /v1/ec2/metrics
```

## Authentication

Authentication is accomplished via an encrypted pre-shared key passed via the `X-Auth-Token` header.

## License

GNU Affero General Public License v3.0 (GNU AGPLv3)  
Copyright © 2021 Yale University
