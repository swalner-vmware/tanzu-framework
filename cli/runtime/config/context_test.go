// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
)

func setup(t *testing.T) {
	LocalDirName = fmt.Sprintf(".tanzu-test-%s", randString())
	cfg := &v1alpha1.ClientConfig{
		KnownContexts: []*v1alpha1.Context{
			{
				Name: "test-mc",
				Type: v1alpha1.CtxTypeK8s,
				ClusterOpts: &v1alpha1.ClusterServer{
					Endpoint:            "test-endpoint",
					Path:                "test-path",
					Context:             "test-context",
					IsManagementCluster: true,
				},
			},
			{
				Name: "test-tmc",
				Type: v1alpha1.CtxTypeTMC,
				GlobalOpts: &v1alpha1.GlobalServer{
					Endpoint: "test-endpoint",
				},
			},
		},
		CurrentContext: map[v1alpha1.ContextType]string{
			v1alpha1.CtxTypeK8s: "test-mc",
			v1alpha1.CtxTypeTMC: "test-tmc",
		},
	}

	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	err := StoreClientConfig(cfg)
	require.NoError(t, err)
}

func cleanup() {
	cleanupDir(LocalDirName)
}

func TestGetContext(t *testing.T) {
	setup(t)
	defer cleanup()

	tcs := []struct {
		name    string
		ctxName string
		errStr  string
	}{
		{
			name:    "success k8s",
			ctxName: "test-mc",
		},
		{
			name:    "success tmc",
			ctxName: "test-tmc",
		},
		{
			name:    "failure",
			ctxName: "test",
			errStr:  "could not find context \"test\"",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			c, err := GetContext(tc.ctxName)
			if tc.errStr == "" {
				assert.Equal(t, tc.ctxName, c.Name)
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errStr)
			}
		})
	}
}

func TestContextExists(t *testing.T) {
	setup(t)
	defer cleanup()

	tcs := []struct {
		name    string
		ctxName string
		ok      bool
	}{
		{
			name:    "success k8s",
			ctxName: "test-mc",
			ok:      true,
		},
		{
			name:    "success tmc",
			ctxName: "test-tmc",
			ok:      true,
		},
		{
			name:    "failure",
			ctxName: "test",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := ContextExists(tc.ctxName)
			assert.Equal(t, tc.ok, ok)
			assert.NoError(t, err)
		})
	}
}

func TestAddContext(t *testing.T) {
	setup(t)
	defer cleanup()

	tcs := []struct {
		name    string
		ctx     *v1alpha1.Context
		current bool
		errStr  string
	}{
		{
			name: "success k8s current",
			ctx: &v1alpha1.Context{
				Name: "test-mc1",
				Type: v1alpha1.CtxTypeK8s,
				ClusterOpts: &v1alpha1.ClusterServer{
					Endpoint:            "test-endpoint",
					Path:                "test-path",
					Context:             "test-context",
					IsManagementCluster: true,
				},
			},
			current: true,
		},
		{
			name: "success k8s not_current",
			ctx: &v1alpha1.Context{
				Name: "test-mc2",
				Type: v1alpha1.CtxTypeK8s,
				ClusterOpts: &v1alpha1.ClusterServer{
					Endpoint:            "test-endpoint",
					Path:                "test-path",
					Context:             "test-context",
					IsManagementCluster: true,
				},
			},
		},
		{
			name: "success tmc current",
			ctx: &v1alpha1.Context{
				Name: "test-tmc1",
				Type: v1alpha1.CtxTypeTMC,
				GlobalOpts: &v1alpha1.GlobalServer{
					Endpoint: "test-endpoint",
				},
			},
			current: true,
		},
		{
			name: "success tmc not_current",
			ctx: &v1alpha1.Context{
				Name: "test-tmc2",
				Type: v1alpha1.CtxTypeTMC,
				GlobalOpts: &v1alpha1.GlobalServer{
					Endpoint: "test-endpoint",
				},
			},
		},
		{
			name: "failure k8s",
			ctx: &v1alpha1.Context{
				Name: "test-mc",
				Type: v1alpha1.CtxTypeK8s,
				ClusterOpts: &v1alpha1.ClusterServer{
					Endpoint:            "test-endpoint",
					Path:                "test-path",
					Context:             "test-context",
					IsManagementCluster: true,
				},
			},
			errStr: "context \"test-mc\" already exists",
		},
		{
			name: "failure tmc",
			ctx: &v1alpha1.Context{
				Name: "test-tmc",
				Type: v1alpha1.CtxTypeTMC,
				GlobalOpts: &v1alpha1.GlobalServer{
					Endpoint: "test-endpoint",
				},
			},
			errStr: "context \"test-tmc\" already exists",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.errStr == "" {
				ok, err := ContextExists(tc.ctx.Name)
				require.False(t, ok)
				require.NoError(t, err)
			}

			err := AddContext(tc.ctx, tc.current)
			if tc.errStr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errStr)
			}

			ok, err := ContextExists(tc.ctx.Name)
			assert.True(t, ok)
			assert.NoError(t, err)
			ok, err = ServerExists(tc.ctx.Name)
			assert.True(t, ok)
			assert.NoError(t, err)
		})
	}
}

