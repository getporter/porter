package arm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	resourcesSDK "github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources" // nolint: lll
	"github.com/Azure/go-autorest/autorest"

	porterCtx "github.com/deislabs/porter/pkg/context"
)

type deploymentStatus string

const (
	deploymentStatusNotFound  deploymentStatus = "NOT_FOUND"
	deploymentStatusRunning   deploymentStatus = "RUNNING"
	deploymentStatusSucceeded deploymentStatus = "SUCCEEDED"
	deploymentStatusFailed    deploymentStatus = "FAILED"
	deploymentStatusUnknown   deploymentStatus = "UNKNOWN"
)

// Deployer is an interface to be implemented by any component capable of
// deploying resource to Azure using an ARM template
type Deployer interface {
	FindTemplate(kind, template string) ([]byte, error)
	Deploy(
		deploymentName string,
		resourceGroupName string,
		location string,
		template []byte,
		armParams map[string]interface{},
		tags map[string]string,
	) (map[string]interface{}, error)
	Update(
		deploymentName string,
		resourceGroupName string,
		location string,
		template []byte,
		armParams map[string]interface{},
		tags map[string]string,
	) (map[string]interface{}, error)
	Delete(deploymentName string, resourceGroupName string) error
}

// deployer is an ARM-based implementation of the Deployer interface
type deployer struct {
	context           *porterCtx.Context
	groupsClient      resourcesSDK.GroupsClient
	deploymentsClient resourcesSDK.DeploymentsClient
}

// NewDeployer returns a new ARM-based implementation of the Deployer interface
func NewDeployer(
	context *porterCtx.Context,
	groupsClient resourcesSDK.GroupsClient,
	deploymentsClient resourcesSDK.DeploymentsClient,
) Deployer {
	return &deployer{
		context:           context,
		groupsClient:      groupsClient,
		deploymentsClient: deploymentsClient,
	}
}

// Deploy idempotently handles ARM deployments. To do this, it checks for the
// existence and status of a deployment before choosing to create a new one,
// poll until success or failure, or return an error.
func (d *deployer) Deploy(
	deploymentName string,
	resourceGroupName string,
	location string,
	template []byte,
	armParams map[string]interface{},
	tags map[string]string,
) (map[string]interface{}, error) {

	// Get the deployment and its current status
	deployment, ds, err := d.getDeploymentAndStatus(
		deploymentName,
		resourceGroupName,
	)
	if err != nil {
		return nil, fmt.Errorf(
			`error deploying "%s" in resource group "%s": error getting `+
				`deployment: %s`,
			deploymentName,
			resourceGroupName,
			err,
		)
	}

	// Handle according to status...
	switch ds {
	case deploymentStatusNotFound:
		// The deployment wasn't found, which means we are free to proceed with
		// initiating a new deployment
		if deployment, err = d.doDeployment(
			deploymentName,
			resourceGroupName,
			location,
			template,
			armParams,
			tags,
		); err != nil {
			return nil, fmt.Errorf(
				`error deploying "%s" in resource group "%s": %s`,
				deploymentName,
				resourceGroupName,
				err,
			)
		}
	case deploymentStatusRunning:
		// The deployment exists and is currently running, which means we'll poll
		// until it completes. The return at the end of the function will return the
		// deployment's outputs.
		if deployment, err = d.pollUntilComplete(
			deploymentName,
			resourceGroupName,
		); err != nil {
			return nil, fmt.Errorf(
				`error deploying "%s" in resource group "%s": %s`,
				deploymentName,
				resourceGroupName,
				err,
			)
		}
	case deploymentStatusSucceeded:
		// The deployment exists and has succeeded already. There's nothing to do.
		// The return at the end of the function will return the deployment's
		// outputs.
	case deploymentStatusFailed:
		// The deployment exists and has failed already.
		return nil, fmt.Errorf(
			`error deploying "%s" in resource group "%s": deployment is in failed `+
				`state`,
			deploymentName,
			resourceGroupName,
		)
	case deploymentStatusUnknown:
		fallthrough
	default:
		// Unrecognized state
		return nil, fmt.Errorf(
			`error deploying "%s" in resource group "%s": deployment is in an `+
				`unrecognized state`,
			deploymentName,
			resourceGroupName,
		)
	}

	return getOutputs(deployment)
}

