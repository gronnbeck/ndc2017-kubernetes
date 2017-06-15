# Live Coding Manuscript

## Setup

To setup everything we need to first setup the cluster and add all configurations and secrets.

### Source secrets into env

```sh
$ source env.secret
```

### Switch to correct project

```sh
$ ./scripts/project-auth.sh unacast-talk-kubernetes-scale # change to what you use
```

### Start cluster

```sh
$ ./scripts/start-cluster.sh ndc 50
```

### Create config and secret

```sh
$ ./scripts/secret-privkey.sh
$ ./scripts/secret-redis-simple.sh
$ ./scripts/secret-ssl-tunnel.sh
$ ./scripts/configmap-redis.sh
```

#### Start helloworldserver

```sh
$ ./scripts/helloworld.yaml
```

#### Start vegeta for helloworldserver

```sh
$ ./scripts/helloworld-vegeta.yaml
```

#### Start aggregator

```sh
$ ./scripts/aggregator.yaml
```

#### Setup aggregator port-forward

```sh
$ k port-forward <pod-name> 8089:8082
```

#### Serve performance dashboard

```sh
$ cd ...
```

#### Serve currency

```sh
$ cd ...
```

### Start

## Session 1

Here we go. I've cheated a bit and prepared a framework we can use, or we wouldn't have finished on time. So let's quickly step through what I've setup.

`show terminal`

As you can see I've started a Kubernetes cluster on Google Cloud with 50 nodes. Each node has 16 cores each. In other words, we have about 800 cores to our disposal. This will be fun. Just remind me to shut it down before I leave today.

I've also loaded a few configs and secrets we'll need during the coding session. All of them are related to the Redis database we are going to use in "production". Let's step into the code I've prepared.

`show main.go`

I've setup a simple `main.go` file. I've setup a `http` endpoint and a `handler` for root (`/`) accepting both `GET` and `POST`. I've made the endpoint to timeout if the runtime is more than `400 ms`. It's a bit hacky to do so for an entire endpoint. But it will save us time.

The function we are going to implement are setup of the redis client, `get` , and `post` in first lines of this file.

```go
var client *pool.Pool
getHandleFunc := get(client)
postHandleFunc := post(client)
```

We implement `get` in `impl-get.go` (step to file) and `post` in `impl-post.go` (step to file). And to make things easier I've implemented some helpers for handling Redis in `helpers.go`.

`show helpers.go`

I've implemented three simple methods `hasValue`, `getValue`, and `setValue`. And the way I handle errors in this file is not recommended. It's not a good practice for production grade systems but it works for presenting. Please forgive me for being sloppy.

OK! Let's get started with implementing this random and arbitrary API.

`open main.go`

First, let us setup redis. For this I use a simple internal package I created a long time ago called `setupredis`. It's a thin wrapper for `mediocregopher/radix.v2` I created for a previous project. A side note on thin wrappers are that you should try not to use external thin wrappers. But as helpers I believe they are useful.

```go
client, err := setupredis.NewWait(*redisAddr, *redisPass)
if err != nil {
  panic(err)
}
```

Pretty simple. The `NewWait` function retries a few times if it wasn't able to get a connection to the specified Redis instance. Now that we have setup the redis instance let's implement a simple `get` and `post`. We'll start with `get`.

`open impl-get.go`

Hey! So first let's check if we have a value using `hasValue`. If we don't have a that value. We want to return an error using `NewErrorResponse` informing the user that value is not set yet.

If we have the value we want to return a response containing the value.

```go
func get(client *pool.Pool) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        has := hasValue(client, "current-value")

        if !has {
            resp := domain.NewErrorResponse("Not initialised")
            w.Write(resp.JSON())
            return
        }

        val := getValue(client, "current-value")

        resp := domain.NewResponse(val)

        w.Write(resp.JSON())
    }
}
```

Let's jump into `post` and implement that.

`open impl-post.go`

First, let's get the data from the request and then parse it into a `Request` structure by using the helper function `RequestFromBytes`. And set that value. I'm skipping the validation. Yeah!

```go
func post(client *pool.Pool) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        byt, err := ioutil.ReadAll(r.Body)
        if err != nil {
            panic(err)
        }
        req := domain.RequestFromBytes(byt)
        setValue(client, "current-value", req)
    }
}
```

### Deploying it

Now that the code is in place. Let's deploy it to Kubernetes and performance test it. First, we need to pack the program into a Docker container and push it to a registry.

`open Dockerfile`

The Dockerfile is pretty generic and it's designed to just execute the `./run.sh` program.

`open ./scripts/build.sh`

Here we build `go` with Linux architecture in mind. So we can deploy it directly into a Linux container without having to worry about the run time environment. This is one place where go really shines.

`open ./scripts/push-build.sh`

This is pretty simple. It just pushes the container to the docker hub registry.

#### Building the container

```sh
$ ./scripts/build.sh
```

#### Pushing it to docker registry

```sh
$ ./scripts/push-build.sh
```

Since it's a bit cumbersome to write the deployment manifest I've cheated and prewritten it. Let's step through it together.

The first part describes what object we're describing and how many replicas we want to run. It's always good to run more than one instance of a services. It forces you to be stateless.

```yaml
kind: Deployment
apiVersion: apps/v1beta1
metadata:
  name: redis-api
  labels:
      name: redis-api
      app: redis-api
      tier: api
spec:
  replicas: 3
```

