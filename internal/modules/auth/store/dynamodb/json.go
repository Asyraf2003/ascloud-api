package dynamodb

import "encoding/json"

func jsonOrEmpty(v map[string]any) (string, error) {
	if len(v) == 0 {
		return "", nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
