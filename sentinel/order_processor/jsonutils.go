package orderprocessor

import "encoding/json"

func UnmarshallData(data []byte) (map[string]interface{}, error) {
    var receivedData map[string]interface{}
    err := json.Unmarshal(data, &receivedData)
    return receivedData, err
}

func MarshallData(data map[string]interface{}) ([]byte, error) {
    return json.Marshal(data)
}