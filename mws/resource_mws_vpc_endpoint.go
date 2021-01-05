package mws

import (
	"context"
	"fmt"

	"github.com/databrickslabs/databricks-terraform/common"
	"github.com/databrickslabs/databricks-terraform/internal"
	"github.com/databrickslabs/databricks-terraform/internal/util"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// NewVPCEndpointAPI creates VPCEndpointAPI instance from provider meta
func NewVPCEndpointAPI(ctx context.Context, m interface{}) VPCEndpointAPI {
	return VPCEndpointAPI{m.(*common.DatabricksClient), ctx}
}

// VPCEndpointAPI exposes the mws VPC endpoint API
type VPCEndpointAPI struct {
	client  *common.DatabricksClient
	context context.Context
}

// Create creates the VPC endpoint registeration process
func (a VPCEndpointAPI) Create(vpcEndpoint *VPCEndpoint) error {
	vpcEndpointAPIPath := fmt.Sprintf("/accounts/%s/vpc-endpoints", vpcEndpoint.AccountID)
	return a.client.Post(a.context, vpcEndpointAPIPath, vpcEndpoint, &vpcEndpoint)
}

// Read returns the VPCEndpoint object along with metadata and any additional errors when attaching to workspace
func (a VPCEndpointAPI) Read(mwsAcctID, vpcEndpointID string) (VPCEndpoint, error) {
	var mwsVPCEndpoint VPCEndpoint
	vpcEndpointAPIPath := fmt.Sprintf("/accounts/%s/vpc-endpoints/%s", mwsAcctID, vpcEndpointID)
	err := a.client.Get(a.context, vpcEndpointAPIPath, nil, &mwsVPCEndpoint)
	return mwsVPCEndpoint, err
}

// Delete deletes the VPCEndpoint object given a VPCEndpoint id
func (a VPCEndpointAPI) Delete(mwsAcctID, vpcEndpointID string) error {
	vpcEndpointAPIPath := fmt.Sprintf("/accounts/%s/vpc-endpoints/%s", mwsAcctID, vpcEndpointID)
	if err := a.client.Delete(a.context, vpcEndpointAPIPath, nil); err != nil {
		return err
	}
	return nil
}

// List lists all the available network objects in the mws account
func (a VPCEndpointAPI) List(mwsAcctID string) ([]VPCEndpoint, error) {
	var mwsVPCEndpointList []VPCEndpoint
	vpcEndpointAPIPath := fmt.Sprintf("/accounts/%s/vpc-endpoints", mwsAcctID)
	err := a.client.Get(a.context, vpcEndpointAPIPath, nil, &mwsVPCEndpointList)
	return mwsVPCEndpointList, err
}

// ResourceVPCEndpoint ...
func ResourceVPCEndpoint() *schema.Resource {
	s := internal.StructToSchema(VPCEndpoint{}, func(s map[string]*schema.Schema) map[string]*schema.Schema {
		// nolint
		s["vpc_endpoint_name"].ValidateFunc = validation.StringLenBetween(4, 256)
		return s
	})
	p := util.NewPairSeparatedID("account_id", "aws_vpc_endpoint_id", "/")
	return util.CommonResource{
		Schema: s,
		Create: func(ctx context.Context, d *schema.ResourceData, c *common.DatabricksClient) error {
			var vpcEndpoint VPCEndpoint
			if err := internal.DataToStructPointer(d, s, &vpcEndpoint); err != nil {
				return err
			}
			if err := NewVPCEndpointAPI(ctx, c).Create(&vpcEndpoint); err != nil {
				return err
			}
			d.Set("aws_vpc_endpoint_id", vpcEndpoint.AwsVPCEndpointID)
			p.Pack(d)
			return nil
		},
		Read: func(ctx context.Context, d *schema.ResourceData, c *common.DatabricksClient) error {
			accountID, vpcEndpointID, err := p.Unpack(d)
			if err != nil {
				return err
			}
			vpcEndpoint, err := NewVPCEndpointAPI(ctx, c).Read(accountID, vpcEndpointID)
			if err != nil {
				return err
			}
			return internal.StructToData(vpcEndpoint, s, d)
		},
		Delete: func(ctx context.Context, d *schema.ResourceData, c *common.DatabricksClient) error {
			accountID, vpcEndpointID, err := p.Unpack(d)
			if err != nil {
				return err
			}
			return NewVPCEndpointAPI(ctx, c).Delete(accountID, vpcEndpointID)
		},
	}.ToResource()
}
