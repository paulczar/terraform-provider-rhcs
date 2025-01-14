package exec

import (
	"context"
	"fmt"

	CON "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/constants"
	h "github.com/terraform-redhat/terraform-provider-rhcs/tests/utils/helper"
)

type MachinePoolArgs struct {
	Cluster            string              `json:"cluster,omitempty"`
	OCMENV             string              `json:"ocm_environment,omitempty"`
	Name               string              `json:"name,omitempty"`
	Token              string              `json:"token,omitempty"`
	URL                string              `json:"url,omitempty"`
	MachineType        string              `json:"machine_type,omitempty"`
	Replicas           int                 `json:"replicas,omitempty"`
	AutoscalingEnabled bool                `json:"autoscaling_enabled,omitempty"`
	UseSpotInstances   bool                `json:"use_spot_instances,omitempty"`
	MaxReplicas        int                 `json:"max_replicas,omitempty"`
	MinReplicas        int                 `json:"min_replicas,omitempty"`
	MaxSpotPrice       float64             `json:"max_spot_price,omitempty"`
	Labels             map[string]string   `json:"labels,omitempty"`
	Taints             []map[string]string `json:"taints,omitempty"`
	ID                 string              `json:"id,omitempty"`
	AvailabilityZone   string              `json:"availability_zone,omitempty"`
	SubnetID           string              `json:"subnet_id,omitempty"`
	MultiAZ            bool                `json:"multi_availability_zone,omitempty"`
}
type MachinePoolService struct {
	CreationArgs *MachinePoolArgs
	ManifestDir  string
	Context      context.Context
}

type MachinePoolOutput struct {
	ID                 string            `json:"machine_pool_id,omitempty"`
	Name               string            `json:"name,omitempty"`
	ClusterID          string            `json:"cluster_id,omitempty"`
	Replicas           int               `json:"replicas,omitempty"`
	MachineType        string            `json:"machine_type,omitempty"`
	AutoscalingEnabled bool              `json:"autoscaling_enabled,omitempty"`
	Labels             map[string]string `json:"labels,omitempty"`
}

func (mp *MachinePoolService) Init(manifestDirs ...string) error {
	mp.ManifestDir = CON.AWSVPCDir
	if len(manifestDirs) != 0 {
		mp.ManifestDir = manifestDirs[0]
	}
	ctx := context.TODO()
	mp.Context = ctx
	err := runTerraformInit(ctx, mp.ManifestDir)
	if err != nil {
		return err
	}
	return nil

}

func (mp *MachinePoolService) Create(createArgs *MachinePoolArgs, extraArgs ...string) error {
	createArgs.URL = CON.GateWayURL
	mp.CreationArgs = createArgs
	args := combineStructArgs(createArgs, extraArgs...)
	_, err := runTerraformApplyWithArgs(mp.Context, mp.ManifestDir, args)
	if err != nil {
		return err
	}
	return nil
}

func (mp *MachinePoolService) Output() (MachinePoolOutput, error) {
	mpDir := CON.MachinePoolDir
	if mp.ManifestDir != "" {
		mpDir = mp.ManifestDir
	}
	var output MachinePoolOutput
	out, err := runTerraformOutput(context.TODO(), mpDir)
	if err != nil {
		return output, err
	}
	if err != nil {
		return output, err
	}
	replicas := h.DigInt(out["replicas"], "value")
	machine_type := h.DigString(out["machine_type"], "value")
	name := h.DigString(out["name"], "value")
	autoscaling_enabled := h.DigBool(out["autoscaling_enabled"])
	output = MachinePoolOutput{
		Replicas:           replicas,
		MachineType:        machine_type,
		Name:               name,
		AutoscalingEnabled: autoscaling_enabled,
	}
	return output, nil
}

func (mp *MachinePoolService) Destroy(createArgs ...*MachinePoolArgs) error {
	if mp.CreationArgs == nil && len(createArgs) == 0 {
		return fmt.Errorf("got unset destroy args, set it in object or pass as a parameter")
	}
	destroyArgs := mp.CreationArgs
	if len(createArgs) != 0 {
		destroyArgs = createArgs[0]
	}
	destroyArgs.URL = CON.GateWayURL
	args := combineStructArgs(destroyArgs)
	err := runTerraformDestroyWithArgs(mp.Context, mp.ManifestDir, args)

	return err
}

func NewMachinePoolService(manifestDir ...string) *MachinePoolService {
	mp := &MachinePoolService{}
	mp.Init(manifestDir...)
	return mp
}
