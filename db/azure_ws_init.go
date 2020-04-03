package db

import (
	"encoding/json"
	"github.com/stikkireddy/databricks-tf-provider/client"
	"github.com/stikkireddy/databricks-tf-provider/client/service"
	"log"
	"net/http"
)

const (
	ManagementResourceEndpoint string = "https://management.core.windows.net/"
	ADBResourceID              string = "2ff814a6-3304-4ab8-85cb-cd0e6f879c1d"
)

type AzureAuth struct {
	TokenPayload           *TokenPayload
	ManagementToken        string
	AdbWorkspaceResourceID string
	AdbAccessToken         string
	AdbPlatformToken       string
}

type TokenPayload struct {
	ManagedResourceGroup string
	AzureRegion          string
	WorkspaceName        string
	ResourceGroup        string
	SubscriptionId       string
	ClientSecret         string
	ClientID             string
	TenantID             string
}

type WsProps struct {
	ManagedResourceGroupID string `json:"managedResourceGroupId"`
}

type WorkspaceRequest struct {
	Properties *WsProps `json:"properties"`
	Name       string   `json:"name"`
	Location   string   `json:"location"`
}

func (a *AzureAuth) getManagementToken(option client.DBClientOption) error {
	log.Println("Creating Azure Databricks management OAuth token.")
	url := "https://login.microsoftonline.com/" + a.TokenPayload.TenantID + "/oauth2/token"
	payload := "grant_type=client_credentials&client_id=" + a.TokenPayload.ClientID + "&client_secret=" + a.TokenPayload.ClientSecret + "&resource=" + ManagementResourceEndpoint
	headers := map[string]string{
		"Content-Type":  "application/x-www-form-urlencoded",
		"cache-control": "no-cache",
	}

	var responseMap map[string]interface{}
	resp, err := client.PerformQuery(option, http.MethodPost, url, "2.0", headers, false, true, payload)
	if err != nil {
		return err
	}
	err = json.Unmarshal(resp, &responseMap)
	if err != nil {
		return err
	}
	a.ManagementToken = responseMap["access_token"].(string)
	return nil
}

func (a *AzureAuth) getWorkspaceId(option client.DBClientOption) error {
	log.Println("Getting Workspace ID via management token.")
	url := "https://management.azure.com/subscriptions/" + a.TokenPayload.SubscriptionId + "/resourceGroups/" + a.TokenPayload.ResourceGroup + "/providers/Microsoft.Databricks/workspaces/" + a.TokenPayload.WorkspaceName + "" +
		"?api-version=2018-04-01"

	payload := &WorkspaceRequest{
		Properties: &WsProps{ManagedResourceGroupID: "/subscriptions/" + a.TokenPayload.SubscriptionId + "/resourceGroups/" + a.TokenPayload.ManagedResourceGroup},
		Name:       a.TokenPayload.WorkspaceName,
		Location:   a.TokenPayload.AzureRegion,
	}
	headers := map[string]string{
		"Content-Type":  "application/json",
		"cache-control": "no-cache",
		"Authorization": "Bearer " + a.ManagementToken,
	}

	var responseMap map[string]interface{}
	resp, err := client.PerformQuery(option, http.MethodPut, url, "2.0", headers, true, true, payload)
	if err != nil {
		return err
	}
	err = json.Unmarshal(resp, &responseMap)
	if err != nil {
		return err
	}
	log.Println(responseMap)
	a.AdbWorkspaceResourceID = responseMap["id"].(string)
	return err
}

func (a *AzureAuth) getADBPlatformToken(option client.DBClientOption) error {
	log.Println("Creating Azure Databricks platform OAuth token.")
	url := "https://login.microsoftonline.com/" + a.TokenPayload.TenantID + "/oauth2/token"
	payload := "grant_type=client_credentials&client_id=" + a.TokenPayload.ClientID + "&client_secret=" + a.TokenPayload.ClientSecret + "&resource=" + ADBResourceID
	headers := map[string]string{
		"Content-Type":  "application/x-www-form-urlencoded",
		"cache-control": "no-cache",
	}

	var responseMap map[string]interface{}
	resp, err := client.PerformQuery(option, http.MethodPost, url, "2.0", headers, false, true, payload)
	if err != nil {
		return err
	}
	err = json.Unmarshal(resp, &responseMap)
	if err != nil {
		return err
	}
	a.AdbPlatformToken = responseMap["access_token"].(string)
	return nil
}

func (a *AzureAuth) getWorkspaceAccessToken(option client.DBClientOption) error {
	log.Println("Creating workspace token")
	apiLifeTimeInSeconds := int32(600)
	comment := "Secret made via SP"
	url := "https://" + a.TokenPayload.AzureRegion + ".azuredatabricks.net/api/2.0/token/create"
	payload := struct {
		LifetimeSeconds int32  `json:"lifetime_seconds,omitempty"`
		Comment         string `json:"comment,omitempty"`
	}{
		apiLifeTimeInSeconds,
		comment,
	}
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"X-Databricks-Azure-Workspace-Resource-Id": a.AdbWorkspaceResourceID,
		"X-Databricks-Azure-SP-Management-Token":   a.ManagementToken,
		"cache-control":                            "no-cache",
		"Authorization":                            "Bearer " + a.AdbPlatformToken,
	}

	var responseMap map[string]interface{}
	resp, err := client.PerformQuery(option, http.MethodPost, url, "2.0", headers, true, true, payload)
	if err != nil {
		return err
	}
	err = json.Unmarshal(resp, &responseMap)
	if err != nil {
		return err
	}
	a.AdbAccessToken = responseMap["token_value"].(string)
	return nil
}

func (a *AzureAuth) initWorkspaceAndGetClient(option client.DBClientOption) (service.DBApiClient, error) {
	var dbClient service.DBApiClient
	err := a.getManagementToken(option)
	if err != nil {
		return dbClient, err
	}

	err = a.getWorkspaceId(option)
	if err != nil {
		return dbClient, err
	}
	err = a.getADBPlatformToken(option)
	if err != nil {
		return dbClient, err
	}
	err = a.getWorkspaceAccessToken(option)
	if err != nil {
		return dbClient, err
	}

	var newOption client.DBClientOption
	newOption.Token = a.AdbAccessToken
	newOption.Host = "https://" + a.TokenPayload.AzureRegion + ".azuredatabricks.net"
	dbClient.Init(newOption)

	_, err = dbClient.Clusters().ListNodeTypes()
	return dbClient, err
}
