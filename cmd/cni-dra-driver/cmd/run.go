/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/containerd/nri/pkg/stub"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cni-dra-driver/pkg/cni"
	"sigs.k8s.io/cni-dra-driver/pkg/discovery"
	"sigs.k8s.io/cni-dra-driver/pkg/driver"
	"sigs.k8s.io/cni-dra-driver/pkg/nri"
	"sigs.k8s.io/cni-dra-driver/pkg/status"
	"sigs.k8s.io/cni-dra-driver/pkg/store"
)

type runOptions struct {
	pluginName    string
	pluginIndex   string
	CNIPath       string
	CNICacheDir   string
	ChrootDir     string
	DRADriverName string
	NodeName      string
	verbosity     int
}

func newCmdRun() *cobra.Command {
	runOpts := &runOptions{}

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run the cni-dra-driver",
		Long:  `Run the cni-dra-driver`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runOpts.run(cmd.Context())
		},
	}

	cmd.Flags().StringVar(
		&runOpts.pluginName,
		"plugin-name",
		"cni-dra-driver",
		"Plugin name to register to NRI.",
	)

	cmd.Flags().StringVar(
		&runOpts.pluginIndex,
		"plugin-index",
		"",
		"plugin index to register to NRI.",
	)

	cmd.Flags().StringVar(
		&runOpts.CNIPath,
		"cni-path",
		"/opt/cni/bin",
		"CNI Path.",
	)

	cmd.Flags().StringVar(
		&runOpts.CNICacheDir,
		"cni-cache-dir",
		"/var/lib/cni/cni-dra-driver",
		"CNI Cache dir.",
	)

	cmd.Flags().StringVar(
		&runOpts.ChrootDir,
		"chroot-dir",
		"/hostroot",
		"ChrootDir.",
	)

	cmd.Flags().StringVar(
		&runOpts.DRADriverName,
		"dra-driver-name",
		"cni.dra.networking.x-k8s.io",
		"DRA Driver Name.",
	)

	cmd.Flags().StringVar(
		&runOpts.NodeName,
		"node-name",
		"",
		"Node Name.",
	)

	cmd.Flags().IntVar(
		&runOpts.verbosity,
		"verbosity",
		0,
		"Log Level.",
	)

	return cmd
}

func (ro *runOptions) run(ctx context.Context) error {
	opts := []stub.Option{
		stub.WithPluginName(ro.pluginName),
		stub.WithPluginIdx(ro.pluginIndex),
	}

	klog.InitFlags(nil)
	_ = flag.Set("v", fmt.Sprintf("%d", ro.verbosity))
	flag.Parse()

	clientCfg, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to InClusterConfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(clientCfg)
	if err != nil {
		return fmt.Errorf("failed to NewForConfig: %v", err)
	}

	memoryStore := store.NewMemory()

	cnish := status.Handler{
		ClientSet:  clientset,
		DriverName: ro.DRADriverName,
	}

	cniRuntime := cni.New(
		ro.DRADriverName,
		ro.ChrootDir,
		[]string{ro.CNIPath},
		ro.CNICacheDir,
	)

	draDriver, err := driver.Start(
		ctx,
		ro.DRADriverName,
		ro.NodeName,
		clientset,
		memoryStore,
		cniRuntime,
	)
	if err != nil {
		return fmt.Errorf("failed to dra.Start: %v", err)
	}
	defer draDriver.Stop()

	resourceDiscovery := &discovery.Resources{
		PublishResourcesFunc: draDriver.PublishResources,
		Interval:             10 * time.Second,
		NodeName:             ro.NodeName,
	}
	go resourceDiscovery.Run(ctx)

	p := &nri.Plugin{
		CNI:              cniRuntime,
		PodResourceStore: memoryStore,
		UpdateStatusFunc: cnish.UpdateStatus,
	}

	p.Stub, err = stub.New(p, opts...)
	if err != nil {
		return fmt.Errorf("failed to create plugin stub: %v", err)
	}

	err = p.Stub.Run(ctx)
	if err != nil {
		return fmt.Errorf("plugin exited with error: %v", err)
	}

	return nil
}
