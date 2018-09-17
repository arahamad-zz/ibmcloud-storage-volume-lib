/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package softlayer_block

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/softlayer/softlayer-go/datatypes"
	"github.com/softlayer/softlayer-go/sl"

	"github.com/uber-go/zap"
	"github.com/arahamad/ibmcloud-storage-volume-lib/lib/provider"
	"github.com/arahamad/ibmcloud-storage-volume-lib/volume-providers/softlayer/messages"
	utils "github.com/arahamad/ibmcloud-storage-volume-lib/volume-providers/softlayer/utils"
)

func (sls *SLBlockSession) SnapshotOrder(volumeRequest provider.Volume) error {
	// Step 1- validate input which are required
	sls.logger.Info("Requested volume is:", zap.Object("Volume", volumeRequest))
	if volumeRequest.SnapshotSpace == nil {
		sls.logger.Error("No proper input, please provide volume ID and snapshot space size")
		return messages.GetUserError("E0013", nil)
	}
	volid := utils.ToInt(volumeRequest.VolumeID)
	snapshotSize := *volumeRequest.SnapshotSpace
	if volid == 0 || snapshotSize == 0 {
		sls.logger.Error("No proper input, please provide volume ID and snapshot space size")
		return messages.GetUserError("E0013", nil)
	}

	// Step 2- Get volume details
	mask := "id,billingItem[location,hourlyFlag],storageType[keyName],storageTierLevel,provisionedIops,staasVersion,hasEncryptionAtRest"
	storageObj := sls.backend.GetNetworkStorageService()
	storage, err := storageObj.ID(volid).Mask(mask).GetObject()
	if err != nil {
		return messages.GetUserError("E0011", nil, volid, "Please check the volume id")
	}
	sls.logger.Info("in SnapshotOrder Volum Object ---->", zap.Object("Volume", storage))

	// Step 3: verify original volume exists or not
	if storage.BillingItem == nil {
		return messages.GetUserError("E0014", nil, volid)
	}

	if storage.BillingItem.Location == nil || storage.BillingItem.Location.Id == nil {
		sls.logger.Error("Original Volume does not have location ID", zap.Object("Location", storage.BillingItem.Location))
		return messages.GetUserError("E0024", nil, volid)
	}
	datacenterID := *storage.BillingItem.Location.Id

	// Step 4: Get billing item category code
	if storage.BillingItem.CategoryCode == nil {
		return messages.GetUserError("E0015", nil, volid)
	}
	billingItemCategoryCode := *storage.BillingItem.CategoryCode
	order_type_is_saas := true
	if billingItemCategoryCode == "storage_as_a_service" {
		order_type_is_saas = true
	} else if billingItemCategoryCode == "storage_service_enterprise" {
		order_type_is_saas = false
	} else {
		return messages.GetUserError("E0016", nil, volid)
	}

	// Step 5: Get the product package by using billing item category code
	packageDetails, errPackage := utils.GetPackageDetails(sls.logger, sls.backend, billingItemCategoryCode)
	if errPackage != nil {
		return messages.GetUserError("E0017", nil, billingItemCategoryCode)
	}
	finalPackageID := *packageDetails.Id

	// Step 6: Get required price for snapshot space as per volume type
	finalPrices := []datatypes.Product_Item_Price{}
	if order_type_is_saas {
			volume_storage_type := *storage.StorageType.KeyName
			if strings.Contains(volume_storage_type, "ENDURANCE") {
						volumeTier := utils.GetEnduranceTierIopsPerGB(sls.logger, storage)
						finalPrices = []datatypes.Product_Item_Price{
							datatypes.Product_Item_Price{Id: sl.Int(utils.GetSaaSSnapshotSpacePrice(sls.logger, packageDetails, snapshotSize, volumeTier, 0))},
						}
			} else if strings.Contains(volume_storage_type, "PERFORMANCE") {
						if !utils.IsVolumeCreatedWithStaaS(storage) {
							return messages.GetUserError("E0018", nil, volid)
						}
						iops := utils.ToInt(*storage.ProvisionedIops)
						finalPrices = []datatypes.Product_Item_Price{
							datatypes.Product_Item_Price{Id: sl.Int(utils.GetSaaSSnapshotSpacePrice(sls.logger, packageDetails, snapshotSize, "", iops))},
						}
			} else 	{
						return messages.GetUserError("E0019", nil, volume_storage_type)
			}
		} else {	// 'storage_service_enterprise' package
		volumeTier := utils.GetEnduranceTierIopsPerGB(sls.logger, storage)
		finalPrices = []datatypes.Product_Item_Price{
			datatypes.Product_Item_Price{Id: sl.Int(utils.GetEnterpriseSpacePrice(sls.logger, packageDetails, "snapshot", snapshotSize, volumeTier))},
		}
	}
	/*
	if upgrade:
        complex_type = 'SoftLayer_Container_Product_Order_Network_Storage_Enterprise_SnapshotSpace_Upgrade'
    else:
        complex_type = 'SoftLayer_Container_Product_Order_Network_Storage_Enterprise_SnapshotSpace'
	*/

	// Step 7: Create order
	cpo := datatypes.Container_Product_Order{
		ComplexType: sl.String("SoftLayer_Container_Product_Order_Network_Storage_Enterprise_SnapshotSpace"),
		Quantity:    sl.Int(1),
		Location:    sl.String(strconv.Itoa(datacenterID)),
		PackageId:   sl.Int(finalPackageID),
		Prices:      finalPrices,
	}

	sp := &datatypes.Container_Product_Order_Network_Storage_Enterprise_SnapshotSpace{
		VolumeId:                sl.Int(volid),
		Container_Product_Order: cpo,
	}
	sls.logger.Info("Order deails ... ", zap.Object("OrderDeails", sp))
	/*orderContainer := &datatypes.Container_Product_Order_Network_Storage_Enterprise_SnapshotSpace_Upgrade{
		Container_Product_Order_Network_Storage_Enterprise_SnapshotSpace : sp1,
	}*/

	// Step 8: place order
	productOrderObj := sls.backend.GetProductOrderService()
	snOrderID, snError := productOrderObj.PlaceOrder(sp, sl.Bool(false))
	if snError != nil {
		return messages.GetUserError("E0020", snError, volid, snapshotSize)
	}
	sls.logger.Info("Successfully placed Snapshot order .... ", zap.Object("orderID", *snOrderID.OrderId), zap.Object("VolumeID", volid), zap.Object("Size", snapshotSize))
	sls.logger.Info("Snapshot order details.... ", zap.Object("orderDetails", snOrderID))
	time.Sleep(300)
	sls.logger.Info("Snapshot order details.... ", zap.Object("orderDetails", snOrderID))
	return nil
	// TODO: need to keep checking if order is ready or not
}

