package cmd

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	validate    bool
	isStable    bool
	isCandidate bool
}

var mto = &manifestsTemplatingOptions{}

var manifestsTemplatingCmd = &cobra.Command{
	Use: "manifests-templating",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runManifestsTemplating(cmd, args)
	},
}

func init() {
	manifestsTemplatingCmd.Flags().StringVarP(&mto.name, "name", "", "",
		"name of ManifestsTemplate")
	manifestsTemplatingCmd.Flags().StringVarP(&mto.namespace, "namespace", "", "default",
		"namespace of ManifestsTemplate")
	manifestsTemplatingCmd.Flags().StringVarP(&mto.basefile, "load", "f", "manifests_template.yaml",
		"filename of manifests based on ManifestsTemplate")
	manifestsTemplatingCmd.Flags().BoolVarP(&mto.validate, "validate", "", true,
		"validate manifests if this flag is true")
	manifestsTemplatingCmd.Flags().BoolVarP(&mto.isStable, "is-stable", "", false,
		"using Stable template if this flag is true")
	manifestsTemplatingCmd.Flags().BoolVarP(&mto.isCandidate, "is-candidate", "", false,
		"using Candidate template if this flag is true")

	RootCmd.AddCommand(manifestsTemplatingCmd)
}

func runManifestsTemplating(cmd *cobra.Command, files []string) error {
	// validate
	if len(files) < 1 {
		return fmt.Errorf("required one or more filenames")
	}
	if !mto.isStable && !mto.isCandidate {
		return fmt.Errorf("required either --is-stable or --is-candidate options")
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
	}
	if mt.Name == "" {
		if mto.name == "" {
			return fmt.Errorf("required --name option")
		}
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
		mt.Spec.StableData = make(map[string]unstructured.Unstructured)
	}
	if mt.Spec.CandidateData == nil {
		mt.Spec.CandidateData = make(map[string]unstructured.Unstructured)
	}

	// load manifests & construct ManifestsTemplate
	for _, file := range files {
		switch err := utils.ValidateFile(file).(type) {
		case *errors.ErrorFileNotFound:
			return fmt.Errorf("%w\n", err)
		case *errors.ErrorIsDirectory:
			fmt.Printf("%s: skip\n", err)
		case nil:
			// pass
		default:
			panic("unknown error occured")
		}

		// validate schema of manifest
		if mto.validate {
			if _, err := builder.
				Unstructured().
				Schema(validator).
				ContinueOnError().
				FilenameParam(false, &resource.FilenameOptions{Filenames: []string{file}}).
				Flatten().
				Do().Infos(); err != nil {
				return err
			}
		}

		// load manifest
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		var data unstructured.Unstructured
		if err := yaml.Unmarshal(b, &data); err != nil {
			return err
		}
		filename := filepath.Base(file)
		if mto.isStable {
			mt.Spec.StableData[filename] = data
		} else if mto.isCandidate {
			mt.Spec.CandidateData[filename] = data
		}
	}

	// write to output manifest
	data, err := yaml.Marshal(mt)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(mto.basefile, data, 0644); err != nil {
		return err
	}
	fmt.Printf("output to %s\n", mto.basefile)
	return nil
}
