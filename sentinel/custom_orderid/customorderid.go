package customorderid

import (
	"fmt"
	"math/rand"
)

func GenerateCustomOrderId() string {
    randomNumber := rand.Intn(900000) + 100000
    return fmt.Sprintf("prax%d", randomNumber)
}