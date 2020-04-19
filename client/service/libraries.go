package service

import (
	"encoding/json"
	"github.com/databrickslabs/databricks-terraform/client/model"
	"net/http"
)

// TokensAPI exposes the Secrets API
type LibrariesAPI struct {
	Client DBApiClient
}

func (a LibrariesAPI) Create(clusterId string, libraries []model.Library) error {
	var libraryInstallRequest = struct {
		ClusterId string          `json:"cluster_id,omitempty" url:"cluster_id,omitempty"`
		Libraries []model.Library `json:"libraries,omitempty" url:"libraries,omitempty"`
	}{
		ClusterId: clusterId,
		Libraries: libraries,
	}

	_, err := a.Client.performQuery(http.MethodPost, "/libraries/install", "2.0", nil, libraryInstallRequest, nil)

	return err
}

func (a LibrariesAPI) Delete(clusterId string, libraries []model.Library) error {
	var libraryInstallRequest = struct {
		ClusterId string          `json:"cluster_id,omitempty" url:"cluster_id,omitempty"`
		Libraries []model.Library `json:"libraries,omitempty" url:"libraries,omitempty"`
	}{
		ClusterId: clusterId,
		Libraries: libraries,
	}

	_, err := a.Client.performQuery(http.MethodPost, "/libraries/uninstall", "2.0", nil, libraryInstallRequest, nil)

	return err
}

func (a LibrariesAPI) List(clusterId string) ([]model.LibraryStatus, error) {
	var libraryStatusListResp struct {
		ClusterId       string                `json:"cluster_id,omitempty" url:"cluster_id,omitempty"`
		LibraryStatuses []model.LibraryStatus `json:"library_statuses,omitempty" url:"libraries,omitempty"`
	}
	var libraryInstallRequest = struct {
		ClusterId string `json:"cluster_id,omitempty" url:"cluster_id,omitempty"`
	}{
		ClusterId: clusterId,
	}

	resp, err := a.Client.performQuery(http.MethodGet, "/libraries/cluster-status", "2.0", nil, libraryInstallRequest, nil)
	if err != nil {
		return libraryStatusListResp.LibraryStatuses, err
	}

	err = json.Unmarshal(resp, &libraryStatusListResp)

	return libraryStatusListResp.LibraryStatuses, err
}
