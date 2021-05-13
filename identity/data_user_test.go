package identity

import (
	"testing"

	"github.com/databrickslabs/terraform-provider-databricks/qa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)



func TestDataSourceUser(t *testing.T) {
	d, err := qa.ResourceFixture{
		Fixtures: []qa.HTTPFixture{
			{
				Method:   "GET",
				Resource: "/api/2.0/preview/scim/v2/Users?filter=userName%20eq%20%27mr.test%40example.com%27",
				Response: []ScimUser{
					{
						ID:       "123",
						UserName: "mr.test@example.com",
					},
				},
			},
		},
		Read:        true,
		NonWritable: true,
		Resource:    DataSourceUser(),
		ID:          ".",
		State: map[string]interface{}{
			"user_name": "ds",
		},
	}.Apply(t)
	require.NoError(t, err)
	assert.Equal(t, "123", d.Id())
	assert.Equal(t, d.Get("user_name"), "mr.test@example.com")
	assert.Equal(t, d.Get("home"), "/Users/mr.test@example.com")
	assert.Equal(t, d.Get("alphanumeric"), "mr_test")
}
