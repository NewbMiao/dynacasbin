DynamoDB Adapter
====

DynamoDB Adapter is the [DynamoDB](https://aws.amazon.com/dynamodb/) adapter for [Casbin](https://github.com/casbin/casbin). With this library, Casbin can load policy from DynamoDB or save policy to it.

> code is inspired by [github.com/hooqtv/dynacasbin](https://github.com/HOOQTV/dynacasbin), and autoSave support for it.

## Installation

    go get github.com/newbmiao/dynacasbin

## Simple Example

```go
package main

import (
	"github.com/casbin/casbin"
	"github.com/newbmiao/dynacasbin"
	"github.com/aws/aws-sdk-go/aws"
)

func main() {
	// Initialize a DynamoDB adapter and use it in a Casbin enforcer:
	config := &aws.Config{} // Your AWS configuration
	ds := "casbin-rules"
	a := dynacasbin.NewAdapter(config, ds) // Your aws configuration and data source.
	e := casbin.NewEnforcer("examples/rbac_model.conf", a)

	// Since autoSave is support.No need use LoadPolicy()
	//e.LoadPolicy()

	// Check the permission.
	e.Enforce("alice", "data1", "read")

	// Modify the policy. autoSave is support
	e.AddPolicy("jack", "data3", "read")
	e.RemovePolicy("alice", "data1", "read")
	e.RemoveFilteredPolicy(0, "data2_admin")

	// Since autoSave is support.No need use SavePolicy(), cause recreate table has latency, will be failed
	//e.SavePolicy()
}
```
## Notes
-  No need use LoadPolicy and SavePolicy. 
SavePolicy is overwrite to unimplement now. Cause dynamodb recreate table has latency, 
which is unreliable.

- About RemoveFilteredPolicy
RemoveFilteredPolicy is implement by getAllItems and then filter items to delete.
This may has latency when data is so big. Use as appropriate. 
 
## Getting Help

- [Casbin](https://github.com/casbin/casbin)
- [guregu/dynamo](https://github.com/guregu/dynamo)