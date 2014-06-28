# Jiffy

In-process pub/sub in Go.

## How it works

Jiffy allows consumers to subscribe to messages broadcast to topics.
Additionally, historical messages to topics are cached with a TTL,
making them fetchable by a subscription at any given time.

## Usage

```go
import "github.com/DavidHuie/jiffy"
```

To gain access to a topic, fetch it by name:

```go
topic := jiffy.GetTopic("my-topic")
```

To receive all cached messages from a topic:

```go
messages := topic.FetchData()
```

Now, create a subscription by using a unique name specific to the consumer and
a TTL that will determine when the subscription is destroyed:

```go
subscription := topic.GetSubscription("my-id", 60*time.Second)
```

Note: the subscription can be reused by a consumer if the subscription
is remade with the same parameters before the TTL expires.

To fetch a message published to the topic, listen on the subscription's
response channel:

```go
message := <-subscription.Response
```

To publish a message to the topic, first create message with a name and a
payload:

```go
newMessage := jiffy.NewMessage("my-message-key", "Hello!")
```

To publish a message to all subscribers of a topic:

```go
topic.Publish(newMessage)
```

To publish a message to all subscribers of a topic and cache it with a TTL:

```go
topic.RecordAndPublish(newMessage, 60*time.Second)
```

Note: only one message per message name is cached in a topic, resulting in
newly published messages superceding previous messages in the cache if they have
the same name.

## Copyright

Copyright (c) 2014 David Huie. See LICENSE.txt for further details.