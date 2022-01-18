// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package utils

import (
	"testing"
)

func TestParseVMSSResourceID(t *testing.T) {
	type test struct {
		description    string
		id             string
		expectedResult VirtualMachineScaleSetVM
		expectedError  bool
	}

	table := []test{
		// KO tests
		{
			description:    "From empty id",
			id:             "",
			expectedResult: VirtualMachineScaleSetVM{},
			expectedError:  true,
		},
		{
			description:    "Unexpected format",
			id:             "subscriptionsmysubid",
			expectedResult: VirtualMachineScaleSetVM{},
			expectedError:  true,
		},
		{
			description:    "Unexpected separator",
			id:             "subscriptions-mysubid",
			expectedResult: VirtualMachineScaleSetVM{},
			expectedError:  true,
		},
		{
			description: "Incomplete format",
			id:          "/subscriptions/mysubid",
			expectedResult: VirtualMachineScaleSetVM{
				SubscriptionID: "mysubid",
			},
			expectedError: true,
		},
		// OK test
		{
			description: "Correct format",
			id:          "/subscriptions/mysubid/resourcegroups/myrd/providers/myprovider/virtualmachinescalesets/myvmss/virtualmachines/myinsid",
			expectedResult: VirtualMachineScaleSetVM{
				SubscriptionID:    "mysubid",
				NodeResourceGroup: "myrd",
				VMScaleSet:        "myvmss",
				InstanceID:        "myinsid",
			},
			expectedError: false,
		},
	}

	for _, entry := range table {
		result := VirtualMachineScaleSetVM{}
		err := ParseVMSSResourceID(entry.id, &result)
		errorOcurred := err != nil
		if errorOcurred != entry.expectedError || entry.expectedResult != result {
			t.Fatalf("Failed test %q: result %+v (error %t - %s) vs expected %+v (error %t)",
				entry.description, result, errorOcurred, err, entry.expectedResult, entry.expectedError)
		}
	}
}
