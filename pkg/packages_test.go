// Copyright 2018 jsonnet-bundler authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pkg

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	v1 "github.com/jsonnet-bundler/jsonnet-bundler/spec/v1"
	"github.com/jsonnet-bundler/jsonnet-bundler/spec/v1/deps"
)

// flakyInstaller fails its first failTimes Install calls then succeeds.
type flakyInstaller struct {
	failTimes int
	calls     int
}

func (f *flakyInstaller) Install(ctx context.Context, name, dir, version string) (string, error) {
	f.calls++
	if f.calls <= f.failTimes {
		return "", errors.New("flake")
	}
	return "ok", nil
}

func TestKnown(t *testing.T) {
	testDeps := deps.NewOrdered()
	testDeps.Set("ksonnet-lib", deps.Dependency{
		Source: deps.Source{
			GitSource: &deps.Git{
				Scheme: deps.GitSchemeHTTPS,
				Host:   "github.com",
				User:   "ksonnet",
				Repo:   "ksonnet-lib",
				Subdir: "/ksonnet.beta.4",
			},
		},
	})

	paths := []string{
		"github.com",
		"github.com/ksonnet",
		"github.com/ksonnet/ksonnet-lib",
		"github.com/ksonnet/ksonnet-lib/ksonnet.beta.4",
		"github.com/ksonnet/ksonnet-lib/ksonnet.beta.4/k.libsonnet",
		"github.com/ksonnet-util", // don't know that one
		"ksonnet.beta.4",          // the symlink
	}

	want := []string{
		"github.com",
		"github.com/ksonnet",
		"github.com/ksonnet/ksonnet",
		"github.com/ksonnet/ksonnet-lib",
		"github.com/ksonnet/ksonnet-lib/ksonnet.beta.4",
		"github.com/ksonnet/ksonnet-lib/ksonnet.beta.4/k.libsonnet",
	}

	w := make(map[string]bool)
	for _, k := range want {
		w[k] = true
	}

	for _, p := range paths {
		if known(testDeps, p) != w[p] {
			t.Fatalf("expected %s to be %v", p, w[p])
		}
	}
}

func TestInstallWithRetry(t *testing.T) {
	orig := downloadInitialBackoff
	t.Cleanup(func() { downloadInitialBackoff = orig })
	downloadInitialBackoff = time.Millisecond

	t.Run("first try succeeds", func(t *testing.T) {
		f := &flakyInstaller{failTimes: 0}
		v, err := installWithRetry(context.Background(), f, "x", "y", "z")
		require.NoError(t, err)
		require.Equal(t, "ok", v)
		require.Equal(t, 1, f.calls)
	})

	t.Run("eventual success retries", func(t *testing.T) {
		f := &flakyInstaller{failTimes: 1}
		_, err := installWithRetry(context.Background(), f, "x", "y", "z")
		require.NoError(t, err)
		require.Equal(t, 2, f.calls, "should retry once")
	})

	t.Run("gives up after max attempts", func(t *testing.T) {
		f := &flakyInstaller{failTimes: 99}
		_, err := installWithRetry(context.Background(), f, "x", "y", "z")
		require.Error(t, err)
		require.Equal(t, downloadAttempts, f.calls)
	})

	t.Run("cancelled ctx returns immediately", func(t *testing.T) {
		f := &flakyInstaller{failTimes: 99}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := installWithRetry(ctx, f, "x", "y", "z")
		require.ErrorIs(t, err, context.Canceled)
		// First Install runs once before the ctx check; the loop must
		// not keep retrying after that.
		require.LessOrEqual(t, f.calls, 1)
	})
}

