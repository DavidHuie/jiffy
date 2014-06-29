# Jiffy

[![Build Status](https://travis-ci.org/DavidHuie/jiffy.svg?branch=master)](https://travis-ci.org/DavidHuie/jiffy)

In-process pub/sub in Go. Use this package to add one-to-many messaging
to your Go program.

## Getting the package

```shell
$ go get github.com/DavidHuie/jiffy
```

```go
import "github.com/DavidHuie/jiffy"
```

## Working with topics

To gain access to a topic, fetch it by name:

```go
topic := jiffy.GetTopic("my-topic")
```

To receive all cached messages from a topic:

```go
messages := topic.CachedData()
```

## Creating subscriptions

Create a subscription by using a unique name specific to the consumer and
a TTL that will determine when the subscription is destroyed:

```go
subscription := topic.GetSubscription("my-id", time.Minute)
```

Note: the subscription can be reused by a consumer if the subscription
is remade with the same parameters before the TTL expires.

## Receiving messages

To receive messages published to the topic, listen on the subscription's
response channel:

```go
message := <-subscription.Response
```

## Publishing messages

To publish a message to the topic, first create message with a name, a
payload, and a TTL:

```go
newMessage := jiffy.NewMessage("my-message-key", "Hello!", time.Minute)
```

To publish a message to all subscribers of a topic:

```go
topic.Publish(newMessage)
```

To publish a message to all subscribers of a topic and cache it with a TTL:

```go
topic.RecordAndPublish(newMessage, time.Minute)
```

Note: only one message per message name is cached in a topic, resulting in
newly published messages superceding previous messages in the cache if they have
the same name.

## Copyright

Copyright (c) 2014 David Huie. See LICENSE.txt for further details.