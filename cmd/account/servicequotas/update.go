package servicequotas

import (
	"fmt"

	//"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicequotas"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	k8spkg "github.com/openshift/osd-utils-cli/pkg/k8s"
	awsprovider "github.com/openshift/osd-utils-cli/pkg/provider/aws"
)

// newCmdUpdate implements servicequotas describe
func newCmdUpdate(streams genericclioptions.IOStreams, flags *genericclioptions.ConfigFlags) *cobra.Command {
	ops := newUpdateOptions(streams, flags)
	updateCmd := &cobra.Command{
		Use:               "update",
		Short:             "Update AWS service-quotas",
		Args:              cobra.NoArgs,
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(ops.complete(cmd))
			cmdutil.CheckErr(ops.run())
		},
		Aliases: []string{"request-service-quota-increase", "update-quotas", "update-quota"},
	}

	ops.k8sclusterresourcefactory.AttachCobraCliFlags(updateCmd)

	updateCmd.Flags().StringVarP(&ops.queryServiceCode, "service-code", "", "ec2", "Query for ServiceCode")
	updateCmd.Flags().StringVarP(&ops.queryQuotaCode, "quota-code", "q", "L-1216C47A", "Query for QuotaCode")
	updateCmd.Flags().Float64VarP(&ops.desiredValue, "desired-value", "", -1, "Desired Value for Quota")

	updateCmd.Flags().BoolVarP(&ops.allRegions, "all-regions", "", false, "Loop through all supported regions")

	updateCmd.Flags().BoolVarP(&ops.verbose, "verbose", "v", false, "Verbose output")

	return updateCmd
}

// describeOptions defines the struct for running list account command
type updateOptions struct {
	k8sclusterresourcefactory k8spkg.ClusterResourceFactoryOptions

	queryServiceCode string
	queryQuotaCode   string
	desiredValue     float64

	verbose    bool
	allRegions bool

	genericclioptions.IOStreams
}

func newUpdateOptions(streams genericclioptions.IOStreams, flags *genericclioptions.ConfigFlags) *updateOptions {
	return &updateOptions{
		k8sclusterresourcefactory: k8spkg.ClusterResourceFactoryOptions{
			Flags: flags,
		},
		IOStreams: streams,
	}
}

func (o *updateOptions) complete(cmd *cobra.Command) error {
	k8svalid, err1 := o.k8sclusterresourcefactory.ValidateIdentifiers()
	if !k8svalid {
		if err1 != nil {
			return err1
		}
	}

	awsvalid, err2 := o.k8sclusterresourcefactory.Awscloudfactory.ValidateIdentifiers()
	if !awsvalid {
		if err2 != nil {
			return err2
		}
	}

	_, err3 := GetSupportedRegions(o.k8sclusterresourcefactory.Awscloudfactory.Region, o.allRegions)
	if err3 != nil {
		return err3
	}

	return nil
}

func (o *updateOptions) run() error {
	regions, error := GetSupportedRegions(o.k8sclusterresourcefactory.Awscloudfactory.Region, o.allRegions)
	if error != nil {
		return error
	}

	for _, region := range regions {
		o.runOnceByRegion(region)
	}

	return nil
}

func (o *updateOptions) runOnceByRegion(region string) error {
	// override region in factory class
	o.k8sclusterresourcefactory.Awscloudfactory.Region = region

	awsClient, err := o.k8sclusterresourcefactory.GetCloudProvider(o.verbose)
	if err != nil {
		return err
	}

	request := &servicequotas.RequestServiceQuotaIncreaseInput{
		ServiceCode:  &o.queryServiceCode,
		DesiredValue: &o.desiredValue, // *float64 `type:"double" required:"true"`
		QuotaCode:    &o.queryQuotaCode,
	}

	response, err := awsprovider.Client.RequestServiceQuotaIncrease(awsClient, request)
	if err != nil {
		return err
	}

	fmt.Println(response)

	return nil
}
