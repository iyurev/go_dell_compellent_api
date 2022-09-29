package compellent_api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Rest API  endpoints.
const AuthPath = "/ApiConnection/Login"
const FluidVolFolderPath = "/FluidFs/FluidFsNasVolumeFolder"
const FluidNasVolumePath = "/FluidFs/FluidFsNasVolume"
const FluidFsFluidFsCluster = "/FluidFs/FluidFsCluster"
const FluidFsNfsExportPath = "/FluidFs/FluidFsNfsExport"

const RequestTimeout = 120

func (r *Response) Error() string {
	return fmt.Sprintf("Bad response!! Response body: %s\n Response code: %d\n", r.Body, r.StatusCode)
}

//////////////////////////

type CompelentREST struct {
	Headers     http.Header
	RestClient  *http.Client
	RestBaseUrl string
	InstId      string
	EmptyBody   []byte
}

//NewCompelentREST - Create new http client + particular headers  for work with dell storage api
func NewCompelentREST(baseurl, username, password string, port int) (*CompelentREST, error) {
	ignoreSslCheck := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	restUrl := fmt.Sprintf("https://%s:%d/api/rest", baseurl, port)
	authUrl := fmt.Sprintf("%s%s", restUrl, AuthPath)
	client := http.Client{Transport: ignoreSslCheck, Timeout: time.Second * RequestTimeout}
	authReq, err := http.NewRequest("POST", authUrl, nil)
	if err != nil {
		return &CompelentREST{}, err
	}
	//Add headers
	baseHeaders := http.Header{}
	baseHeaders.Add("Accept", "application/json")
	baseHeaders.Add("Content-Type", "application/json")
	baseHeaders.Add("x-dell-api-version", "4.0")
	authReq.Header = baseHeaders
	//Add basic auth header
	authReq.SetBasicAuth(username, password)

	resp, err := client.Do(authReq)
	if err != nil {
		return &CompelentREST{}, err
	}
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &CompelentREST{}, err
	}
	if resp.StatusCode != 200 {
		return &CompelentREST{}, fmt.Errorf("Bad auth request!!!, Responce body: %s\n", respData)
	}
	var login_resp_data LoginResp
	unm_err := json.Unmarshal(respData, &login_resp_data)
	if unm_err != nil {
		log.Fatal(unm_err)
	}
	session_secret := resp.Header.Get("Set-Cookie")
	if session_secret != "" {
		authReq.Header.Add("Cookie", session_secret)
		return &CompelentREST{
			Headers:     authReq.Header,
			RestBaseUrl: restUrl,
			RestClient:  &client,
			InstId:      login_resp_data.InstanceId,
			EmptyBody:   []byte("{}"),
		}, nil
	} else {
		return &CompelentREST{}, fmt.Errorf("Empty session secret header!!")
	}
}

// GetFluidNASVolumeFolder - Get NAS volume folder info by volume name
func (comp *CompelentREST) GetFluidNASVolumeFolder(name, cluster_id string) (*FluidNasVolumeFolder, error) {
	if name != "" {
		url := fmt.Sprintf("%s%s/GetList", comp.RestBaseUrl, FluidVolFolderPath)
		filter, err := NewSimpleFilter(map[string]string{"Name": name, "clusterId": cluster_id})
		if err != nil {
			return &FluidNasVolumeFolder{}, err
		}
		response, err := comp.Request(url, "POST", filter)
		if err != nil {
			return &FluidNasVolumeFolder{}, err
		}
		if response.StatusCode == 200 {
			var vol_folders []FluidNasVolumeFolder
			if err := json.Unmarshal(response.Body, &vol_folders); err != nil {
				return &FluidNasVolumeFolder{}, err
			}
			if len(vol_folders) == 1 {
				return &vol_folders[0], nil
			}
			return &FluidNasVolumeFolder{}, fmt.Errorf("Can't find volume folder: %s\n", name)
		}
		return &FluidNasVolumeFolder{}, response
	}
	return &FluidNasVolumeFolder{}, fmt.Errorf("Empty volume folder name!!")
}

// GetFluidFsClusterInfo - Get FluidFS cluster info  by cluster name
func (comp *CompelentREST) GetFluidFsClusterInfo(clusterName string) (*FluidFSCluster, error) {
	filter, err := NewSimpleFilter(map[string]string{"instanceName": clusterName})
	if err != nil {
		return &FluidFSCluster{}, err
	}
	url := fmt.Sprintf("%s%s/GetList", comp.RestBaseUrl, FluidFsFluidFsCluster)
	response, err := comp.Request(url, "POST", filter)
	if err != nil {
		return &FluidFSCluster{}, err
	}
	var fluidfs_clusters []FluidFSCluster
	if response.StatusCode == 200 {
		if err := json.Unmarshal(response.Body, &fluidfs_clusters); err != nil {
			return &FluidFSCluster{}, err
		}
	}
	if len(fluidfs_clusters) == 1 {
		return &fluidfs_clusters[0], nil
	}
	return &FluidFSCluster{}, response
}

