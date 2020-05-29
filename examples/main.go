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
		Region: aws.String("ap-southeast-1"), // your region
	} // Your AWS configuration
	ds := "casbin-rules"
	a, err := dynacasbin.NewAdapter(config, ds) // Your aws configuration and data source.
	if err != nil {
		panic(err)
	}
	e := casbin.NewEnforcer("rbac_model.conf", a)

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
