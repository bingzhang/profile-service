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

4. Ensure `/private/var/lib/profile` directory and its access
```
$ sudo mkdir -p "/var/lib/profile"
$ sudo chmod 777 "/var/lib/profile"
```

5. Start the service
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

5. Ensure `/private/var/lib/profile` directory and its access
```
$ sudo mkdir -p "/var/lib/profile"
$ sudo chmod 777 "/var/lib/profile"
```

6. Start the service
```
$ ./profile_demo
```

## profile API

### __GET__ profile API
Retrieves user profile.

__Request:__
```
GET https://profile.inabyte.com/profile?uuid=<user_uuid>
```

__Response:__
  - 200
```
{
  "uuid": <user_uuid>,
  "name": <name>,
  "phone": <phone>,
  "birth_date": <birth_date>,
  "role": "student" | "staff" | "other"
}
```

  - 404 "User Not found."

  - 500 <error_description>

__Example:__
```
curl -X GET https://profile.inabyte.com/profile?uuid=e92e429f-84b9-4dcc-bf90-f969137d2402
```

### __POST__ profile API
Adds/updates user profile.

__Request:__
```
POST https://profile.inabyte.com/profile
{
  "uuid": <user_uuid>,
  "name": <name>,
  "phone": <phone>,
  "birth_date": <birth_date>,
  "role": "student" | "staff" | "other"
}
```

__Response:__
  - 200 OK
  - 500 <error_description>

__Example:__
```
curl -X POST -d '{"uuid":"e92e429f-84b9-4dcc-bf90-f969137d2402", "name":"John Paul", "phone":"+1 650-207-7211", "birth_date":"1956/03/30", "role":"other"}' -H "Content-Type: application/json" https://profile.inabyte.com/profile
```

### __DELETE__ profile API
Deletes user profile.

__Request:__
```
DELETE https://profile.inabyte.com/profile?uuid=<user_uuid>
```

__Response:__
  - 200 OK
  - 500 <error_description>

__Example:__
```
curl -X DELETE https://profile.inabyte.com/profile?uuid=e92e429f-84b9-4dcc-bf90-f969137d2402
```

## ui/config API

All API entries take "lang" optional parameter. If it is omitted "en" is assumed. The initial predefined configs are "en", "es" and "zh"

### __GET__ ui/config API
Retrieves current UI config.

__Request:__
```
GET https://profile.inabyte.com/ui/config?lang=en
```

__Response:__
  - 200
```
<config data>
```

  - 404 "UI Config not set."

  - 500 <error_description>

__Example:__
```
curl -X GET https://profile.inabyte.com/ui/config?lang=en
```

### __POST__ ui/config API
Updates current UI config.

__Request:__
```
POST https://profile.inabyte.com/ui/config?lang=en
<config data>
```

__Response:__
  - 200 OK
  - 404 <error_description>
  - 500 "Failed to store ui config: ..."

__Example:__
```
curl -X POST -d "@uiconfig.json" -H "Content-Type: application/json" https://profile.inabyte.com/ui/config?lang=en
```

### __DELETE__ ui/config API
Resets current UI config to initial versions.

__Request:__
```
DELETE https://profile.inabyte.com/ui/config?lang=en
```

__Response:__
  - 200 OK
  - 404 "Unable to perform operation"
  - 500 <error_description>

__Example:__
```
curl -X DELETE https://profile.inabyte.com/ui/config?lang=en
```

## events API

### __GET__ events API
Retrieves events.

__Paramters:__
  - "time" (YYYY-MM-DD HH:MM:SS) Retrieves events that start in date of time and that are not finished yet.
  - "role" (student|staff|other) Retrieves events that has specific user role.

__Request:__
```
GET https://profile.inabyte.com/events?time=2019-03-05%2015:00:00&role=student
```

__Response:__
  - 200
```
[
{
  "id": <event id, integer>,
  "name": <name>,
  "time": <YYYY-MM-DD HH:MM:SS>,
  "duration": <duration in minutes>,

  "location_description": <location description>,
  "location_latitude": <location latitude>,
  "location_longtitude": <location longtitude>,
  "location_floor": <location floor>,

  "purchase_description": <purchase description>,
  "info_url": <info url>,

  "category": <category>,
  "sub_category": <sub category>,

  "user_role": "student" | "staff" | "other"
}
]
```

  - 500 <error_description>

__Example:__
```
curl -X GET https://profile.inabyte.com/events?time=2019-03-05%2015:00:00&role=student
```

### __POST__ events API
Adds new events.

__Request:__
```
POST https://profile.inabyte.com/events
[
  {
    "name": <name>,
    "time": <YYYY-MM-DD HH:MM:SS>,
    "duration": <duration in minutes>,

    "location_description": <location description>,
    "location_latitude": <location latitude>,
    "location_longtitude": <location longtitude>,
    "location_floor": <location floor>,

    "purchase_description": <purchase description>,
    "info_url": <info url>,

    "category": <category>,
    "sub_category": <sub category>,

    "user_role": "student" | "staff" | "other"
  },
  ...
]
```

__Response:__
  - 200 OK
  - 500 <error_description>

__Example:__
```
curl -X POST -d '[{"name":"Students Conference", "time":"2019-03-05 13:00:00", "duration":240, "location":{"description":"B216, RTX", "latitude":57.0861893, "longtitude":9.9578803, "floor":1}, "purchase_description":"Please Buy", "info_url":"http://www.inabyte.com", "category":"lecture", "sub_category":"physics", "user_role":"student"}]' -H "Content-Type: application/json" https://profile.inabyte.com/events
```

### __DELETE__ events API
Deletes events. If not parameter is applied all records from events table are deleted.

__Paramters:__
  - "id" (#,#,#) Comma separated list of identifiers to delete.

__Request:__
```
DELETE https://profile.inabyte.com/events?id=#,#,#
```

__Response:__
  - 200 OK
  - 500 <error_description>

__Example:__
```
curl -X DELETE https://profile.inabyte.com/events?id=2,3
```
