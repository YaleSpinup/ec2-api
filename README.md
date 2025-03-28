# ec2-api

This API is a reimplementation of the internal `ec2-api` and provides simple restful API access to EC2.

## API Endpoints

```
# Base Path
GET /v2/ec2/

# General Endpoints
GET /v2/ec2/ping
GET /v2/ec2/version
GET /v2/ec2/metrics

# Managing Instances
GET /v2/ec2/{account}/instances
GET /v2/ec2/{account}/instances/types
GET /v2/ec2/{account}/instances/{id}
GET /v2/ec2/{account}/instances/{id}/volumes
GET /v2/ec2/{account}/instances/{id}/volumes/{vid}
GET /v2/ec2/{account}/instances/{id}/snapshots
GET /v2/ec2/{account}/instances/{id}/ssm/ready
POST /v2/ec2/{account}/instances
POST /v2/ec2/{account}/instances/{id}/volumes
PUT /v2/ec2/{account}/instances/{id}
PUT /v2/ec2/{account}/instances/{id}/power
PUT /v2/ec2/{account}/instances/{id}/ssm/command
PUT /v2/ec2/{account}/instances/{id}/ssm/association
PUT /v2/ec2/{account}/instances/{id}/tags
PUT /v2/ec2/{account}/instances/{id}/attribute
DELETE /v2/ec2/{account}/instances/{id}
DELETE /v2/ec2/{account}/instances/{id}/volumes/{vid}

# Managing Security Groups (SG)
GET /v2/ec2/{account}/sgs
GET /v2/ec2/{account}/sgs/{id}
POST /v2/ec2/{account}/sgs
PUT /v2/ec2/{account}/sgs/{id}
PUT /v2/ec2/{account}/sgs/{id}/tags
DELETE /v2/ec2/{account}/sgs/{id}

# Managing Volumes
GET /v2/ec2/{account}/volumes
GET /v2/ec2/{account}/volumes/{id}
GET /v2/ec2/{account}/volumes/{id}/modifications
GET /v2/ec2/{account}/volumes/{id}/snapshots
POST /v2/ec2/{account}/volumes
PUT /v2/ec2/{account}/volumes/{id}
PUT /v2/ec2/{account}/volumes/{id}/tags
DELETE /v2/ec2/{account}/volumes/{id}

# Managing Snapshots
GET /v2/ec2/{account}/snapshots
GET /v2/ec2/{account}/snapshots/{id}
POST /v2/ec2/{account}/snapshots
PUT /v2/ec2/{account}/snapshots/synctags
DELETE /v2/ec2/{account}/snapshots/{id}

# Managing Images
GET /v2/ec2/{account}/images?name={name}
GET /v2/ec2/{account}/images
GET /v2/ec2/{account}/images/{id}
POST /v2/ec2/{account}/images
PUT /v2/ec2/{account}/images/{id}/tags
DELETE /v2/ec2/{account}/images/{id}

# Managing Subnets
GET /v2/ec2/{account}/subnets?vpc={vpc}
GET /v2/ec2/{account}/subnets

# Managing VPCs
GET /v2/ec2/{account}/vpcs
GET /v2/ec2/{account}/vpcs/{id}

# Miscellaneous Endpoints
DELETE /v2/ec2/{account}/instanceprofiles/{name}
```

## SSM Readiness Check

The SSM readiness check endpoint allows you to verify if an EC2 instance has the Systems Manager agent properly installed, configured, and connected.

```
GET /v2/ec2/{account}/instances/{id}/ssm/ready
```

### Parameters
- `{account}`: AWS account ID or account alias
- `{id}`: EC2 instance ID (e.g., 'i-12345678')

### Response
```json
{
  "instanceId": "i-1234567890abcdef0",
  "ready": true
}
```

### Response Codes
- 200 OK: Request was successful (instance may or may not be ready)
- 400 Bad Request: Invalid request parameters
- 403 Forbidden: Authorization error
- 500 Internal Server Error: Server error while processing the request

## Authentication

Authentication is accomplished via an encrypted pre-shared key passed via the `X-Auth-Token` header.

## Authors

E Camden Fisher <camden.fisher@yale.edu>  
Brandon Tassone <brandon.tassone@yale.edu>

## License

GNU Affero General Public License v3.0 (GNU AGPLv3)  
Copyright © 2021 Yale University
