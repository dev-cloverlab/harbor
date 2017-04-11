# harbor

harbor is API server manages the docker containers

```
% harbor -p :9999 -r "10000:12000" -w ./harbor
```

options

```
  -p string
    	server listen port (default ":8080")
  -r string
    	port range for containers (default "10000:12000")
  -w string
    	application workspace path (default "./harbor")
```

# API

## Register Project

```
% curl -XPOST -d 'payload={"name":"${PROJECT_NAME}", "repo":"${REPO_NAME}' localhost:9999/register
```

response

```
{
    "name": "project name",
    "repo": "repositry name",
    "branch": [
        "deployed branch informations ([]branch.Config)"
    ],
    "created_at": "create time"
}
```

## Unregister project

```
% curl -XPOST -d 'payload={"name":"${PROJECT_NAME}"}' localhost:9999/unregister
```

response

```
{
    "portsAvailable": "available port range",
    "portsAllocated": [
        "allocated port list ([]int)"
    ],
}
```

## Get project list

```
curl -XGET localhost:9999/list
```

response

```
{
    "projects: [
        "project list ([]project.Config)
    ]
}
```

## Get deploy branch status

```
% curl -XGET localhost:9999/br?name=${PROJECT_NAME}&branch=${BRANCH_NAME}
```

response 

```
{
    "name": "branch name",
    "port": [
        "allocated port list ([]int)
    ],
    "state": "deploy state (0:unknown 1:started 2:done)",
    "work": "deploy path",
    "notice": "deploy message",
    "deployed_at": "deploy time"
}
```

## docker-compose up

```
% curl -XPOST -d 'payload={"name":"${PROJECT_NAME}", "branch":"${BRANCH_NAME}' localhost:9999/up
```

response 

```
# same as /br 
```

## docker-compose down

```
% curl -XPOST -d 'payload={"name":"${PROJECT_NAME}", "branch":"${BRANCH_NAME}' localhost:9999/down
```

response

```
# same as /register
```

# Examples

execute server

```
% harbor -p :9999 -r "10000:12000" -w /tmp/harbor
```

registe project to harbor

```
% curl -XPOST -d 'payload={"name":"${PROJECT_NAME}", "repo":"${REPO_NAME}' localhost:9999/register
```

create project like below 

```
% tree                                                                                                                                             hatajoe/umedago (hoge) hatajoe
.
├── docker-compose.yml
└── public
    └── index.html

1 directory, 3 files
```

docker-compose.yml

```
nginx:
  image: nginx
  ports:
   - "${PORT1}:80"
  volumes: 
    - ./public/:/usr/share/nginx/html/
```

index.html

```
Hello, World!
```

create repository and commit

```
% git init
% git add .
% git commit -m "Initial commit"
% hub create
```

add .git/hooks/pre-push and add executable permission

```
#!/bin/sh

PROJECT_NAME="replace here"

HARBOR_DOMAIN=localhost:9999
BRANCH=$(git rev-parse --abbrev-ref HEAD)

RES=`curl -XGET http://$HARBOR_DOMAIN/br?name=$PROJECT_NAME\&branch=$BRANCH`
WORK=`echo ${RES} | jq -r '.work'`

# deploy

mkdir -p ${WORK}
rsync -rv --exclude=.git . ${WORK}

# docker-compose up
FLG=1
curl -XPOST -d "payload={\"name\":\"$PROJECT_NAME\", \"branch\":\"$BRANCH\"}" http://$HARBOR_DOMAIN/up

while [ $FLG == 1 ]
do
	sleep 1
    exit;
	RES=`curl -XGET http://$HARBOR_DOMAIN/br?name=$PROJECT_NAME\&branch=$BRANCH`
	if [ $? == 0 ]; then
		STATE=`echo $RES | jq -r '.state'`
		if [ "$STATE" == "2" ]; then
			FLG=2
		fi
	fi
done

exit 0
```

deploy

```
% git push -u origin master
```

maybe this works...

```
% open http://localhost:10000
```

