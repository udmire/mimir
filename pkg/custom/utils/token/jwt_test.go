package token

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Sign(t *testing.T) {
	claim := &CustomClaims{
		Admin: true,
		Name:  "mason",
	}
	claim.Subject = "test"
	sign, err := NewSigner([]byte("this-is-a-secret")).Sign(claim)
	assert.Nil(t, err)
	fmt.Println(sign)
}
