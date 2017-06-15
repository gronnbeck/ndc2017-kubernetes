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
$ k apply -f ./manifest/helloworld.yaml
```

#### Start vegeta for helloworldserver

```sh
$ k apply -f ./manifest/helloworld-vegeta.yaml
```

#### Start aggregator

```sh
$ k apply -f ./manifest/aggregator.yaml
```

#### Setup aggregator port-forward

```sh
$ k get po
$ k port-forward <pod-name> 8089:8082
```

#### Serve performance dashboard

```sh
$ cd dev/talks/high-performance-api-on-kubernetes/01-the-tools/dashboard
$ serve
```

#### Serve currency

```sh
$ cd talks/high-performance-api-on-kubernetes/01-the-tools/caseui
$ ./serve
```

## Session 1

Here we go. So I've cheated a bit and prepared a framework we can use to get started. If I hadn't we wouldn't have been able to finish on time. So let's quickly step through what I've prepared.

`show terminal`

As you can see I've started a Kubernetes cluster on Google Cloud with fifty nodes. Each node has 16 cores, in other words, we have 800 cores to our disposal. It's a larger cluster than I usually work with so I'm a bit excited as well.

I've also loaded a few configs and secrets into the cluster, we'll need them during the coding session. All of them are related to the Redis database we are going to use. Let's leave that aside for now and just focus on the code I've prepared instead.

`show main.go`

I've setup a simple `http` server, with an endpoint and an `handler` for root (`/`) accepting both `GET` and `POST`. I've made the endpoint to timeout if it has been running for more than `400 ms`. It's a bit hacky to do so for an entire endpoint. But it will save us a lot of time debugging if something isn't working properly.

In this session we are going to 1) setup the redis client, 2) implement both the `get` and `post` functions, And 3) get it up and running on Kubernetes.

```go
var client *pool.Pool
getHandleFunc := get(client)
postHandleFunc := post(client)
```

We implement `get` in the `impl-get.go` file (step to file: in this particular function), and `post` in the `impl-post.go` file (step to file), which is identical to the `get` function except it's name `post`. And to make things easier I've implemented some helpers for working with Redis, in a file called `helpers.go`.

`show helpers.go`

Here I've implemented three simple methods `hasValue`, `getValue`, and `setValue`. And the way I handle errors in this file is not best practice, and should not be included in production code, but please forgive me for being lazy and sloppy.

OK! Let's get started with implementing this arbitrary API.

`open main.go`

First, let us setup the redis client. For this I use a simple internal package I created a long time ago called `setupredis`. It's a thin wrapper around another Redis library making it easier to setup a Redis client, reliably.

```go
client, err := setupredis.NewWait(*redisAddr, *redisPass)
if err != nil {
  panic(err)
}
```

The `NewWait` function retries until it's able to get a connection to the specified Redis instance, it fails after 5 minutes. Now that we have setup the redis client let's implement `get` and `post`, and we'll start with `get`.

`open impl-get.go`

Hey! So first let's check if we have a value using `hasValue`. If we don't have a that value. We want to return an error using the `NewErrorResponse` function, informing the user that value is not yet set.

If we have the value, we want to return a response containing the value stored in Redis.

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

We start by getting the data from the request, and then parse it into a `Request` domain structure by using the helper function `RequestFromBytes`. And then send that value to Redis.

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

The Dockerfile is pretty generic and it's designed to just execute the `./run.sh` program, rhat runs the compiled go binary. Let's look at the build script.

`open ./scripts/build.sh`

Here we build `go` with Linux architecture. So we can deploy it directly into a Linux container without having to worry about the run time environment. Being able to compile a go program on any architecture for any other supported architecture is one place where go really shines. Now, Let's look at the script that pushes the container to the registry.

`open ./scripts/push-build.sh`

This is pretty simple. It just pushes the container to the my account on docker hubs.

#### Building the container

```sh
$ ./scripts/build.sh
```

#### Pushing it to docker registry

```sh
$ ./scripts/push-build.sh
```

Since it's a bit cumbersome to write the deployment manifest for Kubernetes. I've cheated and prewritten it. Let's step through it together.

The first part describes that we're using the deployment object, and how many replicas we want to run. It's always good to run more than one instance per services. It forces you to be stateless since we shouldn't care what replica we have to access.

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

Second, we specify the container we want to deploy (That is this part). In this case we want to deploy the `redis-api:ndc` image we just registered under my DockerHub account.

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

The last thing is the setup for the SSL tunnel to the Redis instance on compose. (Show the part by marking it) But that's not important for this talk, so we're not going to dig into that.

OK! Let's deploy this baby.

`open terminal`

```sh
$ k apply -f ./manifest/deployment.yaml
$ k apply -f ./manifest/service.yaml
```

Let's check that everything is up and running as expected

```sh
$ k get po -lapp=redis-api
```

```sh
$ k get svc
```

Everything seems to be working.

and let's add something that updates the API. I have cheated here as well, and created a little something that sends a POST request to our API every second.

```sh
$ k apply -f ./manifest/updater.yaml
```

Let's check that it actually works. Jupp! It works smoothly.

And now let's check how it scales by pointing the vegeta service to attack the redis api.

```sh
$ k apply -f ./manifest/api-vegeta.yaml
```

and now let's apply some more pressure on the service by scaling up the vegetaserver each pod will increase the load by a 1000 requests per second.

x: {4000,6000,7000,8000,9000} requests per second.

```sh
$ k scale deployment vegetaserver --replicas=x/1000
```

`open performance dashboard`

We clearly hit a bottleneck here somewhere around 8000 request per second. But we don't know if it is because of the service we created (our code) or because of Redis instances. My hunch is that it's Redis since what we created is basically nothing. But to be sure we need to actually check the data.

Let's look at the CPU usage of the pods.

```sh
$ k top po -lapp=redis-api
```

(Oh. The APIs are crashing so we're not able to get any usage info. (Do a k get po))

Doesn't really look like the problem is with the services.

But just to be sure let's increase the number of API instances from 3 to 6\. And if that solves the problem I would be very surprised. And this talk would be a failure.

```sh
$ k scale deployment redis-api --replicas=6
```

`wait until everything seems to be running`

It doesn't seem to have any effect on the performance of the API at all. Let's scale everything down to something that works, say, 4000 request per second. And in the mean time try to make the API scale using a different strategy.

## Session 2

OK. So we need to build a version of the api to support the solution I just presented. To support this we need to

1. setup a read-only side container, and
2. rewrite the API to write to the master instance and read from the local read-only instances

We start with the latter. The change is pretty simple. First, we need to rewrite the input flags

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

Now that that is settled. We need to specify a side container running Redis. Luckily that is pretty similar to specifying a main container.

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

```yaml
- name: redis-config
  configMap:
    name: redis-sidekick
