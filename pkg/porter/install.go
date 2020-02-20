package porter

import (
	"fmt"

	"get.porter.sh/porter/pkg/credentialsgenerator"
	"get.porter.sh/porter/pkg/manifest"
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/pkg/errors"
)

const generateCredCode = "generate new credential set"
const quitGenerateCode = "quit"

// InstallOptions that may be specified when installing a bundle.
// Porter handles defaulting any missing values.
type InstallOptions struct {
	BundleLifecycleOpts
}

// InstallBundle accepts a set of pre-validated InstallOptions and uses
// them to install a bundle.
func (p *Porter) InstallBundle(opts InstallOptions) error {
	err := p.prepullBundleByTag(&opts.BundleLifecycleOpts)
	if err != nil {
		return errors.Wrap(err, "unable to pull bundle before installation")
	}

	err = p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}

	err = p.ensureLocalBundleIsUpToDate(opts.bundleFileOptions)
	if err != nil {
		return err
	}

	if len(opts.CredentialIdentifiers) == 0 {
		var credSetNames []string
		credSets, err := p.Credentials.ReadAll()
		if err != nil {
			return errors.Wrap(err, "failed to read exisiting credential sets")
		}

		shouldGenerateCred := false
		var selectedOption string
		if len(credSets) > 0 {
			for _, credSet := range credSets {
				credSetNames = append(credSetNames, credSet.Name)
			}

			selectCredPrompt := &survey.Select{
				Message: "Choose a set of credentials to use while installing this bundle",
				Options: append(credSetNames, generateCredCode, quitGenerateCode),
				Default: credSetNames[0],
			}
			survey.AskOne(selectCredPrompt, &selectedOption, nil)

			switch selectedOption {
			case generateCredCode:
				shouldGenerateCred = true
			case quitGenerateCode:
				return nil
			default:
				opts.CredentialIdentifiers = append(opts.CredentialIdentifiers, selectedOption)
			}
		} else {
			shouldGenerateCredPrompt := &survey.Confirm{
				Message: "No credential identifier given. Generate one ?",
			}
			survey.AskOne(shouldGenerateCredPrompt, &shouldGenerateCred, nil)
			if !shouldGenerateCred {
				fmt.Fprintln(p.Out, "Credentials are mandatory to install this bundle")
				return nil
			}
		}

		if shouldGenerateCred {
			bundle, err := p.CNAB.LoadBundle(opts.CNABFile)
			if err != nil {
				return errors.Wrap(err, "failed to load bundle while generating credentials")
			}

			var credIdentifierName string
			inputCredNamePrompt := &survey.Input{
				Message: "Enter credential identifier name",
				Default: bundle.Name,
			}
			survey.AskOne(inputCredNamePrompt, &credIdentifierName, nil)

			genOpts := credentialsgenerator.GenerateOptions{
				Name:        credIdentifierName,
				Credentials: bundle.Credentials,
			}

			err = p.generateAndSaveCredentialSetForCNABFile(opts.CNABFile, genOpts)
			if err != nil {
				return errors.Wrap(err, "failed to generate credentials")
			}
			fmt.Fprintln(p.Out, "Credentials generated and saved successfully for future use")

			opts.CredentialIdentifiers = append(opts.CredentialIdentifiers, credIdentifierName)
		}

	}

	deperator := newDependencyExecutioner(p)
	err = deperator.Prepare(opts.BundleLifecycleOpts, p.CNAB.Install)
	if err != nil {
		return err
	}

	err = deperator.Execute(manifest.ActionInstall)
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "installing %s...\n", opts.Name)
	return p.CNAB.Install(opts.ToActionArgs(deperator))
}