// Create the snapshot from the volume
func (sls *SLBlockSession) SnapshotCreate(volume *provider.Volume, tags map[string]string) (*provider.Snapshot, error) {
	// Step 1: Validate input
	if volume == nil {
		return nil, messages.GetUserError("E0011", nil, nil, "nil volume struct")
	}
	volumeID := utils.ToInt(volume.VolumeID)
	if volumeID == 0 {
		return nil, messages.GetUserError("E0011", nil, volumeID, "Not a valid volume ID")
	}

	// Step 2: Get the volume details
	block_mask := "id,billingItem[location,hourlyFlag],snapshotCapacityGb,storageType[keyName],capacityGb,originalVolumeSize,provisionedIops,storageTierLevel,osType[keyName],staasVersion,hasEncryptionAtRest"
	storageObj := sls.backend.GetNetworkStorageService()
	originalVolume, err := storageObj.ID(volumeID).Mask(block_mask).GetObject()
	if err != nil {
		return nil, messages.GetUserError("E0011", err, volumeID, "Not a valid volume ID")
	}

	// Step 3: verify original volume exists
	if originalVolume.BillingItem == nil {
		return nil, messages.GetUserError("E0014", nil, volumeID)
	}

	// Step 3: Verify that the original volume has snapshot space (needed for duplication)
	if originalVolume.SnapshotCapacityGb == nil || utils.ToInt(*originalVolume.SnapshotCapacityGb) <= 0 {
		return nil, messages.GetUserError("E0023", nil, volumeID)
	}

	newtags, _ := json.Marshal(tags)
	snapshotTags := string(newtags)
	snapshotvol, err := storageObj.ID(volumeID).CreateSnapshot(&snapshotTags)
	if err != nil {
		return nil, messages.GetUserError("E0029", err, volumeID)
	}
	sls.logger.Info("Successfully created snapshot for given volume ... ", zap.Object("VolumeID", volumeID), zap.Object("SnapshotID", snapshotvol))//*snapshotvol.Id

	// Setep 4: Converting to local type
	snapshot := &provider.Snapshot{}
	snapshot.SnapshotID = strconv.Itoa(*snapshotvol.Id)
	snapshot.SnapshotSpace = snapshotvol.CapacityGb
	snapshot.Volume = *volume
	snapshot.CreationTime, _ = time.Parse(time.RFC3339, snapshotvol.CreateDate.String())
	snapshot.SnapshotTags = tags
	return snapshot, err
}

// Delete the snapshot
func (sls *SLBlockSession) SnapshotDelete(del *provider.Snapshot) error {
	// Step 1- Validate inputes
	if del == nil {
		return messages.GetUserError("E0030", nil)
	}
	snapshotId := utils.ToInt(del.SnapshotID)
	if snapshotId == 0 {
		return messages.GetUserError("E0030", nil)
	}

	//! Step 2- Delete the snapshot from SL
	storageObj := sls.backend.GetNetworkStorageService()
	_, err := storageObj.ID(snapshotId).DeleteObject()
	if err != nil {
		return messages.GetUserError("E0031", err, snapshotId)
	}
	sls.logger.Info("Successfully deleted snapshot ....", zap.Object("SnapshotID", snapshotId))
	return nil
}