```

Then we need to respecify environment variables we need, we'll get back to what the values has to be, in just a minute,

```yaml
- name: REDIS_ADDR_READ
  value: ""
- name: REDIS_ADDR_WRITE
  value: ""
- name: REDIS_PASS_WRITE
  value: ""
```

The redis instance in the side container will bind to `localhost:6379` and since the main container will have direct access to that side container. We just have to specify that the `REDIS_ADDR_READ` to be `localhost:6379`. And since the `master` redis instance is the same as the one we used before. We set `REDIS_ADDR_WRITE` to `localhost:7007` and reuse the password from before.

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

Lastly, we make `./run` use the environment variables we just specified

```bash
./redisapi \
  --read-redis-addr=$REDIS_ADDR_READ --read-redis-pass=$REDIS_PASS_READ \
  --write-redis-addr=$REDIS_ADDR_WRITE --write-redis-addr=$REDIS_PASS_WRITE
```

OK! Let's build and deploy this and check if it scales better.

Like earlier we need to build and push it to the docker registry before de deply.

```sh
$ ./scripts/build.sh
$ ./scripts/push-build.sh
```

```sh
$ k apply -f ./manifest/deployment.yaml
```

Let's see if it scales as expected. First, we apply 4000 request per second. And logically it should be able to support 12 000 request per second. Since we have three redis instances now.

x: {3000, 6000, 12000, 24000, 48000, 96000, 200000, 400000, 800000}

```sh
$ k scale deployment vegetaserver --replicas=x/1000
```

OK! Assuming that each redis instance can handle about 4000 request per second. Let's deploy a 100 of them and add 400 000 req / second traffic to the service.

```sh
$ k scale deployment redis-api --replicas=3
```

```sh
$ k scale deployment redis-api --replicas=3
```

## Shutdown

Shutdown cluster by running

```sh
$ ./scripts/stop-cluster.sh ndc
```
