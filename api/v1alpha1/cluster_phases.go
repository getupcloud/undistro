package v1alpha1

type ClusterPhase string

const (
	NewPhase          = ClusterPhase("")
	InitializedPhase  = ClusterPhase("Initialized")
	ProvisioningPhase = ClusterPhase("Provisioning")
	ProvisionedPhase  = ClusterPhase("Provisioned")
	FailedPhase       = ClusterPhase("Failed")
	DeletingPhase     = ClusterPhase("Deleting")
	DeletedPhase      = ClusterPhase("Deleted")
	UpgradingPhase    = ClusterPhase("Upgrading")
)
