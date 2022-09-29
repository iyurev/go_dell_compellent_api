package compellent_api

import (
	"fmt"
	"k8s.io/klog"
	net2 "net"
	"regexp"
	"strconv"
)

// NewSimpleNFSExport - Create  data structure for simple NFS export
func NewSimpleNFSExport(nasVolId int, cluster_id string, access_details []AccessDetails) FluidFsNfsExport {
	return FluidFsNfsExport{
		ClusterId:           cluster_id,
		NasVolumeId:         nasVolId,
		FolderPath:          "/",
		KerberosV5:          false,
		KerberosV5Integrity: false,
		KerberosV5Privacy:   false,
		UnixStyle:           true,
		AccessDetails:       access_details,
	}
}

// NewSimpleAccessDetails - Create simple AccessDetails - it allow RW for anybody
func NewSimpleAccessDetails(network string, mask int) AccessDetails {
	return AccessDetails{
		ExportTo:        "ClientsInNetwork",
		ExportToClients: network,
		ExportToPrefix:  mask,
		ReadWrite:       true,
		TrustUsers:      "NoRootSquash",
	}
}

// AccessDetailtFromSC - Parse string like "192.168.1.0/24, 192.168.0.1/32" from storage class parameter and return AccessDetails for REST API
func AccessDetailtFromSC(accessParameter string) ([]AccessDetails, error) {
	var accessDetails []AccessDetails
	stripWhitespacesReg, err := regexp.Compile("[\t\n\f\r ]")
	if err != nil {
		return []AccessDetails{}, err
	}
	clearStr := stripWhitespacesReg.ReplaceAllString(accessParameter, "")
	if splReg, err := regexp.Compile("(\\,+)"); err != nil {
		return []AccessDetails{}, err
	} else {
		splSlice := splReg.Split(clearStr, -1)
		for _, v := range splSlice {
			ip, netw, err := net2.ParseCIDR(v)
			if err != nil {
				return []AccessDetails{}, err
			}

			if netPrefReg, err := regexp.Compile("\\/"); err != nil {
				return []AccessDetails{}, err
			} else {

				netPref, _ := strconv.Atoi(netPrefReg.Split(v, -1)[1])
				if netw.Mask[3] == 255 {
					accessDetail := AccessDetails{
						ExportTo:        "OneClient",
						ExportToClients: ip.String(),
						ExportToPrefix:  0,
						ReadWrite:       true,
						TrustUsers:      "NoRootSquash",
					}
					accessDetails = append(accessDetails, accessDetail)
				} else {
					accessDetail := AccessDetails{
						ExportTo:        "ClientsInNetwork",
						ExportToClients: ip.String(),
						ExportToPrefix:  netPref,
						ReadWrite:       true,
						TrustUsers:      "NoRootSquash",
					}
					accessDetails = append(accessDetails, accessDetail)
				}

			}
		}
	}
	if len(accessDetails) == 0 {
		return []AccessDetails{}, fmt.Errorf("can't parse access details")
	}
	return accessDetails, nil
}

func NewSimpleNasVolume(clusterId, volName string, volSize int64, folderId int) *FluidFsNasVolume {
	volSize = volSize / 1024 * 2
	strVolSize := strconv.FormatInt(volSize, 10)
	return &FluidFsNasVolume{ClusterId: clusterId,
		Name:                         volName,
		Size:                         strVolSize,
		NasVolumeFolderId:            folderId,
		AclToUnix777MappingEnabled:   false,
		InteroperabilityPolicy:       "Unix",
		DefaultUnixFilePermissions:   "0775",
		DefaultUnixFolderPermissions: "0775"}

}

func (comp *CompelentREST) CreateNfsPV(clusterName, pvName string, pvSize int64, folderName string, accessParameters string) error {
	fluidCluster, err := comp.GetFluidFsClusterInfo(clusterName)
	if err != nil {
		return err
	}

	if volFolder, err := comp.GetFluidNASVolumeFolder(folderName, fluidCluster.ClusterId); err != nil {
		return err
	} else {
		nasVol := NewSimpleNasVolume(fluidCluster.ClusterId, pvName, pvSize, volFolder.FolderId)

		existsVolumes, err := comp.GetNASVolume(nasVol.Name, fluidCluster.ClusterId, volFolder.FolderId)
		if err != nil {
			return err
		}
		if len(existsVolumes) != 0 {
			klog.Infof("Volume %s is already exists. But we're continue...", nasVol.Name)
			nasVol = &existsVolumes[0]
		}

		if len(existsVolumes) == 0 {
			if err := comp.CreateFluidNASVolume(nasVol); err != nil {
				return err
			}
		}
		accessDetails, err := AccessDetailtFromSC(accessParameters)
		if err != nil {
			return err
		}
		nfsExports, err := comp.GetNFSExport(pvName, fluidCluster.ClusterId)
		if err != nil {
			return err
		}
		if len(nfsExports) != 0 {
			klog.Infof("NFS export for volume %s is already exists..\n", nasVol.Name)
			return nil
		}
		nfsExport := NewSimpleNFSExport(nasVol.NasVolumeId, fluidCluster.ClusterId, accessDetails)

		if err := comp.CreateFluidFsNfsExport(&nfsExport); err != nil {
			return err
		}
	}
	return nil
}

func (comp *CompelentREST) RemoveNfsPV(clusterName, pvName string) error {
	fluidCluster, err := comp.GetFluidFsClusterInfo(clusterName)
	if err != nil {
		return err
	}
	//Get NAS Volume data by name
	nasVolumes, err := comp.GetNASVolume(pvName, fluidCluster.ClusterId, -1)
	if err != nil {
		return err
	}
	//Check if current volume exists
	if len(nasVolumes) == 1 {
		//Get all nfs exports for founded nas volume
		nfsExports, err := comp.GetNFSExport(pvName, fluidCluster.ClusterId)
		if err != nil {
			return err
		}
		//Delete every NFS export for volume
		if len(nfsExports) != 0 {
			for _, export := range nfsExports {
				if err := comp.DeleteFluidFsNfsExport(&export); err != nil {
					return err
				}
			}
		}
		if len(nfsExports) == 0 {
			klog.Infof("There is no NFS exports for volume: %s\n", pvName)
		}
		//Delete NAS Volume
		if nasVolumes[0].Name == pvName {
			if err := comp.DeleteNasVolume(&nasVolumes[0]); err != nil {
				return err
			}
		}
		return nil
	}
	klog.Infof("Theres is no NAS volume for given persistent volume name: %s\n", pvName)
	return nil

}
