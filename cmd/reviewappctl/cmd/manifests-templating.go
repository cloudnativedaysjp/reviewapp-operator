package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/resource"
	"sigs.k8s.io/yaml"

	dreamkastv1alpha1 "github.com/cloudnativedaysjp/reviewapp-operator/api/v1alpha1"
	"github.com/cloudnativedaysjp/reviewapp-operator/cmd/reviewappctl/pkg/errors"
	"github.com/cloudnativedaysjp/reviewapp-operator/cmd/reviewappctl/pkg/utils"
)

type manifestsTemplatingOptions struct {
	name        string
	namespace   string
	basefile    string
	output      string
	validate    bool
	isStable    bool
	isCandidate bool
}

var mto = &manifestsTemplatingOptions{}

var manifestsTemplatingCmd = &cobra.Command{
	Use: "manifests-templating",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !(len(args) > 0) {
			return fmt.Errorf("invalid arguments")
		}
		return runManifestsTemplating(cmd, args)
	},
}

func init() {
	manifestsTemplatingCmd.Flags().StringVarP(&mto.name, "name", "", "",
		"name of ManifestsTemplate")
	manifestsTemplatingCmd.Flags().StringVarP(&mto.namespace, "namespace", "", "default",
		"namespace of ManifestsTemplate")
	manifestsTemplatingCmd.Flags().StringVarP(&mto.basefile, "load", "f", "",
		"filename of manifests based on ManifestsTemplate")
	manifestsTemplatingCmd.Flags().StringVarP(&mto.output, "output", "o", "manifests_template.yaml",
		"output filename of manifest that ManifestsTemplate is written")
	manifestsTemplatingCmd.Flags().BoolVarP(&mto.validate, "validate", "", true,
		"TODO: description")
	manifestsTemplatingCmd.Flags().BoolVarP(&mto.isStable, "is-stable", "", false,
		"TODO: description")
	manifestsTemplatingCmd.Flags().BoolVarP(&mto.isCandidate, "is-candidate", "", false,
		"TODO: description")

	RootCmd.AddCommand(manifestsTemplatingCmd)
}

func runManifestsTemplating(cmd *cobra.Command, files []string) error {
	// validation
	if mto.name == "" && mto.basefile == "" {
		return fmt.Errorf("required either --name or --load option")
	}
	if !mto.isStable && !mto.isCandidate {
		return fmt.Errorf("required either --is-stable or --is-candidate options")
	}
	if len(files) < 1 {
		return fmt.Errorf("required one or more filenames")
	}

	// declear ManifestsTemplate
	var mt dreamkastv1alpha1.ManifestsTemplate
	if err := utils.ValidateFile(mto.basefile); err == nil {
		fmt.Println("load basefile...")
		b, err := ioutil.ReadFile(mto.basefile)
		if err != nil {
			return err
		}
		if err := yaml.Unmarshal(b, &mt); err != nil {
			return err
		}
	} else {
		fmt.Println("new struct of ManifestsTemplate...")
		mt = dreamkastv1alpha1.ManifestsTemplate{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ManifestsTemplate",
				APIVersion: "dreamkast.cloudnativedays.jp/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      mto.name,
				Namespace: mto.namespace,
			},
		}
	}
	if mt.Spec.StableData == nil {
		mt.Spec.StableData = make(map[string]string)
	}
	if mt.Spec.CandidateData == nil {
		mt.Spec.CandidateData = make(map[string]string)
	}

	// load manifests & construct ManifestsTemplate
	for _, file := range files {
		switch err := utils.ValidateFile(file).(type) {
		case errors.ErrorFileNotFound:
			return fmt.Errorf("%w\n", err)
		case errors.ErrorIsDirectory:
			fmt.Printf("%s: skip\n", err)
		case nil:
			// pass
		default:
			panic("unknown error occured")
		}

		// validate schema of manifest
		if _, err := builder.
			Unstructured().
			Schema(validator).
			ContinueOnError().
			FilenameParam(mto.validate, &resource.FilenameOptions{Filenames: []string{file}}).
			Flatten().
			Do().Infos(); err != nil {
			return err
		}

		// load manifest
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		if mto.isStable {
			mt.Spec.StableData[file] = string(b)
		} else if mto.isCandidate {
			mt.Spec.CandidateData[file] = string(b)
		}
	}

	// write to output manifest
	data, err := yaml.Marshal(mt)
	if err != nil {
		return err
	}
	ioutil.WriteFile(mto.output, data, 0644)
	fmt.Printf("output to %s\n", mto.output)
	return nil
}
