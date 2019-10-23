Dynacasbin
==========

Dynacasbin is the [DynamoDB](https://aws.amazon.com/dynamodb/) adapter for [Casbin](https://github.com/casbin/casbin). With this library, Casbin can load policy from DynamoDB and auto save during add/remove policy.

> code is inspired by [github.com/hooqtv/dynacasbin](https://github.com/HOOQTV/dynacasbin), and autoSave support for it.

## Installation

    go get github.com/newbmiao/dynacasbin

## Simple Example

```go
package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/casbin/casbin"
	"github.com/newbmiao/dynacasbin"
)

func main() {
	// Initialize a DynamoDB adapter and use it in a Casbin enforcer:
	//use aws credentials default
	config := &aws.Config{
		Region: aws.String("us-east-2"), // your region
	} // Your AWS configuration
	ds := "casbin-rules"
	a, err := dynacasbin.NewAdapter(config, ds) // Your aws configuration and data source.
	if err != nil {
		panic(err)
	}
	e := casbin.NewEnforcer("examples/rbac_model.conf", a)

	// Since autoSave is support.No need use LoadPolicy()
	//e.LoadPolicy()

	// Check the permission.
	if e.Enforce("alice", "data1", "read") {
		fmt.Println("alice can read data1")
	} else {
		fmt.Println("alice can not read data1")
	}

	// Modify the policy. autoSave is support
	e.AddPolicy("jack", "data3", "read")
	e.RemovePolicy("alice", "data1", "read")
	e.RemoveFilteredPolicy(0, "data2_admin")

	fmt.Println(e.GetPolicy())

	// Since autoSave is support.No need use SavePolicy()
	//e.SavePolicy()
}
```
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