// CreateFluidNASVolume - Create NAS volume
func (comp *CompelentREST) CreateFluidNASVolume(volume *FluidFsNasVolume) error {
	url := fmt.Sprintf("%s%s", comp.RestBaseUrl, FluidNasVolumePath)
	requestBody, err := json.Marshal(volume)
	if err != nil {
		return err
	}
	response, err := comp.Request(url, "POST", requestBody)
	if err != nil {
		return err
	}
	if response.StatusCode == 201 {
		err := json.Unmarshal(response.Body, volume)
		if err != nil {
			return err
		}
		return nil
	}
	return response
}

// CreateFluidFsNfsExport - Create NFS export for NAS volume
func (comp *CompelentREST) CreateFluidFsNfsExport(nfsExport *FluidFsNfsExport) error {
	url := fmt.Sprintf("%s%s", comp.RestBaseUrl, FluidFsNfsExportPath)
	requestBody, err := json.Marshal(nfsExport)
	if err != nil {
		return err
	}
	response, err := comp.Request(url, "POST", requestBody)
	if err != nil {
		return err
	}
	if response.StatusCode == 201 {
		if err := json.Unmarshal(response.Body, nfsExport); err != nil {
			return err
		}
		return nil
	}
	return response
}

// DeleteNasVolume - DELETE api/rest/FluidFs/FluidFsNasVolume/{InstanceId}
func (comp *CompelentREST) DeleteNasVolume(nasVolume *FluidFsNasVolume) error {
	url := fmt.Sprintf("%s%s/%s", comp.RestBaseUrl, FluidNasVolumePath, nasVolume.InstanceId)
	response, err := comp.Request(url, "DELETE", comp.EmptyBody)
	if err != nil {
		return err
	}
	if response.StatusCode == 200 {
		return nil
	}
	return response
}

// DeleteFluidFsNfsExport - DELETE api/rest/FluidFs/FluidFsNfsExport/{InstanceId}
func (comp *CompelentREST) DeleteFluidFsNfsExport(nfsExport *FluidFsNfsExport) error {
	url := fmt.Sprintf("%s%s/%s", comp.RestBaseUrl, FluidFsNfsExportPath, nfsExport.InstanceId)
	response, err := comp.Request(url, "DELETE", comp.EmptyBody)
	if err != nil {
		return err
	}
	if response.StatusCode == 200 {
		return nil
	}
	return response
}

// GetNASVolume - Get NAS volume info by name
func (comp *CompelentREST) GetNASVolume(volumeName, clusterId string, volFolderId int) ([]FluidFsNasVolume, error) {
	var nas_volumes []FluidFsNasVolume
	url := fmt.Sprintf("%s%s/GetList", comp.RestBaseUrl, FluidNasVolumePath)
	filter, err := NewSimpleFilter(map[string]string{"Name": volumeName, "clusterId": clusterId})
	if volFolderId != -1 {
		volFolder := strconv.Itoa(volFolderId)
		filter, err = NewSimpleFilter(map[string]string{"instanceName": volumeName, "clusterId": clusterId, "nasVolumeFolderId": volFolder})
	}
	if err != nil {
		return nas_volumes, err
	}
	response, err := comp.Request(url, "POST", filter)
	if err != nil {
		return nas_volumes, err
	}
	if response.StatusCode == 200 {

		if err := json.Unmarshal(response.Body, &nas_volumes); err != nil {
			return nas_volumes, err
		}
		return nas_volumes, nil
	}
	return nas_volumes, response
}

// GetNFSExport - Get NFS export by  underlying NAS volume name
func (comp *CompelentREST) GetNFSExport(volumeName, clusterId string) ([]FluidFsNfsExport, error) {
	url := fmt.Sprintf("%s%s/GetList", comp.RestBaseUrl, FluidFsNfsExportPath)
	var nfsExports []FluidFsNfsExport
	filter, err := NewSimpleFilter(map[string]string{"VolumeName": volumeName, "clusterId": clusterId})
	if err != nil {
		return []FluidFsNfsExport{}, err
	}
	response, err := comp.Request(url, "POST", filter)
	if err != nil {
		return []FluidFsNfsExport{}, err
	}
	if response.StatusCode != 200 {
		return []FluidFsNfsExport{}, response
	}
	if err := json.Unmarshal(response.Body, &nfsExports); err != nil {
		return []FluidFsNfsExport{}, err
	}

	return nfsExports, nil
}

func NewSimpleFilter(kvs map[string]string) ([]byte, error) {
	if len(kvs) != 0 {
		var filters []FilterItem
		for k, v := range kvs {
			filter := FilterItem{
				AttributeValue: v,
				AttributeName:  k,
				FilterType:     "Equals",
			}
			filters = append(filters, filter)
		}
		filterList := Filter{
			FilterType: "AND",
			Filters:    filters,
		}
		filter := map[string]Filter{"filter": filterList}
		jsonFilter, err := json.Marshal(filter)
		if err != nil {
			return []byte{}, err
		}
		return jsonFilter, nil

	}
	return nil, fmt.Errorf("Empty input conditions!!")
}

// Request - just calls http client request.
func (comp *CompelentREST) Request(url, method string, body []byte) (*Response, error) {
	request, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return &Response{}, err
	}
	request.Header = comp.Headers
	response, err := comp.RestClient.Do(request)
	if err != nil {
		return &Response{}, err
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return &Response{}, err
	}
	return &Response{
		StatusCode: response.StatusCode,
		Body:       responseData,
	}, nil
}
