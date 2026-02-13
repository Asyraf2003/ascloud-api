package dynamodb

import (
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func avS(v string) types.AttributeValue { return &types.AttributeValueMemberS{Value: v} }

func avN(v int64) types.AttributeValue {
	return &types.AttributeValueMemberN{Value: strconv.FormatInt(v, 10)}
}

func getS(m map[string]types.AttributeValue, k string) string {
	v, ok := m[k].(*types.AttributeValueMemberS)
	if !ok || v == nil {
		return ""
	}
	return v.Value
}

func getN(m map[string]types.AttributeValue, k string) int64 {
	v, ok := m[k].(*types.AttributeValueMemberN)
	if !ok || v == nil {
		return 0
	}
	n, _ := strconv.ParseInt(v.Value, 10, 64)
	return n
}

func has(m map[string]types.AttributeValue, k string) bool { _, ok := m[k]; return ok }
