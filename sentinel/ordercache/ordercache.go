package ordercache

import (
	"encoding/json"
)

type OrderDetails struct {
    Details map[string]interface{}
}

var orderCache = make(map[int64]OrderDetails)

func IsEqual(map1, map2 map[string]interface{}) bool {
    json1, err1 := json.Marshal(map1)
    json2, err2 := json.Marshal(map2)
    if err1 != nil || err2 != nil {
        return false
    }
    return string(json1) == string(json2)
}

func GetOrderDetails(orderData map[string]interface{}) map[string]interface{} {
    orderDetails := make(map[string]interface{})
    orderDetails["permId"] = orderData["permId"]
    orderDetails["status"] = orderData["status"]
    orderDetails["filled"] = orderData["filled"]
    orderDetails["remaining"] = orderData["remaining"]
    orderDetails["avgFillPrice"] = orderData["avgFillPrice"]
    orderDetails["lastFillPrice"] = orderData["lastFillPrice"]
    return orderDetails
}

func IsDuplicate(permID int64, orderDetails map[string]interface{}) bool {
    for _, cachedOrder := range orderCache {
        if IsEqual(cachedOrder.Details, orderDetails) {
            return true
        }
    }
    return false
}

func CacheOrder(permID int64, orderDetails map[string]interface{}) {
    orderCache[permID] = OrderDetails{
        Details: orderDetails,
    }
}