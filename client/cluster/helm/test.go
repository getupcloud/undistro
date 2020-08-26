package helm

import (
	"helm.sh/helm/v3/pkg/action"
)

func (h *HelmV3) Uninstall(releaseName string, opts UninstallOptions) error {
	cfg, err := newActionConfig(h.kubeConfig, h.infoLogFunc(opts.Namespace, releaseName), opts.Namespace, "")
	if err != nil {
		return err
	}

	uninstall := action.NewUninstall(cfg)
	opts.configure(uninstall)

	_, err = uninstall.Run(releaseName)
	return err
}

func (opts UninstallOptions) configure(action *action.Uninstall) {
	action.DisableHooks = opts.DisableHooks
	action.DryRun = opts.DryRun
	action.KeepHistory = opts.KeepHistory
	action.Timeout = opts.Timeout
}

func (h *HelmV3) Test(releaseName string, opts TestOptions) error {
	cfg, err := newActionConfig(h.kubeConfig, h.infoLogFunc(opts.Namespace, releaseName), opts.Namespace, "")
	if err != nil {
		return err
	}

	test := action.NewReleaseTesting(cfg)
	opts.configure(test)

	if _, err := test.Run(releaseName); err != nil {
		return err
	}

	return nil
}

func (opts TestOptions) configure(action *action.ReleaseTesting) {
	action.Timeout = opts.Timeout
}
