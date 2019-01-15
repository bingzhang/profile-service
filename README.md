# Profile Demo Server

## Installation on https://profile.inabyte.com

1. Install GO: https://golang.org/doc/install

2. Clone this repo (outside GOPATH `~/go/src`)
```
$ cd <your path> 
$ git clone git@myles.inabyte.com:profile.git
Cloning into 'profile'...
remote: Counting objects: 16, done.
remote: Compressing objects: 100% (15/15), done.
remote: Total 16 (delta 5), reused 0 (delta 0)
Receiving objects: 100% (16/16), 5.91 KiB | 2.95 MiB/s, done.
Resolving deltas: 100% (5/5), done.
```

3. Make the project
```
$ cd profile
$ make
▶ running gofmt…
▶ setting GOPATH…
▶ setting DEPPATH…
▶ building github.com/golang/dep/cmd/dep…
▶ building github.com/mjibson/esc…
▶ retrieving dependencies…
▶ building golang.org/x/lint/golint…
▶ running go-lint…
▶ building executable(s)… 0.0.0 2019-01-15T09:44:32+0200
# profile
ld: warning: text-based stub file /System/Library/Frameworks//CoreFoundation.framework/CoreFoundation.tbd and library file /System/Library/Frameworks//CoreFoundation.framework/CoreFoundation are out of sync. Falling back to library file for linking.
ld: warning: text-based stub file /System/Library/Frameworks//Security.framework/Security.tbd and library file /System/Library/Frameworks//Security.framework/Security are out of sync. Falling back to library file for linking.
```

4. Start the service
```
$ ./bin/profile
```

## Local Installation

1. Install GO: https://golang.org/doc/install

2. Clone this repo in `~/go/src`

3. Install SQLite3 package
```
$ go get github.com/mattn/go-sqlite3
```

4. Build the project
```
$ cd ~/go/src/profile_demo
$ go build
```

5. Start the service
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
