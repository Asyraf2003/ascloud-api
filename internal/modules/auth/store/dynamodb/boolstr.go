package dynamodb

func boolStr(v bool) string {
	if v {
		return "1"
	}
	return "0"
}