// TestEnsureParallel verifies that Ensure handles many dependencies in
// parallel without losing input order, regardless of the Jobs setting. It
// uses local-source dependencies so no network is required, and relies on
// the race detector (run via `go test -race`) to catch concurrent map writes.
func TestEnsureParallel(t *testing.T) {
	for _, jobs := range []int{1, 4, 25, 100} {
		t.Run(fmt.Sprintf("jobs=%d", jobs), func(t *testing.T) {
			origCwd, err := os.Getwd()
			require.NoError(t, err)
			t.Cleanup(func() { _ = os.Chdir(origCwd) })

			origJobs := Jobs
			t.Cleanup(func() { Jobs = origJobs })
			Jobs = jobs

			base := t.TempDir()
			require.NoError(t, os.Chdir(base))

			vendorDir := filepath.Join(base, "vendor")
			require.NoError(t, os.MkdirAll(filepath.Join(vendorDir, ".tmp"), 0755))

			const N = 25
			direct := deps.NewOrdered()
			expectedOrder := make([]string, 0, N)
			for i := 0; i < N; i++ {
				name := fmt.Sprintf("pkg-%02d", i)
				require.NoError(t, os.Mkdir(filepath.Join(base, name), 0755))
				d := deps.Dependency{
					Source: deps.Source{
						LocalSource: &deps.Local{Directory: name},
					},
				}
				direct.Set(d.Name(), d)
				expectedOrder = append(expectedOrder, d.Name())
			}

			locked, err := Ensure(
				v1.JsonnetFile{Dependencies: direct, LegacyImports: false},
				vendorDir,
				deps.NewOrdered(),
			)
			require.NoError(t, err)
			require.Equal(t, expectedOrder, locked.Keys(),
				"parallel ensure must preserve input dependency order")

			for _, name := range expectedOrder {
				info, err := os.Lstat(filepath.Join(vendorDir, name))
				require.NoError(t, err, "missing entry for %s", name)
				require.NotZero(t, info.Mode()&os.ModeSymlink,
					"%s should be a symlink", name)
			}
		})
	}
}

// TestEnsureParallelError verifies that a failure in one parallel download
// surfaces as a non-nil error from Ensure rather than hanging or panicking.
func TestEnsureParallelError(t *testing.T) {
	origCwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(origCwd) })

	origJobs := Jobs
	t.Cleanup(func() { Jobs = origJobs })
	Jobs = 4

	base := t.TempDir()
	require.NoError(t, os.Chdir(base))

	vendorDir := filepath.Join(base, "vendor")
	require.NoError(t, os.MkdirAll(filepath.Join(vendorDir, ".tmp"), 0755))

	direct := deps.NewOrdered()
	bad := deps.Dependency{
		Source: deps.Source{
			LocalSource: &deps.Local{Directory: "does-not-exist"},
		},
	}
	direct.Set(bad.Name(), bad)

	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("good-%d", i)
		require.NoError(t, os.Mkdir(filepath.Join(base, name), 0755))
		d := deps.Dependency{
			Source: deps.Source{
				LocalSource: &deps.Local{Directory: name},
			},
		}
		direct.Set(d.Name(), d)
	}

	_, err = Ensure(
		v1.JsonnetFile{Dependencies: direct, LegacyImports: false},
		vendorDir,
		deps.NewOrdered(),
	)
	require.Error(t, err)
}

func TestCleanLegacyName(t *testing.T) {
	testList := func(name string) *deps.Ordered {
		l := deps.NewOrdered()
		l.Set("ksonnet-lib", deps.Dependency{
			LegacyNameCompat: name,
			Source: deps.Source{
				GitSource: &deps.Git{
					Scheme: deps.GitSchemeHTTPS,
					Host:   "github.com",
					User:   "ksonnet",
					Repo:   "ksonnet-lib",
					Subdir: "/ksonnet.beta.4",
				}},
		})
		return l
	}
	cases := map[string]bool{
		"ksonnet":        false,
		"ksonnet.beta.4": true,
	}

	for name, want := range cases {
		list := testList(name)
		CleanLegacyName(list)
		if (list.GetOrDefault("ksonnet-lib", deps.Dependency{}).LegacyNameCompat == "") != want {
			t.Fatalf("expected `%s` to be removed: %v", name, want)
		}
	}
}
