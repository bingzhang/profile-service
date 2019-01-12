# Profile Demo Server

## Instalation

1. Install GO: https://golang.org/doc/install

2. Clone this repo in `~/go/src`

3. Build the project
```
$ cd ~/go/src/profile_demo
$ go build
```

4. Start the service
```
$ ./profile_demo
```

## Usage

### __GET__ profile API

__Request:__
```
GET http://localhost:8082/profile?uuid=<user_uuid>
```

__Response:__
  - 200
```
{
  "uuid":<user_uuid>,
  "name":<name>,
  "phone":<phone>,
  "birth_date":<birth_date>,
}
```

  - 404 "User Not found."

  - 500 <error_description>

__Example:__
```
curl http://localhost:8082/profile?uuid=e92e429f-84b9-4dcc-bf90-f969137d2402
```

### __POST__ profile API

__Request:__
```
POST http://localhost:8082/profile
{
  "uuid":<user_uuid>,
  "name":<name>,
  "phone":<phone>,
  "birth_date":<birth_date>,
}
```

__Response:__
  - 200 OK
  - 500 <error_description>

__Example:__
```
curl -X POST -d '{"uuid":"e92e429f-84b9-4dcc-bf90-f969137d2402", "name":"John Paul", "phone":"+1 650-207-7211", "birth_date":"1956/03/30"}' -H "Content-Type: application/json" http://localhost:8082/profile
```

### __DELETE__ profile API

__Request:__
```
DELETE http://localhost:8082/profile?uuid=<user_uuid>
```

__Response:__
  - 200 OK
  - 500 <error_description>

__Example:__
```
curl -X DELETE http://localhost:8082/profile?uuid=e92e429f-84b9-4dcc-bf90-f969137d2402
```
