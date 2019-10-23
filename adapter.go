package dynacasbin

import (
	"fmt"
	"regexp"

	"crypto/md5"
	"github.com/casbin/casbin/model"
	"github.com/casbin/casbin/persist"
	"github.com/guregu/dynamo"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type (
	// Adapter structs holds dynamoDB config and service	Adapter struct {
	Adapter struct {
		Config         *aws.Config
		Service        *dynamodb.DynamoDB
		DB             *dynamo.DB
		DataSourceName string
	}

	CasbinRule struct {
		ID    string `dynamo:"ID"`    //md5 of rule
		PType string `dynamo:"PType"` //hash key
		V0    string `dynamo:"V0"`    //sort key
		V1    string `dynamo:"V1"`
		V2    string `dynamo:"V2"`
		V3    string `dynamo:"V3"`
		V4    string `dynamo:"V4"`
		V5    string `dynamo:"V5"`
	}
)

// NewAdapter is the constructor for adapter
func NewAdapter(config *aws.Config, ds string) (*Adapter, error) {
	a := &Adapter{}
	a.Config = config
	a.DataSourceName = ds
	s, err := session.NewSession(config)
	if err != nil {
		return a, err
	}
	a.Service = dynamodb.New(s, a.Config)
	s, _ = session.NewSession()
	a.DB = dynamo.New(s, a.Config)
	return a, err
}

// use md5(line) to prevent overwrites of an existing item
func generateID(line CasbinRule) string {
	data := []byte(fmt.Sprint(line))
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has)
}

func loadPolicyLine(line CasbinRule, model model.Model) {
	lineText := line.PType
	if line.V0 != "" {
		lineText += ", " + line.V0
	}
	if line.V1 != "" {
		lineText += ", " + line.V1
	}
	if line.V2 != "" {
		lineText += ", " + line.V2
	}
	if line.V3 != "" {
		lineText += ", " + line.V3
	}
	if line.V4 != "" {
		lineText += ", " + line.V4
	}
	if line.V5 != "" {
		lineText += ", " + line.V5
	}

	persist.LoadPolicyLine(lineText, model)
}

//!important: call Enforcer.LoadPolicy rather than call Adapter.LoadPolicy.
// cause call Adapter.LoadPolicy multi times will repeat policys multi times.
func (a *Adapter) LoadPolicy(model model.Model) error {
	p, err := a.getAllItems()
	if err != nil {
		panic(err)
	}

	for _, v := range p {
		loadPolicyLine(v, model)
	}

	return err
}

func savePolicyLine(ptype string, rule []string) CasbinRule {
	line := CasbinRule{}

	line.PType = ptype
	if len(rule) > 0 {
		line.V0 = rule[0]
	}
	if len(rule) > 1 {
		line.V1 = rule[1]
	}
	if len(rule) > 2 {
		line.V2 = rule[2]
	}
	if len(rule) > 3 {
		line.V3 = rule[3]
	}
	if len(rule) > 4 {
		line.V4 = rule[4]
	}
	if len(rule) > 5 {
		line.V5 = rule[5]
	}

	//set md5 id
	line.ID = generateID(line)
	return line
}

//save all policy
func (a *Adapter) SavePolicy(model model.Model) error {
	//IMPORTANT: No need use it now.
	var lines []CasbinRule

	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			line := savePolicyLine(ptype, rule)
			lines = append(lines, line)
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			line := savePolicyLine(ptype, rule)
			lines = append(lines, line)
		}
	}

	_, err := a.saveItems(lines)
	return err
}

func (a *Adapter) saveItems(rules []CasbinRule) (int, error) {
	items := make([]interface{}, len(rules))

	for i := 0; i < len(rules); i++ {
		items[i] = rules[i]
	}

	return a.DB.Table(a.DataSourceName).Batch().Write().Put(items...).Run()
}

