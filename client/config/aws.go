package config

import (
	"bytes"
	"encoding/base64"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/session"
	cfn "github.com/aws/aws-sdk-go/service/cloudformation"
	undistrov1 "github.com/getupcloud/undistro/api/v1alpha1"
	"sigs.k8s.io/cluster-api-provider-aws/cmd/clusterawsadm/cloudformation/bootstrap"
	cloudformation "sigs.k8s.io/cluster-api-provider-aws/cmd/clusterawsadm/cloudformation/service"
)

const (
	defaultAWSRegion              = "us-east-1"
	awsSshKeyNameKey              = "AWS_SSH_KEY_NAME"
	awsControlPlaneMachineTypeKey = "AWS_CONTROL_PLANE_MACHINE_TYPE"
	awsWorkerMachineTypeKey       = "AWS_NODE_MACHINE_TYPE"
	awsRegionKey                  = "AWS_REGION"
	awsCredentialsKey             = "AWS_B64ENCODED_CREDENTIALS"

	awsCredentialsTemplate = `[default]
	aws_access_key_id = {{ .AccessKeyID }}
	aws_secret_access_key = {{ .SecretAccessKey }}
	region = {{ .Region }}
	{{if .SessionToken }}
	aws_session_token = {{ .SessionToken }}
	{{end}}
	`
)

type awsCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Region          string
}

func (c awsCredentials) renderAWSDefaultProfile() (string, error) {
	tmpl, err := template.New("AWS Credentials").Parse(awsCredentialsTemplate)
	if err != nil {
		return "", err
	}
	var credsFileStr bytes.Buffer
	err = tmpl.Execute(&credsFileStr, c)
	if err != nil {
		return "", err
	}
	return credsFileStr.String(), nil
}

func (c awsCredentials) setBase64EncodedAWSDefaultProfile(v VariablesClient) error {
	profile, err := c.renderAWSDefaultProfile()
	if err != nil {
		return err
	}
	b64 := base64.StdEncoding.EncodeToString([]byte(profile))
	v.Set(awsCredentialsKey, b64)
	return nil
}

func awsPreConfig(cl *undistrov1.Cluster, v VariablesClient) error {
	v.Set(awsSshKeyNameKey, cl.Spec.InfrastructureProvider.SSHKey)
	v.Set(awsControlPlaneMachineTypeKey, cl.Spec.ControlPlaneNode.MachineType)
	v.Set(awsWorkerMachineTypeKey, cl.Spec.WorkerNode.MachineType)
	return nil
}

func newAWSCreds(v VariablesClient) (*awsCredentials, error) {
	creds := awsCredentials{}
	region, err := v.Get(awsRegionKey)
	if region == "" || err != nil {
		region = defaultAWSRegion
		v.Set(awsRegionKey, region)
	}
	creds.Region = region
	conf := aws.NewConfig()
	conf.CredentialsChainVerboseErrors = aws.Bool(true)
	chain := defaults.CredChain(conf, defaults.Handlers())
	chainCreds, err := chain.Get()
	if err != nil {
		return nil, err
	}
	creds.Region = region
	creds.AccessKeyID = chainCreds.AccessKeyID
	creds.SecretAccessKey = chainCreds.SecretAccessKey
	creds.SessionToken = chainCreds.SessionToken
	return &creds, nil
}

func awsInit(c Client, firstRun bool) error {
	v := c.Variables()
	if firstRun {
		creds, err := newAWSCreds(v)
		if err != nil {
			return err
		}
		err = creds.setBase64EncodedAWSDefaultProfile(v)
		if err != nil {
			return err
		}
	}
	t := bootstrap.NewTemplate()
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return err
	}
	cfnSvc := cloudformation.NewService(cfn.New(sess))
	return cfnSvc.ReconcileBootstrapStack(t.Spec.StackName, *t.RenderCloudFormation())
}
