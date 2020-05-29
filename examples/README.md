# Example

## Init Dynamodb (for first run)

use [terraform](https://www.terraform.io/docs/cli-index.html) create dynamodb: `casbin-rule`

```shell
cd examples
sh tf.sh init
sh tf.sh plan
sh tf.sh apply
```

## Run demo

```go
cd examples
go run main.go
```

would got

```shell
alice can not read data1
[[jack data3 read]]
```

would got data in dynamodb like:

```json
{
  "ID": "ea291b12f7904091f54b8a602a96e30d",
  "PType": "p",
  "V0": "jack",
  "V1": "data3",
  "V2": "read"
}
```