// Update idempotently handles ARM deployments. To do this, it checks for the
// existence and status of a deployment before choosing to update one,
// poll until success or failure, or return an error.
func (d *deployer) Update(
	deploymentName string,
	resourceGroupName string,
	location string,
	template []byte,
	armParams map[string]interface{},
	tags map[string]string,
) (map[string]interface{}, error) {
	// Get the deployment's current status
	_, ds, err := d.getDeploymentAndStatus(
		deploymentName,
		resourceGroupName,
	)
	if err != nil {
		return nil, fmt.Errorf(
			`error deploying "%s" in resource group "%s": error getting `+
				`deployment: %s`,
			deploymentName,
			resourceGroupName,
			err,
		)
	}

	// Handle according to status...
	switch ds {
	case deploymentStatusNotFound:
		// Update operations should be working against an existing deployment.
		// If we get here, that is a bad thing so we should error.

		return nil, fmt.Errorf(
			`error updating "%s" in resource group "%s": %s`,
			deploymentName,
			resourceGroupName,
			err,
		)
	case deploymentStatusRunning:
		// The deployment exists and is currently running, which means we'll poll
		// until it completes. The return at the end of the function will return the
		// deployment's outputs.

		deployment, err := d.pollUntilComplete(
			deploymentName,
			resourceGroupName,
		)
		if err != nil {
			return nil, fmt.Errorf(
				`error deploying "%s" in resource group "%s": %s`,
				deploymentName,
				resourceGroupName,
				err,
			)
		}
		return getOutputs(deployment)

	case deploymentStatusSucceeded:

		// doDeployment will call deploymentsClient.CreateOrUpdate
		// and update an existing deployment.
		deployment, err := d.doDeployment(
			deploymentName,
			resourceGroupName,
			location,
			template,
			armParams,
			tags,
		)
		if err != nil {
			return nil, fmt.Errorf(
				`error deploying "%s" in resource group "%s": %s`,
				deploymentName,
				resourceGroupName,
				err,
			)
		}
		return getOutputs(deployment)
	case deploymentStatusFailed:
		// The deployment exists and has failed already.
		return nil, fmt.Errorf(
			`error deploying "%s" in resource group "%s": deployment is in failed `+
				`state`,
			deploymentName,
			resourceGroupName,
		)
	case deploymentStatusUnknown:
		fallthrough
	default:
		// Unrecognized state
		return nil, fmt.Errorf(
			`error deploying "%s" in resource group "%s": deployment is in an `+
				`unrecognized state`,
			deploymentName,
			resourceGroupName,
		)
	}
}

func (d *deployer) Delete(
	deploymentName string,
	resourceGroupName string,
) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err := d.deploymentsClient.Delete(
		ctx,
		resourceGroupName,
		deploymentName,
	)
	if err != nil {
		return fmt.Errorf(
			`error deleting deployment "%s" from resource group "%s": %s`,
			deploymentName,
			resourceGroupName,
			err,
		)
	}
	if err := result.WaitForCompletion(
		ctx,
		d.deploymentsClient.Client,
	); err != nil {
		return fmt.Errorf(
			`error deleting deployment "%s" from resource group "%s": %s`,
			deploymentName,
			resourceGroupName,
			err,
		)
	}
	return nil
}

// getDeploymentAndStatus attempts to retrieve and return a deployment. Whether
// it's found or not, a status is returned. (It's not enough to just return the
// deployment and let the caller check status itself, because in the case a
// given deployment doesn't exist, there isn't one to return. Returning a
// separate status indicator resolves that problem.)
func (d *deployer) getDeploymentAndStatus(
	deploymentName string,
	resourceGroupName string,
) (*resourcesSDK.DeploymentExtended, deploymentStatus, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	deployment, err := d.deploymentsClient.Get(
		ctx,
		resourceGroupName,
		deploymentName,
	)
	if err != nil {
		detailedErr, ok := err.(autorest.DetailedError)
		if !ok || detailedErr.StatusCode != http.StatusNotFound {
			return nil, "", err
		}
		return nil, deploymentStatusNotFound, nil
	}
	switch *deployment.Properties.ProvisioningState {
	case "Running":
		return &deployment, deploymentStatusRunning, nil
	case "Succeeded":
		return &deployment, deploymentStatusSucceeded, nil
	case "Failed":
		return &deployment, deploymentStatusFailed, nil
	default:
		return &deployment, deploymentStatusUnknown, nil
	}
}