func TestRemoveContext(t *testing.T) {
	setup(t)
	defer cleanup()

	tcs := []struct {
		name    string
		ctxName string
		ctxType v1alpha1.ContextType
		errStr  string
	}{
		{
			name:    "success k8s",
			ctxName: "test-mc",
			ctxType: v1alpha1.CtxTypeK8s,
		},
		{
			name:    "success tmc",
			ctxName: "test-tmc",
			ctxType: v1alpha1.CtxTypeTMC,
		},
		{
			name:    "failure",
			ctxName: "test",
			errStr:  "context test not found",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.errStr == "" {
				ok, err := ContextExists(tc.ctxName)
				require.True(t, ok)
				require.NoError(t, err)
			}

			err := RemoveContext(tc.ctxName)
			if tc.errStr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errStr)
			}

			ok, err := ContextExists(tc.ctxName)
			assert.False(t, ok)
			assert.NoError(t, err)
			ok, err = ServerExists(tc.ctxName)
			assert.False(t, ok)
			assert.NoError(t, err)
		})
	}
}

func TestSetCurrentContext(t *testing.T) {
	setup(t)
	defer cleanup()

	tcs := []struct {
		name    string
		ctxType v1alpha1.ContextType
		ctxName string
		errStr  string
	}{
		{
			name:    "success k8s",
			ctxName: "test-mc",
			ctxType: v1alpha1.CtxTypeK8s,
		},
		{
			name:    "success tmc",
			ctxName: "test-tmc",
			ctxType: v1alpha1.CtxTypeTMC,
		},
		{
			name:    "failure",
			ctxName: "test",
			errStr:  "could not find context \"test\"",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := SetCurrentContext(tc.ctxName)
			if tc.errStr == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.errStr)
			}

			currCtx, err := GetCurrentContext(tc.ctxType)
			if tc.errStr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.ctxName, currCtx.Name)
			} else {
				assert.Error(t, err)
			}
			currSrv, err := GetCurrentServer()
			assert.NoError(t, err)
			// server is updated only for K8s Ctx Type, so test-mc remains the current server
			assert.Equal(t, "test-mc", currSrv.Name)
		})
	}
}

func TestGetCurrentContext(t *testing.T) {
	setup(t)
	defer cleanup()

	tcs := []struct {
		name    string
		ctxType v1alpha1.ContextType
		ctxName string
		errStr  string
	}{
		{
			name:    "success k8s",
			ctxType: v1alpha1.CtxTypeK8s,
			ctxName: "test-mc",
		},
		{
			name:    "success tmc",
			ctxType: v1alpha1.CtxTypeTMC,
			ctxName: "test-tmc",
		},
		{
			name:    "failure k8s",
			ctxType: v1alpha1.CtxTypeK8s,
			ctxName: "test-mc",
			errStr:  "no current context set for type \"k8s\"",
		},
		{
			name:    "failure tmc",
			ctxType: v1alpha1.CtxTypeTMC,
			ctxName: "test-tmc",
			errStr:  "no current context set for type \"tmc\"",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if tc.errStr != "" {
				err := RemoveContext(tc.ctxName)
				require.NoError(t, err)
			}

			curr, err := GetCurrentContext(tc.ctxType)
			if tc.errStr == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.ctxName, curr.Name)
			} else {
				assert.EqualError(t, err, tc.errStr)
			}
		})
	}
}
