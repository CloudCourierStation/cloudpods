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

package suggestion

import (
	"yunion.io/x/onecloud/cmd/climc/shell"
	modules "yunion.io/x/onecloud/pkg/mcclient/modules/suggestion"
	options "yunion.io/x/onecloud/pkg/mcclient/options/monitor"
)

func init() {

	cmd := shell.NewResourceCmd(modules.SuggestSysRuleManager)
	cmd.List(new(options.SuggestRuleListOptions))
	cmd.Show(new(options.SuggestRuleShowOptions))
	cmd.Perform("enable", new(options.SuggestRuleShowOptions))
	cmd.Perform("disable", new(options.SuggestRuleShowOptions))
	cmd.BatchDelete(new(options.CommonAlertDeleteOptions))
	cmd.Perform("config", new(options.SuggestRuleConfigOptions))

}
