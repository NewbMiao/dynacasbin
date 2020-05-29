Dynacasbin
==========

Dynacasbin is the [DynamoDB](https://aws.amazon.com/dynamodb/) adapter for [Casbin](https://github.com/casbin/casbin). With this library, Casbin can load policy from DynamoDB and auto save during add/remove policy.

> code is inspired by [github.com/hooqtv/dynacasbin](https://github.com/HOOQTV/dynacasbin), and autoSave support for it.

## Installation

```go
go get github.com/newbmiao/dynacasbin
```

## Simple Example

see more detail in [examples](/examples)

## Notes

-  No need use LoadPolicy and SavePolicy.

    Auto save is supported now.
    SavePolicy is overwrite by just batch putItems without recreate table and adapter.loadPolicy after.
    Cause dynamodb recreate table has latency, 
    and adapter.loadPolicy is repeat not reload when call it multi times.
    which is unreliable.

- About policy unique

    ID is hash key which value is unique by use md5(policy)
- About RemoveFilteredPolicy

    RemoveFilteredPolicy is implement by getAllItems and then filter items to delete.
    This may has latency when data is so big. Use as appropriate.

## Getting Help

- [Casbin](https://github.com/casbin/casbin)
- [guregu/dynamo](https://github.com/guregu/dynamo)

## Other Choice

Wana some other choice?

There is a flexible and powerful open policy engine, named [OPA](https://www.openpolicyagent.org/),
 you can refer to [opa-koans](https://github.com/NewbMiao/opa-koans) and get start from there.
