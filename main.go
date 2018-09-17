/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package main

import (
	"fmt"
	"github.com/uber-go/zap"

	//softlayer_block "github.com/arahamad/ibmcloud-storage-volume-lib/volume-providers/softlayer/block"

	"github.com/arahamad/ibmcloud-storage-volume-lib/config"
	"github.com/arahamad/ibmcloud-storage-volume-lib/lib/provider"
	//util "github.com/arahamad/ibmcloud-storage-volume-lib/lib/utils"
	"github.com/arahamad/ibmcloud-storage-volume-lib/provider/local"
	//"github.com/arahamad/ibmcloud-storage-volume-lib/provider/registry"
	provider_util "github.com/arahamad/ibmcloud-storage-volume-lib/provider/utils"
)

func main() {
	// Prepare main logger
	loggerLevel := zap.DynamicLevel()
	loggerLevel.SetLevel(zap.InfoLevel)
	logger := zap.New(
		zap.NewJSONEncoder(zap.RFC3339Formatter("ts")),
		zap.AddCaller(),
		loggerLevel,
	).With(zap.String("name", "ibm-volume-lib/main")).With(zap.String("VolumeLib", "IKS-VOLUME-LIB"))

	// Load config file
	conf := config.ReadConfig("", logger)
	if conf == nil {
		logger.Fatal("Error loading configuration")
	}

	// Prepare provider registry
	providerRegistry, err := provider_util.InitProviders(conf, logger)
	if err != nil {
		logger.Fatal("Error configuring providers", local.ZapError(err))
	}

	//dc_name := "mex01"
	logger.Info("In main before openProviderSession call", zap.Object("providerRegistry", providerRegistry))
	sess, _, err := provider_util.OpenProviderSession(conf, providerRegistry, conf.Softlayer.SoftlayerBlockProviderName, logger)
	if err != nil {
		logger.Error("Failed to get session", zap.Object("Error", err))
		return
	}
	logger.Info("In main after openProviderSession call", zap.Object("sess", sess))
	defer sess.Close()
	logger.Info("Currently you are using provider ....", zap.Object("ProviderName", sess.ProviderName()))
	valid := true
	for valid {
		fmt.Println("\n\nSelect your choice\n 1- Get volume details \n 2- Create snapshot \n 3- list snapshot \n 4- Create volume \n 5- Snapshot details \n 6- Snapshot Order \n 7- Create volume from snapshot\n 8- Delete volume \n 9- Delete Snapshot \n 10- List all Snapshot \nYour choice?:")
		var choiceN int
		var volumeID string
		var snapshotID string
		_, er11 := fmt.Scanf("%d", &choiceN)
		if er11 != nil {
				fmt.Printf("Wrong input, please provide option in int: ")
				fmt.Printf("\n\n")
				continue
		}

		if choiceN == 1 {
				fmt.Println("You selected choice to get volume details")
				fmt.Printf("Please enter volume ID: ")
				_, er11 = fmt.Scanf("%s", &volumeID)
				volume, errr := sess.VolumeGet(volumeID)
				if errr == nil {
					logger.Info("Successfully get volume details ================>", zap.Object("Volume ID", volumeID))
					logger.Info("Volume details are: ", zap.Object("Volume", volume))
				} else {
					logger.Info("Failed to get volume details ================>", zap.Object("VolumeID", volumeID), zap.Object("Error", errr))
				}
				fmt.Printf("\n\n")
		} else if choiceN == 2 {
				fmt.Println("You selected choice to create snapshot")
				fmt.Printf("Please enter volume ID: ")
				_, er11 = fmt.Scanf("%s", &volumeID)
				volume := &provider.Volume{}
				volume.VolumeID = volumeID
				var tags map[string]string
				tags = make(map[string]string)
				tags["tag1"] = "snapshot-tag1"
				snapshot, errr := sess.SnapshotCreate(volume, tags)
				if errr == nil {
					logger.Info("Successfully created snapshot on ================>", zap.Object("VolumeID", volumeID))
					logger.Info("Snapshot details: ", zap.Object("Snapshot", snapshot))
				} else {
					logger.Info("Failed to create snapshot on ================>", zap.Object("VolumeID", volumeID), zap.Object("Error", errr))
				}
				fmt.Printf("\n\n")
		} else if choiceN == 3 {
				fmt.Println("You selected choice to list snapshot from volume\n")
				fmt.Printf("Please enter volume ID to get the snapshots: ")
				_, er11 = fmt.Scanf("%s", &volumeID)
				fmt.Printf("\n\n")
				snapshots, errr := sess.ListAllSnapshots(volumeID)
				if errr == nil {
					logger.Info("Successfully get snapshot details ================>", zap.Object("Snapshot ID", volumeID))
					logger.Info("List of snapshots ", zap.Object("Snapshots are->", snapshots))
				} else {
					logger.Info("Failed to get snapshot details ================>", zap.Object("Snapshot ID", volumeID), zap.Object("Error", errr))
				}
				fmt.Printf("\n\n")
		} else if choiceN == 4 {
			fmt.Println("You selected choice to Create volume\n")
			volume := &provider.Volume{}
			volume.VolumeType = "block"
			dcName := ""
			volSize := 0
			Iops := "0"
			tier := ""
			providerType := ""
			//volume.SnapshotSpace = 0
			var choice int
			fmt.Printf("\nPlease enter storage type choice 1- for endurance  2- for performance: ")
			_, er11 = fmt.Scanf("%d", &choice)
			if choice == 1 {
				providerType = "endurance"
				volume.ProviderType = provider.VolumeProviderType(providerType)
			} else if choice == 2 {
				providerType = "performance"
				volume.ProviderType = provider.VolumeProviderType(providerType)
			}

			fmt.Printf("\nPlease enter datacenter name like dal09, dal10 or mex01  etc: ")
			_, er11 = fmt.Scanf("%s", &dcName)
			volume.Az = dcName

			fmt.Printf("\nPlease enter volume size in GB like 20, 40 80 etc : ")
			_, er11 = fmt.Scanf("%d", &volSize)
			volume.Capacity = &volSize

			if volume.ProviderType == "performance" {
				fmt.Printf("\nPlease enter iops from 1-48000 with multiple of 100: ")
				_, er11 = fmt.Scanf("%s", &Iops)
				volume.Iops = &Iops
			}
			if volume.ProviderType == "endurance" {
				fmt.Printf("\nPlease enter tier like 0.25, 2, 4, 10 iops per GB: ")
				_, er11 = fmt.Scanf("%s", &tier)
				volume.Tier = &tier
			}
			_, errr := sess.VolumeCreate(*volume)
			if errr == nil {
				logger.Info("Successfully ordered volume ================>", zap.Object("StorageType", volume.ProviderType))
			} else {
				logger.Info("Failed to order volume ================>", zap.Object("StorageType", volume.ProviderType), zap.Object("Error", errr))
			}
			fmt.Printf("\n\n")
		} else if choiceN == 5 {
				fmt.Println("You selected choice to get snapshot details\n")
				fmt.Printf("Please enter Snapshot ID: ")
				_, er11 = fmt.Scanf("%s", &volumeID)
				snapdetails, errr := sess.SnapshotGet(volumeID)
				fmt.Printf("\n\n")
				if errr == nil {
					logger.Info("Successfully get snapshot details ================>", zap.Object("Snapshot ID", volumeID))
					logger.Info("Snapshot details ================>", zap.Object("SnapshotDetails", snapdetails))
				} else {
					logger.Info("Failed to get snapshot details ================>", zap.Object("Snapshot ID", volumeID), zap.Object("Error", errr))
				}
				fmt.Printf("\n\n")
		} else if choiceN == 6 {
				fmt.Println("You selected choice to order snapshot\n")
				volume := &provider.Volume{}
				fmt.Printf("Please enter volume ID to create the snapshot space: ")
				_, er11 = fmt.Scanf("%s", &volumeID)
				volume.VolumeID = volumeID
				var size int
				fmt.Printf("Please enter snapshot space size in GB: ")
				_, er11 = fmt.Scanf("%d", &size)
				volume.SnapshotSpace = &size
				er11 := sess.SnapshotOrder(*volume)
				if er11 == nil {
					logger.Info("Successfully ordered snapshot space ================>", zap.Object("Volume ID", volumeID))
				} else {
					logger.Info("failed to order snapshot space================>", zap.Object("Volume ID", volumeID), zap.Object("Error", er11))
				}
				fmt.Printf("\n\n")
		} else if choiceN == 7 {
				fmt.Println("You selected choice to Create volume from snapshot\n")
				var snapshotVol provider.Snapshot
				var tags map[string]string
				fmt.Printf("Please enter original volume ID to create the volume from snapshot: ")
				_, er11 = fmt.Scanf("%s", &volumeID)
				fmt.Printf("Please enter snapshot ID for creating volume:")
				_, er11 = fmt.Scanf("%s", &snapshotID)
				snapshotVol.SnapshotID = snapshotID
				snapshotVol.Volume.VolumeID = volumeID
				vol, errr := sess.VolumeCreateFromSnapshot(snapshotVol, tags)
				if errr == nil {
					logger.Info("Successfully Created volume from snapshot ================>", zap.Object("OriginalVolumeID", volumeID), zap.Object("SnapshotID", snapshotID))
					logger.Info("New volume from snapshot================>", zap.Object("New Volume->", vol))
				} else {
					logger.Info("Failed to create volume from snapshot ================>", zap.Object("OriginalVolumeID", volumeID), zap.Object("SnapshotID", snapshotID), zap.Object("Error", errr))
				}
				fmt.Printf("\n\n")
		} else if choiceN == 8 {
				fmt.Println("You selected choice to delete volume\n")
				volume := &provider.Volume{}
				fmt.Printf("Please enter volume ID for delete:")
				_, er11 = fmt.Scanf("%s", &volumeID)
				volume.VolumeID = volumeID
				er11 = sess.VolumeDelete(volume)
				if er11 == nil {
					logger.Info("Successfully deleted volume ================>", zap.Object("Volume ID", volumeID))
				} else {
					logger.Info("failed volume deletion================>", zap.Object("Volume ID", volumeID), zap.Object("Error", er11))
				}
				fmt.Printf("\n\n")
		} else if choiceN == 9 {
				fmt.Println("You selected choice to delete snapshot\n")
				snapshot := &provider.Snapshot{}
				fmt.Printf("Please enter snapshot ID for delete:")
				_, er11 = fmt.Scanf("%s", &snapshotID)
				snapshot.SnapshotID = snapshotID
				er11 = sess.SnapshotDelete(snapshot)
				if er11 == nil {
					logger.Info("Successfully deleted snapshot ================>", zap.Object("Snapshot ID", snapshotID))
				} else {
					logger.Info("failed snapshot deletion================>", zap.Object("Snapshot ID", snapshotID), zap.Object("Error", er11))
				}
				fmt.Printf("\n\n")
		} else if choiceN == 10 {
				fmt.Println("You selected choice to list all snapshot\n")
				list, _ := sess.SnapshotsList()
				logger.Info("All snapshots ================>", zap.Object("Snapshots", list))
				fmt.Printf("\n\n")
		} else if choiceN == 11 {
			fmt.Println("Get volume ID by using order ID\n")
			fmt.Printf("Please enter volume order ID to get volume ID:")
			_, er11 = fmt.Scanf("%s", &volumeID)
			_, error1 := sess.ListAllSnapshotsForVolume(volumeID)
			if error1 != nil {
				logger.Info("Failed to get volumeID", zap.Object("Error", error1))
			}
		} else {
				fmt.Println("No right choice")
				return
		}
	}
}
