package driver

import (
	"os"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	mount "k8s.io/mount-utils"
)

func TestRunCleanupScanRemovesInactiveSources(t *testing.T) {
	const (
		namespace = "hf-csi"
		nodeID    = "node-1"
	)

	base := t.TempDir()
	oldBase := mountBaseDir
	mountBaseDir = base
	defer func() { mountBaseDir = oldBase }()

	mkSource := func(name string) string {
		t.Helper()
		dir := filepath.Join(base, name)
		file := filepath.Join(dir, "globiom-results", "15228", "out.rds")
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte("data"), 0o644); err != nil {
			t.Fatal(err)
		}
		return dir
	}

	inactive := mkSource("shared-node-1-inactive")
	withCRD := mkSource("shared-node-1-healing")
	withPod := mkSource("shared-node-1-running")
	mounted := mkSource("shared-node-1-mounted")

	mountPod := func(volumeID string) *corev1.Pod {
		return &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
			Name:      mountPodPrefix + volumeID,
			Namespace: namespace,
			Labels:    map[string]string{labelApp: labelAppValue, labelNode: sanitizeLabelValue(nodeID)},
		}}
	}
	client := fake.NewSimpleClientset(
		mountPod(filepath.Base(withPod)),
		mountPod(filepath.Base(mounted)),
	)

	crd := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": hfMountGVR.Group + "/" + hfMountGVR.Version,
		"kind":       "HFMount",
		"metadata":   map[string]interface{}{"name": hfMountName(filepath.Base(withCRD)), "namespace": namespace},
		"spec":       map[string]interface{}{"mountPath": withCRD},
	}}
	dynClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{hfMountGVR: "HFMountList"}, crd)

	m := &PodMounter{
		client:      client,
		namespace:   namespace,
		nodeID:      nodeID,
		checker:     mount.NewFakeMounter([]mount.MountPoint{{Path: mounted, Type: "fuse"}}),
		crd:         newHFMountClient(dynClient, namespace),
		binds:       map[string]string{},
		sourceLocks: map[string]*refMutex{},
	}

	t.Setenv("ACCELERATOR_MOUNT", "")
	m.runCleanupScan()
	if _, err := os.Stat(inactive); err != nil {
		t.Fatalf("inactive source must be kept without ACCELERATOR_MOUNT: %v", err)
	}

	t.Setenv("ACCELERATOR_MOUNT", "1")
	m.runCleanupScan()

	if _, err := os.Stat(inactive); !os.IsNotExist(err) {
		t.Errorf("inactive source should have been removed, stat err = %v", err)
	}
	for name, dir := range map[string]string{"source with CRD": withCRD, "source with mount pod": withPod, "mounted source": mounted} {
		if _, err := os.Stat(filepath.Join(dir, "globiom-results", "15228", "out.rds")); err != nil {
			t.Errorf("%s should have been kept: %v", name, err)
		}
	}
}
