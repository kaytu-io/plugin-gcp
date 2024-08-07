// Google Cloud Compute Instances

package gcp

import (
	computeApi "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"context"
	"errors"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"log"
)

type Compute struct {
	instancesClient   *computeApi.InstancesClient
	machineTypeClient *computeApi.MachineTypesClient
	computeService    *compute.Service
	GCP
}

func NewCompute(scopes []string) *Compute {
	return &Compute{
		GCP: NewGCP(scopes),
	}
}

func (c *Compute) InitializeClient(ctx context.Context) error {

	err := c.GCP.GetCredentials(ctx)
	if err != nil {
		return err
	}

	// log.Println(string(c.GCP.credentials.JSON))
	// log.Println(c.GCP.ProjectID)

	instancesClient, err := computeApi.NewInstancesRESTClient(
		ctx,
		option.WithCredentials(c.GCP.credentials),
	)
	if err != nil {
		return err
	}

	machineTypeClient, err := computeApi.NewMachineTypesRESTClient(
		ctx,
		option.WithCredentials(c.GCP.credentials),
	)
	if err != nil {
		return err
	}

	computeService, err := compute.NewService(
		ctx,
		option.WithCredentials(c.GCP.credentials),
	)
	if err != nil {
		return err
	}

	// log.Println(instancesClient)

	c.instancesClient = instancesClient
	c.machineTypeClient = machineTypeClient
	c.computeService = computeService

	return nil
}

func (c *Compute) CloseClient() error {
	err := c.instancesClient.Close()
	if err != nil {
		return err
	}
	err = c.machineTypeClient.Close()
	if err != nil {
		return err
	}
	return nil
}

func (c *Compute) ListAllInstances(ctx context.Context) error {

	req := &computepb.AggregatedListInstancesRequest{
		Project: c.ProjectID,
	}

	it := c.instancesClient.AggregatedList(ctx, req)

	log.Println("instances found: ")

	for {
		pair, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return err
		}

		instances := pair.Value.Instances

		if len(instances) > 0 {
			// log.Printf("%s\n", pair.Key)
			for _, instance := range instances {
				log.Printf("%s", instance.GetName())
			}
		}
	}
	return nil
}

func (c *Compute) GetAllInstances(ctx context.Context) ([]*computepb.Instance, error) {

	var allInstances []*computepb.Instance

	req := &computepb.AggregatedListInstancesRequest{
		Project: c.ProjectID,
	}

	it := c.instancesClient.AggregatedList(ctx, req)

	log.Println("instances found: ")

	for {
		pair, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		instances := pair.Value.Instances

		// allInstances = append(allInstances, *instances)

		if len(instances) > 0 {
			allInstances = append(allInstances, instances...)

			// log.Printf("%s\n", pair.Key) = append(allInstances, instances...)
			// for _, instance := range instances {
			// 	log.Printf("%s", instance.GetName())
			// 	allInstances = append(allInstances, *instance)
			// }
		}
	}
	return allInstances, nil
}

func (c *Compute) GetDiskDetails(ctx context.Context, zone, diskName string) (*compute.Disk, error) {
	disk, err := c.computeService.Disks.Get(c.ProjectID, zone, diskName).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	return disk, nil
}

func (c *Compute) GetMemory(ctx context.Context, instanceMachineType string, zone string) (*int32, error) {

	request := &computepb.GetMachineTypeRequest{
		Project:     c.ProjectID,
		MachineType: instanceMachineType,
		Zone:        zone,
	}

	machineType, err := c.machineTypeClient.Get(ctx, request)
	if err != nil {
		return nil, err
	}

	memory := machineType.GetMemoryMb()

	return &memory, nil

}