Second, we specify the container we want to want to deploy. In this case it's `redis-api:ndc` registered under my user on DockerHub.

```yaml
- name: redis-api
  imagePullPolicy: Always
  image: gronnbeck/redis-api:ndc
  resources:
    limits:
      cpu: 2
      memory: 1024Mi
    requests:
      cpu: 1
      memory: 500Mi
  env:
  - name: REDIS_ADDR
    value: 127.0.0.1:7007
  - name: REDIS_PASS
    valueFrom:
      secretKeyRef:
        name: redis-api-simple
        key: redis-pass
  ports:
    - name: api
      containerPort: 8080
  readinessProbe:
    httpGet:
      path: /
      port: 8080
    initialDelaySeconds: 5
    timeoutSeconds: 2
```

The last thing is the setup for the SSL tunnel to the Redis instance on compose. Not going to dig into that now.

Let's deploy this baby.

`open terminal`

```sh
$ k apply -f ./manifest/deployment.yaml
$ k apply -f ./manifest/service.yaml
```

Let's check that everything is up and running smoothley

```sh
$ k get po -lapp=redis-api
```

```sh
$ k get svc
```

Everything seems to be running smooth. And now let's check how it scales by pointing vegeta to attack the new services.

```sh
$ k apply -f ./manifest/api-vegeta.yaml
```

and let's add something that updates the API. I have cheated here as well and created a little something that sends POST request to our API.

```sh
$ k apply -f ./manifest/updater.yaml
```

Let's check that it actually works. Jupp! It works smoothly.

Now let's check if it scales with x: {4000,6000,7000,8000,9000} requests per second.

```sh
$ k scale deployment vegetaserver --replicas=x/1000
```

`open performance dashboard`

We clearly hit a bottleneck here. But we don't know if it is because of the service we created or because of Redis. Let's look at the CPU usage of the pods.

```sh
$ k top po -lapp=redis-api
```

Doesn't really look like the problem is with the services, but just to be sure let's scale number of redis instances from 3 to 6.

```sh
$ k scale deployment redis-api --replicas=6
```

`wait until everything seems to be running`

It doesn't seem that this affects the performance of the API at all. Let's scale everything down to something that works. And let's try to make the API scale.

## Session 2

OK. So we need to build a version of the api to support the solution I just presented. To support this we need to

1. setup a read-only side container, and
2. rewrite the API to write to the master instance and read for the local read-only instances

We start with the latter. The change is pretty simple. First, we need to rewrite the input functions

```go
var (
    readRedisAddr = flag.String("read-redis-addr", "localhost:6379", "tcp addr to connect redis tor")
    readRedisPass = flag.String("read-redis-pass", "", "password to redis server")

    writeRedisAddr = flag.String("write-redis-addr", "localhost:6379", "tcp addr to connect redis tor")
    writedRedisPass = flag.String("write-redis-pass", "", "password to redis server")
)

func init() {
    flag.Parse()

    if readRedisAddr == nil || *readRedisAddr == "" {
        panic("read Reids addr cannot be empty")
    }

  if writeRedisAddr == nil || *writeRedisAddr == "" {
        panic("write Reids addr cannot be empty")
    }
}
```

Make sure that the appropriate redis client is set up

```go
readClient, err := setupredis.NewWait(*readRedisAddr, *readRedisPass)
if err != nil {
  panic(err)
}
writeClient, err := setupredis.NewWait(*writeRedisAddr, *writeRedisPass)
if err != nil {
  panic(err)
}
```

and pass the appropriate clients to the right endpoints.

```go
getHandleFunc := get(readClient)
postHandleFunc := post(writeClient)
```

Now that that is settled. We need to specify a side container running Redis. Luckily that is pretty similar to specifying the main container.

```yaml
- name: redis-sidekick
  imagePullPolicy: Always
  image: gronnbeck/redis-with-config
  resources:
    limits:
      cpu: 2
      memory: 400Mi
    requests:
      cpu: 200m
      memory: 100Mi
  volumeMounts:
    - name: redis-config
      mountPath: /usr/local/etc/redis
```

Since the main container will have direct access to the side container we just set up. And the redis instance will bind to `localhost:6379`. Let's update the Kubernetes deployment manifest with the right environment variables.

```yaml
env:
- name: REDIS_ADDR_READ
  value: 127.0.0.1:6379
- name: REDIS_ADDR_WRITE
  value: 127.0.0.1:7007
- name: REDIS_PASS_WRITE
  valueFrom:
    configMapKeyRef:
      name: redis-write
      key: pass
```

and make `./run` use those variables.

```bash
./redisapi \
  --read-redis-addr=$REDIS_ADDR_READ --read-redis-pass=$REDIS_PASS_READ \
  --write-redis-addr=$REDIS_ADDR_WRITE --write-redis-addr=$REDIS_PASS_WRITE
```

OK! Time for deployment. Like earlier we need to build and push it to the docker registry

```sh
$ ./scripts/build.sh
$ ./scripts/push-build.sh
```

and deploy the thing

```sh
$ k apply -f ./manifest/improved-deployment.yaml
```

Let's see that everything runs stably.

x: {3000, 6000, 12000, 24000, 48000, 96000, 200000, 400000, 800000}

```sh
$ k scale deployment vegetaserver --replicas=x/1000
```

```sh
$ k scale deployment redis-api --replicas=3
```

## Shutdown

Shutdown cluster by running

```sh
$ ./scripts/stop-cluster.sh ndc
```
