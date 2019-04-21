// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package guestdrivers

import (
	"context"
	"fmt"

	"yunion.io/x/pkg/utils"

	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudcommon/db/taskman"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/compute/models"
	"yunion.io/x/onecloud/pkg/httperrors"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/util/billing"
)

type SAliyunGuestDriver struct {
	SManagedVirtualizedGuestDriver
}

func init() {
	driver := SAliyunGuestDriver{}
	models.RegisterGuestDriver(&driver)
}

func (self *SAliyunGuestDriver) GetHypervisor() string {
	return api.HYPERVISOR_ALIYUN
}

func (self *SAliyunGuestDriver) GetDefaultSysDiskBackend() string {
	return api.STORAGE_CLOUD_EFFICIENCY
}

func (self *SAliyunGuestDriver) GetMinimalSysDiskSizeGb() int {
	return 20
}

func (self *SAliyunGuestDriver) GetStorageTypes() []string {
	return []string{
		api.STORAGE_CLOUD_EFFICIENCY,
		api.STORAGE_CLOUD_SSD,
		api.STORAGE_CLOUD_ESSD,
		api.STORAGE_PUBLIC_CLOUD,
		api.STORAGE_EPHEMERAL_SSD,
	}
}

func (self *SAliyunGuestDriver) ChooseHostStorage(host *models.SHost, backend string, storageIds []string) *models.SStorage {
	return self.chooseHostStorage(self, host, backend, storageIds)
}

func (self *SAliyunGuestDriver) GetDetachDiskStatus() ([]string, error) {
	return []string{api.VM_READY, api.VM_RUNNING}, nil
}

func (self *SAliyunGuestDriver) GetAttachDiskStatus() ([]string, error) {
	return []string{api.VM_READY, api.VM_RUNNING}, nil
}

func (self *SAliyunGuestDriver) GetRebuildRootStatus() ([]string, error) {
	return []string{api.VM_READY, api.VM_RUNNING}, nil
}

func (self *SAliyunGuestDriver) GetChangeConfigStatus() ([]string, error) {
	return []string{api.VM_READY, api.VM_RUNNING}, nil
}

func (self *SAliyunGuestDriver) GetDeployStatus() ([]string, error) {
	return []string{api.VM_READY, api.VM_RUNNING}, nil
}

func (self *SAliyunGuestDriver) ValidateResizeDisk(guest *models.SGuest, disk *models.SDisk, storage *models.SStorage) error {
	if !utils.IsInStringArray(guest.Status, []string{api.VM_READY, api.VM_RUNNING}) {
		return fmt.Errorf("Cannot resize disk when guest in status %s", guest.Status)
	}
	if disk.DiskType == api.DISK_TYPE_SYS {
		return fmt.Errorf("Cannot resize system disk")
	}
	if !utils.IsInStringArray(storage.StorageType, []string{api.STORAGE_PUBLIC_CLOUD, api.STORAGE_CLOUD_SSD, api.STORAGE_CLOUD_EFFICIENCY}) {
		return fmt.Errorf("Cannot resize %s disk", storage.StorageType)
	}
	return nil
}

func (self *SAliyunGuestDriver) RequestDetachDisk(ctx context.Context, guest *models.SGuest, task taskman.ITask) error {
	return guest.StartSyncTask(ctx, task.GetUserCred(), false, task.GetTaskId())
}

func (self *SAliyunGuestDriver) ValidateCreateData(ctx context.Context, userCred mcclient.TokenCredential, input *api.ServerCreateInput) (*api.ServerCreateInput, error) {
	input, err := self.SManagedVirtualizedGuestDriver.ValidateCreateData(ctx, userCred, input)
	if err != nil {
		return nil, err
	}
	if len(input.Networks) > 2 {
		return nil, httperrors.NewInputParameterError("cannot support more than 1 nic")
	}
	for i, disk := range input.Disks {
		if i == 0 && (disk.SizeMb < 20*1024 || disk.SizeMb > 500*1024) {
			return nil, httperrors.NewInputParameterError("The system disk size must be in the range of 20GB ~ 500Gb")
		}
		switch disk.Backend {
		case api.STORAGE_CLOUD_EFFICIENCY, api.STORAGE_CLOUD_SSD, api.STORAGE_CLOUD_ESSD:
			if disk.SizeMb < 20*1024 || disk.SizeMb > 32768*1024 {
				return nil, httperrors.NewInputParameterError("The %s disk size must be in the range of 20GB ~ 32768GB", disk.Backend)
			}
		case api.STORAGE_PUBLIC_CLOUD:
			if disk.SizeMb < 5*1024 || disk.SizeMb > 2000*1024 {
				return nil, httperrors.NewInputParameterError("The %s disk size must be in the range of 5GB ~ 2000GB", disk.Backend)
			}
		case api.STORAGE_EPHEMERAL_SSD:
			if disk.SizeMb < 5*1024 || disk.SizeMb > 800*1024 {
				return nil, httperrors.NewInputParameterError("The %s disk size must be in the range of 5GB ~ 800GB", disk.Backend)
			}
		}
	}
	return input, nil
}

func (self *SAliyunGuestDriver) GetGuestInitialStateAfterCreate() string {
	return api.VM_READY
}

func (self *SAliyunGuestDriver) GetGuestInitialStateAfterRebuild() string {
	return api.VM_READY
}

func (self *SAliyunGuestDriver) GetLinuxDefaultAccount(desc cloudprovider.SManagedVMCreateConfig) string {
	userName := "root"
	if desc.ImageType == "system" && desc.OsType == "Windows" {
		userName = "Administrator"
	}
	return userName
}

func (self *SAliyunGuestDriver) AllowReconfigGuest() bool {
	return true
}

func (self *SAliyunGuestDriver) IsSupportedBillingCycle(bc billing.SBillingCycle) bool {
	weeks := bc.GetWeeks()
	if weeks >= 1 && weeks <= 4 {
		return true
	}
	months := bc.GetMonths()
	if (months >= 1 && months <= 10) || (months == 12) || (months == 24) || (months == 36) || (months == 48) || (months == 60) {
		return true
	}
	return false
}