func (a *Adapter) getAllItems() ([]CasbinRule, error) {
	var rule []CasbinRule
	err := a.DB.Table(a.DataSourceName).Scan().All(&rule)
	if err != nil {
		return nil, err
	}
	return rule, nil
}

// CreateTable has response for create new table for store
func (a *Adapter) CreateTable() (*dynamodb.CreateTableOutput, error) {
	params := &dynamodb.CreateTableInput{
		TableName: aws.String(a.DataSourceName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("ID"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("ID"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
	}

	out, err := a.Service.CreateTable(params)

	if err != nil {
		matched, err := regexp.MatchString("ResourceInUseException: Cannot create preexisting table", err.Error())
		if err != nil {
			return nil, err
		}

		if !matched {
			return nil, err
		}
	}

	return out, nil
}

// DeleteTable should delete a table
func (a *Adapter) DeleteTable() error {
	params := &dynamodb.DeleteTableInput{
		TableName: aws.String(a.DataSourceName),
	}
	_, err := a.Service.DeleteTable(params)
	return err
}

//This Err will return, if cond check is false
func isConditionalCheckErr(err error) bool {
	if ae, ok := err.(awserr.RequestFailure); ok {
		return ae.Code() == "ConditionalCheckFailedException"
	}
	return false
}

// AddPolicy adds a policy rule to the storage.
func (a *Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	item := savePolicyLine(ptype, rule)
	err := a.DB.Table(a.DataSourceName).Put(item).If("attribute_not_exists(ID)").Run()
	if isConditionalCheckErr(err) {
		return nil
	}
	return err
}

// RemovePolicy removes a policy rule from the storage.
func (a *Adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	item := savePolicyLine(ptype, rule)
	return a.DB.Table(a.DataSourceName).Delete("ID", item.ID).Run()
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.

// IMPORTANT: Use ID as primary partition key and no sort key.
// If has sort key, toggle the comment code of this func to map hash key & sort key.
func (a *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	res, err := a.getAllItems()
	if err != nil {
		return err
	}
	line := &CasbinRule{PType: ptype}

	idx := fieldIndex + len(fieldValues)
	if fieldIndex <= 0 && idx > 0 {
		line.V0 = fieldValues[0-fieldIndex]
	}
	if fieldIndex <= 1 && idx > 1 {
		line.V1 = fieldValues[1-fieldIndex]
	}
	if fieldIndex <= 2 && idx > 2 {
		line.V2 = fieldValues[2-fieldIndex]
	}
	if fieldIndex <= 3 && idx > 3 {
		line.V3 = fieldValues[3-fieldIndex]
	}
	if fieldIndex <= 4 && idx > 4 {
		line.V4 = fieldValues[4-fieldIndex]
	}
	if fieldIndex <= 5 && idx > 5 {
		line.V5 = fieldValues[5-fieldIndex]
	}
	items := make([]dynamo.Keyed, 0)
	for _, item := range res {
		if item.PType == line.PType {
			if (line.V0 != "" && line.V0 != item.V0) ||
				(line.V1 != "" && line.V1 != item.V1) ||
				(line.V2 != "" && line.V2 != item.V2) ||
				(line.V3 != "" && line.V3 != item.V3) ||
				(line.V4 != "" && line.V4 != item.V4) ||
				(line.V5 != "" && line.V5 != item.V5) {
				continue
			}
			//items = append(items, dynamo.Keys{item.ID, item.PType}) //sort key: PType
			items = append(items, dynamo.Keys{item.ID,}) // no sort key
		}
	}

	if len(items) == 0 {
		return nil
	}
	//cnt, err := a.DB.Table(a.DataSourceName).Batch("ID", "PType").Write().Delete(items...).Run() //sort key: PType
	cnt, err := a.DB.Table(a.DataSourceName).Batch("ID").Write().Delete(items...).Run() // no sort key
	if cnt == len(items) {
		return nil
	}
	return err
}
