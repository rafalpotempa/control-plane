package deprovisioning

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/broker"

	"github.com/stretchr/testify/mock"

	hyperscalerMocks "github.com/kyma-project/control-plane/components/kyma-environment-broker/common/hyperscaler/automock"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal"
	provisionerAutomock "github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/provisioner/automock"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-project/control-plane/components/provisioner/pkg/gqlschema"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	fixOperationID            = "17f3ddba-1132-466d-a3c5-920f544d7ea6"
	fixInstanceID             = "9d75a545-2e1e-4786-abd8-a37b14e185b9"
	fixRuntimeID              = "ef4e3210-652c-453e-8015-bba1c1cd1e1c"
	fixGlobalAccountID        = "abf73c71-a653-4951-b9c2-a26d6c2cccbd"
	fixProvisionerOperationID = "e04de524-53b3-4890-b05a-296be393e4ba"
)

func TestInitialisationStep_Run(t *testing.T) {
	accountProviderMock := &hyperscalerMocks.AccountProvider{}
	accountProviderMock.On("MarkUnusedGardenerSecretAsDirty", mock.Anything, mock.AnythingOfType("string")).Return(nil)

	t.Run("Should mark operation as Succeeded when runtime deprovisioning was successful", func(t *testing.T) {
		// given
		log := logrus.New()
		memoryStorage := storage.NewMemoryStorage()

		operation := fixDeprovisioningOperation()
		err := memoryStorage.Operations().InsertDeprovisioningOperation(operation)
		assert.NoError(t, err)

		provisioningOperation := fixProvisioningOperation()
		err = memoryStorage.Operations().InsertProvisioningOperation(provisioningOperation)
		assert.NoError(t, err)

		instance := fixInstanceRuntimeStatus()
		err = memoryStorage.Instances().Insert(instance)
		assert.NoError(t, err)

		provisionerClient := &provisionerAutomock.Client{}
		provisionerClient.On("RuntimeOperationStatus", fixGlobalAccountID, fixProvisionerOperationID).Return(gqlschema.OperationStatus{
			ID:        ptr.String(fixProvisionerOperationID),
			Operation: "",
			State:     gqlschema.OperationStateSucceeded,
			Message:   nil,
			RuntimeID: nil,
		}, nil)

		step := NewInitialisationStep(memoryStorage.Operations(), memoryStorage.Instances(), provisionerClient, accountProviderMock)

		// when
		operation, repeat, err := step.Run(operation, log)

		// then
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(0), repeat)
		assert.Equal(t, domain.Succeeded, operation.State)
	})

	t.Run("Should delete instance when operation has succeeded due to runtime not existing", func(t *testing.T) {
		// given
		log := logrus.New()
		memoryStorage := storage.NewMemoryStorage()

		operation := fixDeprovisioningOperation()
		operation.ProvisionerOperationID = ""
		operation.State = domain.Succeeded
		err := memoryStorage.Operations().InsertDeprovisioningOperation(operation)
		assert.NoError(t, err)

		provisioningOperation := fixProvisioningOperation()
		err = memoryStorage.Operations().InsertProvisioningOperation(provisioningOperation)
		assert.NoError(t, err)

		instance := fixInstanceRuntimeStatus()
		instance.RuntimeID = ""
		err = memoryStorage.Instances().Insert(instance)
		assert.NoError(t, err)

		provisionerClient := &provisionerAutomock.Client{}

		step := NewInitialisationStep(memoryStorage.Operations(), memoryStorage.Instances(), provisionerClient, accountProviderMock)

		// when
		operation, repeat, err := step.Run(operation, log)

		// then
		assert.NoError(t, err)
		assert.Equal(t, time.Duration(0), repeat)
		assert.Equal(t, domain.Succeeded, operation.State)

		inst, err := memoryStorage.Instances().GetByID(operation.InstanceID)
		assert.Error(t, err)
		assert.Nil(t, inst)
	})

}

func fixDeprovisioningOperation() internal.DeprovisioningOperation {
	return internal.DeprovisioningOperation{
		Operation: internal.Operation{
			ID:                     fixOperationID,
			InstanceID:             fixInstanceID,
			ProvisionerOperationID: fixProvisionerOperationID,
			Description:            "",
			UpdatedAt:              time.Now(),
		},
	}
}

func fixProvisioningOperation() internal.ProvisioningOperation {
	planID := broker.AzurePlanID

	return internal.ProvisioningOperation{
		Operation: internal.Operation{
			ID:                     fixOperationID,
			InstanceID:             fixInstanceID,
			ProvisionerOperationID: fixProvisionerOperationID,
			Description:            "",
			UpdatedAt:              time.Now(),
		},
		ProvisioningParameters: fmt.Sprintf(`{"ers_context":{"globalaccount_id":"1"},"plan_id":"%s"}`, planID),
	}
}

func fixInstanceRuntimeStatus() internal.Instance {
	return internal.Instance{
		InstanceID:      fixInstanceID,
		RuntimeID:       fixRuntimeID,
		DashboardURL:    "",
		GlobalAccountID: fixGlobalAccountID,
		CreatedAt:       time.Time{},
		UpdatedAt:       time.Time{},
		DeletedAt:       time.Time{},
		ServicePlanID:   broker.AzurePlanID,
	}
}
