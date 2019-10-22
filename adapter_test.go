package dynacasbin

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/casbin/casbin"
	"os"
	"sort"
	"strings"
	"testing"
)

var config *aws.Config
var a *Adapter
var data [][]string

func testGetPolicy(t *testing.T, e *casbin.Enforcer, res [][]string) {
	myRes := e.GetPolicy()

	res1 := make([]string, len(res))
	myRes1 := make([]string, len(myRes))
	for k, v := range res {
		res1[k] = fmt.Sprint(v)
	}
	for k, v := range myRes {
		myRes1[k] = fmt.Sprint(v)
	}
	sort.Strings(res1)
	sort.Strings(myRes1)

	if fmt.Sprint(res1) != fmt.Sprint(myRes1) {
		t.Error("Policy: ", myRes1, ", supposed to be ", res1)
	}
}

func init() {
	// Initialize a DynamoDB adapter and use it in a Casbin enforcer:
	config = &aws.Config{
		Region:      aws.String("us-east-2"),
		Credentials: credentials.NewSharedCredentials("", ""),
	} // Your AWS configuration
	ds := "casbin-rules"
	var err error
	a, err = NewAdapter(config, ds) // Your aws configuration and data source.
	if err != nil {
		panic(err)
	}

	data = [][]string{
		{"alice", "data1", "read"},
		{"bob", "data2", "write"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
	}
	for _, item := range data {
		//init dynamodb if it's empty
		err = a.AddPolicy("p", "p", item)
		if err != nil {
			if strings.Contains(err.Error(), "ResourceNotFoundException") {
				_, err = a.CreateTable()
				fmt.Println("table not exist, creating... pls try it later")
				if err != nil {
					fmt.Println("createTable error: " + err.Error())
				}
				os.Exit(1)
			} else {
				panic(err)
			}
		}
	}
}

func isPolicyExistInDB(rules [][3]string) bool {
	res, err := a.getAllItems()
	if err != nil {
		panic("adapter getAllItems failed:" + err.Error())
	}
	isExist := false
	for _, v := range res {
		for _, item := range rules {

			if v.V0 == item[0] && v.V1 == item[1] && v.V2 == item[2] {
				isExist = true
			}
		}
	}

	return isExist
}

func TestLoadPolicy(t *testing.T) {
	e := casbin.NewEnforcer("examples/rbac_model.conf", a)
	res := make([][]string, 0)
	items, err := a.getAllItems()
	if err != nil {
		panic("adapter getAllItems failed:" + err.Error())
	}
	for _, item := range items {
		res = append(res, []string{item.V0, item.V1, item.V2})
	}
	testGetPolicy(t, e, res)
}

func TestAddPolicy(t *testing.T) {
	e := casbin.NewEnforcer("examples/rbac_model.conf", a)
	e.AddPolicy("jack", "data3", "read")
	if !isPolicyExistInDB([][3]string{{"jack", "data3", "read"}}) {
		t.Error("TestAddPolicy is not ok")
	}
}

func TestRemovePolicy(t *testing.T) {
	e := casbin.NewEnforcer("examples/rbac_model.conf", a)
	params := []interface{}{"alice", "data1", "read"}
	e.RemovePolicy(params...)

	if isPolicyExistInDB([][3]string{{"alice", "data1", "read"}}) {
		t.Error("TestRemovePolicy is not ok")
	}
}

func TestRemoveFilteredPolicy(t *testing.T) {
	e := casbin.NewEnforcer("examples/rbac_model.conf", a)
	e.RemoveFilteredPolicy(0, "data2_admin")
	//check load is ok
	if isPolicyExistInDB([][3]string{{"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}}) {
		t.Error("TestRemoveFilteredPolicy is not ok")
	}
}