// Get the snapshot
func (sls *SLBlockSession) SnapshotGet(snapshotId string) (*provider.Snapshot, error) {
	// Step 1- Validate inputes
	snapshotID := utils.ToInt(snapshotId)
	if snapshotID == 0 {
		return nil, messages.GetUserError("E0030", nil)
	}

	// Step 2- Get the snapshot details from SL
	filter := fmt.Sprintf(`{"networkStorage":{"nasType":{"operation":"SNAPSHOT"},"id": {"operation":%d}}}`, snapshotID)
	mask := "id,username,capacityGb,createDate,snapshotCapacityGb,parentVolume[snapshotSizeBytes],parentVolume[snapshotCapacityGb],storageType[keyName],serviceResource[datacenter[name]],provisionedIops,lunId,originalVolumeName,storageTierLevel,notes"
	accService := sls.backend.GetAccountService()
	storageSnapshot, err := accService.Filter(filter).Mask(mask).GetNetworkStorage()
	if err != nil {
		return nil, messages.GetUserError("E0032", err, snapshotID)
	}
	sls.logger.Info("########======> Successfully get the snapshot details", zap.Object("snapshot", storageSnapshot[0]))
	if len(storageSnapshot) <= 0 {
		return nil, messages.GetUserError("E0032", err, snapshotID)
	}
	// Setep 3: Converting to local type
	snapshot := utils.ConvertToLocalSnapshotObject(storageSnapshot[0], SoftLayer, VolumeTypeBlock)
	return snapshot, nil
}

// Snapshot list by using tags
func (sls *SLBlockSession) SnapshotsList() ([]*provider.Snapshot, error) {
	// Step 1- Get all snapshots from the SL which belongs to a IBM Infrastructure a/c
	filter := fmt.Sprintf(`{"networkStorage":{"nasType":{"operation":"SNAPSHOT"}}}`)
	accService := sls.backend.GetAccountService()
	storageSnapshots, err := accService.Filter(filter).GetNetworkStorage()
	if err != nil {
		return nil, messages.GetUserError("E0032", err)
	}
	sls.logger.Info("Successfully got all snapshot from SL", zap.Object("snapshots", storageSnapshots))

	// convert to local type
	snList := []*provider.Snapshot{}
	for _, stSnapshot := range storageSnapshots {
			snapshot := utils.ConvertToLocalSnapshotObject(stSnapshot, SoftLayer, VolumeTypeBlock)
			snList = append(snList, snapshot)
	}
	return snList, nil
}

// List all the snapshots for a given volume
func (sls *SLBlockSession) ListAllSnapshots(volumeID string) ([]*provider.Snapshot, error) {
	// Step 1- Validate inputs
	orderID := utils.ToInt(volumeID)
	if orderID == 0 {
		return nil, messages.GetUserError("E0011", nil, "Not a valid volume ID")
	}

	// Step 2- Get volume details
	storageObj := sls.backend.GetNetworkStorageService()
	mask := "id,billingItem[location,hourlyFlag],storageType[keyName],storageTierLevel,provisionedIops,staasVersion,hasEncryptionAtRest"
	_, err := storageObj.ID(orderID).Mask(mask).GetObject()
	if err != nil {
		return nil, messages.GetUserError("E0011", err, orderID, "Not a valid volume ID")
	}

	// Step 3- Get all snapshots from a volume
	snapshotvol, err := storageObj.ID(orderID).GetSnapshots()
	if err != nil {
		return nil, messages.GetUserError("E0034", err, orderID)
	}
	sls.logger.Info("Successfully got all snapshots from given volume ID .....", zap.Object("VolumeID", orderID), zap.Object("Snapshots", snapshotvol))

	// convert to local type
	snList := []*provider.Snapshot{}
	for _, stSnapshot := range snapshotvol {
			snapshot := utils.ConvertToLocalSnapshotObject(stSnapshot, SoftLayer, VolumeTypeBlock)
			snList = append(snList, snapshot)
	}
	return snList, nil
}

func (sls *SLBlockSession) ListAllSnapshotsForVolume(volumeID string) ([]*provider.Snapshot, error) {
	sls.logger.Info("Trying to get for volume", zap.Object("volume", volumeID))
	/*orderID, _ := strconv.Atoi(volumeID)
	storageObj := sls.backend.GetNetworkStorageIscsiService()
	snapshotvol, err := storageObj.ID(orderID).GetSnapshotsForVolume()
	sls.logger.Info("snapshot details are", zap.Object("snapshotvolumesss", snapshotvol), zap.Error(err))
	*/
	orderID:= utils.ToInt(volumeID)
	storageID, errID := utils.GetStorageID(sls.backend, orderID, sls.logger)
	if errID != nil {
		return nil, messages.GetUserError("E0011", errID, orderID)
	}
	sls.logger.Info("===========> SorageID is", zap.Object("VolumeID", storageID))
	return nil, nil
}