func (d *deployer) doDeployment(
	deploymentName string,
	resourceGroupName string,
	location string,
	armTemplate []byte,
	armParams map[string]interface{},
	tags map[string]string,
) (*resourcesSDK.DeploymentExtended, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	res, err := d.groupsClient.CheckExistence(ctx, resourceGroupName)
	if err != nil {
		return nil, fmt.Errorf(
			"error checking existence of resource group: %s",
			err,
		)
	}
	if res.StatusCode == http.StatusNotFound {
		if _, err = d.groupsClient.CreateOrUpdate(
			ctx,
			resourceGroupName,
			resourcesSDK.Group{
				Name:     &resourceGroupName,
				Location: &location,
			},
		); err != nil {
			return nil, fmt.Errorf(
				"error creating resource group: %s",
				err,
			)
		}
	}

	// Unmarshal the template into a map
	var armTemplateMap map[string]interface{}
	err = json.Unmarshal(armTemplate, &armTemplateMap)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling ARM template: %s", err)
	}

	// Deal with the possibility that tags == nil
	if tags == nil {
		tags = make(map[string]string)
	}

	// Augment the provided tags with heritage information
	tags["heritage"] = "porter-azure-mixin"

	// Deal with the possiiblity that params == nil
	if armParams == nil {
		armParams = make(map[string]interface{})
	}

	// Augment the params with tags
	armParams["tags"] = tags

	// Convert a simple map[string]interface{} to the more complex
	// map[string]map[string]interface{} required by the deployments client
	armParamsMap := map[string]interface{}{}
	for key, val := range armParams {
		armParamsMap[key] = map[string]interface{}{
			"value": val,
		}
	}
	// Deploy the template
	fmt.Printf("Starting ARM deployment")
	result, err := d.deploymentsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		deploymentName,
		resourcesSDK.Deployment{
			Properties: &resourcesSDK.DeploymentProperties{
				Template:   &armTemplateMap,
				Parameters: &armParamsMap,
				Mode:       resourcesSDK.Incremental,
			},
		},
	)
	fmt.Printf("Started ARM deployment")
	if err != nil {
		return nil, fmt.Errorf("error submitting ARM template: %s", err)
	}

	if err = result.WaitForCompletion(
		ctx,
		d.deploymentsClient.Client,
	); err != nil {
		return nil,
			fmt.Errorf("error while waiting for deployment to complete: %s", err)
	}

	// Deployment object found via the result doesn't include properties, so we
	// need to make a separate call to retrieve the deployment
	deployment, err := d.deploymentsClient.Get(
		ctx,
		resourceGroupName,
		deploymentName,
	)
	if err != nil {
		fmt.Printf("%s", err)
	}

	return &deployment, nil
}

// pollUntilComplete polls the status of a deployment periodically until the
// deployment succeeds or fails, polling fails, or a timeout is reached
func (d *deployer) pollUntilComplete(
	deploymentName string,
	resourceGroupName string,
) (*resourcesSDK.DeploymentExtended, error) {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	timer := time.NewTimer(time.Minute * 30)
	defer timer.Stop()
	var deployment *resourcesSDK.DeploymentExtended
	var ds deploymentStatus
	var err error
	for {
		select {
		case <-ticker.C:
			if deployment, ds, err = d.getDeploymentAndStatus(
				deploymentName,
				resourceGroupName,
			); err != nil {
				return nil, err
			}
			switch ds {
			case deploymentStatusNotFound:
				// This is an error. We'd only be polling for status on a deployment
				// that exists. If it no longer exists, something is very wrong.
				return nil, errors.New(
					"error polling deployment status; deployment should exist, but " +
						"does not",
				)
			case deploymentStatusRunning:
				// Do nothing == continue the loop
			case deploymentStatusSucceeded:
				// We're done
				return deployment, nil
			case deploymentStatusFailed:
				// The deployment has failed
				return nil, errors.New("deployment has failed")
			case deploymentStatusUnknown:
				fallthrough
			default:
				// The deployment has entered an unknown state
				return nil, errors.New("deployment is in an unrecognized state")
			}
		case <-timer.C:
			// We've reached a timeout
			return nil, errors.New("timed out waiting for deployment to complete")
		}
	}
}

func getOutputs(
	deployment *resourcesSDK.DeploymentExtended,
) (map[string]interface{}, error) {
	outputs, ok := deployment.Properties.Outputs.(map[string]interface{})
	if !ok {
		return nil, errors.New("error decoding deployment outputs")
	}
	retOutputs := map[string]interface{}{}
	for k, v := range outputs {
		output, ok := v.(map[string]interface{})
		if !ok {
			return nil, errors.New("error decoding deployment outputs")
		}
		retOutputs[k] = output["value"]
	}
	return retOutputs, nil
}
