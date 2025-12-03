package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

const (
	defaultNamespace   = "thing"
	defaultConfigMap   = "thing"
	defaultResyncPeriod = 30 * time.Second
)

type Controller struct {
	clientset      *kubernetes.Clientset
	namespace      string
	configMapName  string
	informer       cache.SharedIndexInformer
}

func NewController(clientset *kubernetes.Clientset, namespace, configMapName string) *Controller {
	// Create a field selector to watch only the specific ConfigMap
	fieldSelector := fields.OneTermEqualSelector("metadata.name", configMapName).String()

	// Create a ListWatcher for the ConfigMap
	listWatcher := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		"configmaps",
		namespace,
		fields.ParseSelectorOrDie(fieldSelector),
	)

	// Create the shared informer
	informer := cache.NewSharedIndexInformer(
		listWatcher,
		&corev1.ConfigMap{},
		defaultResyncPeriod,
		cache.Indexers{},
	)

	controller := &Controller{
		clientset:     clientset,
		namespace:     namespace,
		configMapName: configMapName,
		informer:      informer,
	}

	// Add event handlers
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.handleAdd,
		UpdateFunc: controller.handleUpdate,
		DeleteFunc: controller.handleDelete,
	})

	return controller
}

func (c *Controller) handleAdd(obj interface{}) {
	cm := obj.(*corev1.ConfigMap)
	klog.Infof("ConfigMap ADDED: %s/%s", cm.Namespace, cm.Name)
	c.logConfigMapData(cm)
}

func (c *Controller) handleUpdate(oldObj, newObj interface{}) {
	oldCM := oldObj.(*corev1.ConfigMap)
	newCM := newObj.(*corev1.ConfigMap)

	if oldCM.ResourceVersion == newCM.ResourceVersion {
		return
	}

	klog.Infof("ConfigMap UPDATED: %s/%s", newCM.Namespace, newCM.Name)
	c.logConfigMapData(newCM)
}

func (c *Controller) handleDelete(obj interface{}) {
	cm := obj.(*corev1.ConfigMap)
	klog.Infof("ConfigMap DELETED: %s/%s", cm.Namespace, cm.Name)
}

func (c *Controller) logConfigMapData(cm *corev1.ConfigMap) {
	klog.Infof("  ResourceVersion: %s", cm.ResourceVersion)
	klog.Infof("  Data entries: %d", len(cm.Data))
	for key, value := range cm.Data {
		// Truncate long values for logging
		displayValue := value
		if len(value) > 100 {
			displayValue = value[:100] + "..."
		}
		klog.Infof("    %s: %s", key, displayValue)
	}
}

func (c *Controller) Run(ctx context.Context) error {
	klog.Infof("Starting controller for ConfigMap '%s' in namespace '%s'", c.configMapName, c.namespace)

	// Start the informer
	go c.informer.Run(ctx.Done())

	// Wait for the cache to sync
	if !cache.WaitForCacheSync(ctx.Done(), c.informer.HasSynced) {
		return fmt.Errorf("failed to sync cache")
	}

	klog.Info("Controller cache synced, ready to process events")

	// Wait for context cancellation
	<-ctx.Done()
	klog.Info("Shutting down controller")
	return nil
}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func main() {
	var kubeconfig string
	var namespace string
	var configMapName string

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (optional, uses in-cluster config if not provided)")
	flag.StringVar(&namespace, "namespace", defaultNamespace, "Namespace to watch")
	flag.StringVar(&configMapName, "configmap", defaultConfigMap, "ConfigMap name to watch")

	klog.InitFlags(nil)
	flag.Parse()

	klog.Infof("Starting kube-thing-controller")
	klog.Infof("Namespace: %s", namespace)
	klog.Infof("ConfigMap: %s", configMapName)

	// Build the Kubernetes config
	config, err := buildConfig(kubeconfig)
	if err != nil {
		klog.Fatalf("Failed to build config: %v", err)
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Failed to create clientset: %v", err)
	}

	// Create the controller
	controller := NewController(clientset, namespace, configMapName)

	// Set up signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Run the controller
	if err := controller.Run(ctx); err != nil {
		klog.Fatalf("Error running controller: %v", err)
	}

	klog.Info("Controller stopped")
	os.Exit(0)
}
