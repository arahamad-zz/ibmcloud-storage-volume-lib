/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package provider

import (
	"time"
)

type VolumeProvider string

type VolumeProviderType string

type VolumeType string

type SnapshotTags map[string]string

// Volume
type Volume struct {
	// ID of the storage volume, for which we can track the volume
	VolumeID string `json:"volumeID,omitempty"` // order id should be there in the pv object as k10 looks for that in pv object

	// volume provider name
	Provider VolumeProvider `json:"provider"`

	// volume type block or file
	VolumeType VolumeType `json:"volumeType"`

	// Volume provider type i.e  Endurance or Performance or any other name
	ProviderType VolumeProviderType `json:"providerType,omitempty"`

	// The Capacity of the volume, in GiB
	Capacity *int `json:"capacity"`

	// The size of the snapshot space, in GiB
	SnapshotSpace *int `json:"snapshotSpace,omitempty"`

	// Volume IOPS for performance storage type only
	Iops *string `json:"iops"`

	// for endurance storage type only
	Tier *string `json:"tier"`

	// region of the volume
	Region string `json:"region,omitempty"`

	// Availability zone/datacenter/location of the storage volume
	Az string `json:"az,omitempty"`

	// billing type monthly or hourly
	BillingType string `json:"billingType,omitempty"`

	// Time stamp when volume creation was initiated
	CreationTime time.Time `json:"creationTime`

	// storage_as_a_service|enterprise|performance     default from SL is storage_as_a_service
	ServiceOffering *string `json:"serviceOffering`

	// notes field as a map for all note fileds
	// will keep   {"plugin":"ibm-file-plugin-56f7bd4db6-wx4pd","region":"us-south","cluster":"3a3fd80459014aca84f8a7e58e7a3ded","type":"Endurance","pvc":"one30","pv":"pvc-c7b4d6bd-63c5-11e8-811c-3a16fc403383","storgeclass":"ibmc-file-billing","reclaim":"Delete"}
	VolumeNotes map[string]string
}

// Snapshot
type Snapshot struct {
	Volume

	// a unique Snapshot ID which created by the provider
	SnapshotID string `json:"snapshotID,omitempty"`

	// The size of the snapshot, in GiB
	SnapshotSize *int `json:"snapshotSize,omitempty"`

	// Source volume details
	//SnapshotedVolume *Volume `json:"SnapshotedVolume"`

	// Time stamp when snapshot creation was initiated
	SnapshotCreationTime time.Time `json:"snapCreationTime,omitempty"`

	// tags for the snapshot
	SnapshotTags SnapshotTags `json:"tags"`
}
