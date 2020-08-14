package v1alpha1

type ClusterPhase string

const (
	NewPhase          = ClusterPhase("")
	InitializedPhase  = ClusterPhase("initialized")
	ProvisioningPhase = ClusterPhase("provisioning")
	ProvisionedPhase  = ClusterPhase("provisioned")
	ReadyPhase        = ClusterPhase("ready")
)
