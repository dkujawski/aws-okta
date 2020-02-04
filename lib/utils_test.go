package lib

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/bmizerany/assert"
)

type testCaseAIDR struct {
	name        string
	given       string
	shouldError bool
}

func TestAccountIDAndRoleFromRoleARN(t *testing.T) {
	accountID := fmt.Sprintf("%012d", rand.Intn(100000000000))
	roleName := fmt.Sprintf("%03d", rand.Intn(100))

	testCases := []testCaseAIDR{
		testCaseAIDR{
			name: "format resource-id",
			given: fmt.Sprintf(
				"arn:partition:iam::%s:%s",
				accountID,
				roleName),
			shouldError: true,
		},
		testCaseAIDR{
			name: "format resource-type/resource-id",
			given: fmt.Sprintf(
				"arn:partition:iam::%s:role/%s",
				accountID,
				roleName),
			shouldError: false,
		},
		testCaseAIDR{
			name: "format resource-type:resource-id",
			given: fmt.Sprintf(
				"arn:partition:iam::%s:role:%s",
				accountID,
				roleName),
			shouldError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotAccountID, gotRoleName := accountIDAndRoleFromRoleARN(tc.given)
			if tc.shouldError {
				assert.Equal(t, "", gotAccountID)
				assert.Equal(t, tc.given, gotRoleName)
			} else {
				assert.Equal(t, accountID, gotAccountID)
				assert.Equal(t, roleName, gotRoleName)
			}
		})
	}
}
