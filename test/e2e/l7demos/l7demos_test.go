// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package l7demos

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/cilium/cilium/test/e2e/helpers"
)

var (
	testenv env.Environment

	namespace string = envconf.RandomName("l7demos-ns", 16)
)

func TestMain(m *testing.M) {
	cfg, err := envconf.NewFromFlags()
	if err != nil {
		klog.Fatalf("failed to build envconf from flags: %s", err)
	}
	testenv = env.NewWithConfig(cfg)
	ciliumHelmOpts := map[string]string{
		// empty opts
	}
	testenv.Setup(
		helpers.MaybeCreateTempKindCluster(),
		envfuncs.CreateNamespace(namespace),
		helpers.InstallCiliumWithOpts(ciliumHelmOpts),
	)
	testenv.Finish(
		helpers.UninstallCilium(),
		envfuncs.DeleteNamespace(namespace),
		helpers.MaybeDeleteTempKindCluster(),
	)
	os.Exit(testenv.Run(m))
}

func TestStarWarsDemo(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// TODO(tk): replace by Go struct definitions?
	manifests := os.DirFS(filepath.Join(pwd, "..", "..", "k8s", "manifests", "star-wars-demo"))
	pattern := "*.yaml"

	starWarsDemoFeature := features.New("l7 demos").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// load manifests
			r, err := resources.New(cfg.Client().RESTConfig())
			if err != nil {
				t.Fatal(err)
			}
			if err := decoder.DecodeEachFile(ctx, manifests, pattern,
				decoder.CreateHandler(r),           // try to CREATE objects after decoding
				decoder.MutateNamespace(namespace), // inject a namespace into decoded objects, before calling CreateHandler
			); err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Assess("Star Wars Demo", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}

			// check for decoded object object creation
			if err := client.Resources(namespace).Get(ctx, "deathstar", namespace, &v1.Service{}); err != nil {
				t.Fatal(err)
			}
			if err := client.Resources(namespace).Get(ctx, "deathstar", namespace, &appsv1.Deployment{}); err != nil {
				t.Fatal(err)
			}
			if err := client.Resources(namespace).Get(ctx, "spaceship", namespace, &appsv1.Deployment{}); err != nil {
				t.Fatal(err)
			}
			var xwing appsv1.Deployment
			if err := client.Resources(namespace).Get(ctx, "xwing", namespace, &xwing); err != nil {
				t.Fatal(err)
			}

			// get xwing pod names
			t.Logf("xwing deployment: %#v", xwing)

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			return ctx
		}).Feature()
	testenv.Test(t, starWarsDemoFeature)
}
