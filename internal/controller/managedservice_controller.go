package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	v1alpha1 "github.com/cristianverduzco/nanodeploy/api/v1alpha1"
)

// ManagedServiceReconciler reconciles a ManagedService object
type ManagedServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// +kubebuilder:rbac:groups=nanodeploy.io,resources=managedservices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nanodeploy.io,resources=managedservices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete

// Reconcile is the core control loop — called every time a ManagedService is created, updated, or deleted
func (r *ManagedServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling ManagedService", "name", req.Name, "namespace", req.Namespace)

	// 1. Fetch the ManagedService resource
	managedService := &v1alpha1.ManagedService{}
	if err := r.Get(ctx, req.NamespacedName, managedService); err != nil {
		if errors.IsNotFound(err) {
			// Resource was deleted — nothing to do
			logger.Info("ManagedService not found, likely deleted", "name", req.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get ManagedService: %w", err)
	}

	// 2. Set initial phase if not set
	if managedService.Status.Phase == "" {
		if err := r.updateStatus(ctx, managedService, v1alpha1.ServicePhasePending, "Initializing"); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// 3. Route to the correct provisioner based on service type
	switch managedService.Spec.Type {
	case v1alpha1.ServiceTypePostgresql:
		return r.reconcilePostgresql(ctx, managedService)
	case v1alpha1.ServiceTypeRedis:
		return r.reconcileRedis(ctx, managedService)
	default:
		msg := fmt.Sprintf("unsupported service type: %s", managedService.Spec.Type)
		logger.Error(nil, msg)
		if err := r.updateStatus(ctx, managedService, v1alpha1.ServicePhaseFailed, msg); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
}

// reconcilePostgresql handles the full lifecycle of a PostgreSQL ManagedService
func (r *ManagedServiceReconciler) reconcilePostgresql(ctx context.Context, ms *v1alpha1.ManagedService) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Update phase to Provisioning
	if ms.Status.Phase == v1alpha1.ServicePhasePending {
		if err := r.updateStatus(ctx, ms, v1alpha1.ServicePhaseProvisioning, "Provisioning PostgreSQL"); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Ensure the Deployment exists
	if err := r.ensureDeployment(ctx, ms); err != nil {
		logger.Error(err, "Failed to ensure Deployment")
		_ = r.updateStatus(ctx, ms, v1alpha1.ServicePhaseFailed, err.Error())
		return ctrl.Result{}, err
	}

	// Ensure the Service exists
	if err := r.ensureService(ctx, ms); err != nil {
		logger.Error(err, "Failed to ensure Service")
		_ = r.updateStatus(ctx, ms, v1alpha1.ServicePhaseFailed, err.Error())
		return ctrl.Result{}, err
	}

	// Mark as Ready
	endpoint := fmt.Sprintf("%s.%s.svc.cluster.local", ms.Name, ms.Namespace)
	if err := r.updateStatusReady(ctx, ms, endpoint); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("PostgreSQL ManagedService is Ready", "endpoint", endpoint)
	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// reconcileRedis handles the full lifecycle of a Redis ManagedService
func (r *ManagedServiceReconciler) reconcileRedis(ctx context.Context, ms *v1alpha1.ManagedService) (ctrl.Result, error) {
	// Placeholder — we'll implement this in Phase 2
	return ctrl.Result{}, r.updateStatus(ctx, ms, v1alpha1.ServicePhaseFailed, "Redis provisioner not yet implemented")
}

// ensureDeployment creates the Deployment for a ManagedService if it doesn't exist
func (r *ManagedServiceReconciler) ensureDeployment(ctx context.Context, ms *v1alpha1.ManagedService) error {
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: ms.Name, Namespace: ms.Namespace}, deployment)
	if err == nil {
		// Already exists
		return nil
	}
	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get Deployment: %w", err)
	}

	// Build the Deployment
	desired := r.buildDeployment(ms)
	if err := ctrl.SetControllerReference(ms, desired, r.Scheme); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}

	return r.Create(ctx, desired)
}

// ensureService creates the Kubernetes Service for a ManagedService if it doesn't exist
func (r *ManagedServiceReconciler) ensureService(ctx context.Context, ms *v1alpha1.ManagedService) error {
	svc := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: ms.Name, Namespace: ms.Namespace}, svc)
	if err == nil {
		return nil
	}
	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get Service: %w", err)
	}

	desired := r.buildService(ms)
	if err := ctrl.SetControllerReference(ms, desired, r.Scheme); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}

	return r.Create(ctx, desired)
}

// buildDeployment constructs the Deployment object for a ManagedService
func (r *ManagedServiceReconciler) buildDeployment(ms *v1alpha1.ManagedService) *appsv1.Deployment {
	labels := map[string]string{
		"app":                          ms.Name,
		"nanodeploy.io/managed-by":     "nanodeploy",
		"nanodeploy.io/service-type":   string(ms.Spec.Type),
	}

	replicas := ms.Spec.Replicas
	if replicas == 0 {
		replicas = 1
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ms.Name,
			Namespace: ms.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "postgresql",
							Image: fmt.Sprintf("postgres:%s", ms.Spec.Version),
							Ports: []corev1.ContainerPort{
								{ContainerPort: 5432, Protocol: corev1.ProtocolTCP},
							},
							Env: []corev1.EnvVar{
								{Name: "POSTGRES_DB", Value: ms.Spec.DatabaseName},
								{Name: "POSTGRES_PASSWORD", Value: "nanodeploy-default"},
							},
						},
					},
				},
			},
		},
	}
}

// buildService constructs the Kubernetes Service object for a ManagedService
func (r *ManagedServiceReconciler) buildService(ms *v1alpha1.ManagedService) *corev1.Service {
	labels := map[string]string{
		"app":                      ms.Name,
		"nanodeploy.io/managed-by": "nanodeploy",
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ms.Name,
			Namespace: ms.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": ms.Name},
			Ports: []corev1.ServicePort{
				{Port: 5432, Protocol: corev1.ProtocolTCP},
			},
		},
	}
}

// updateStatus is a helper to update the status phase and message
func (r *ManagedServiceReconciler) updateStatus(ctx context.Context, ms *v1alpha1.ManagedService, phase v1alpha1.ServicePhase, message string) error {
	ms.Status.Phase = phase
	ms.Status.Message = message
	ms.Status.LastUpdated = metav1.Now()
	return r.Status().Update(ctx, ms)
}

// updateStatusReady marks a ManagedService as Ready with an endpoint
func (r *ManagedServiceReconciler) updateStatusReady(ctx context.Context, ms *v1alpha1.ManagedService, endpoint string) error {
	ms.Status.Phase = v1alpha1.ServicePhaseReady
	ms.Status.Message = "Service is ready"
	ms.Status.Endpoint = endpoint
	ms.Status.LastUpdated = metav1.Now()
	return r.Status().Update(ctx, ms)
}

// SetupWithManager registers the controller with the operator manager
func (r *ManagedServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ManagedService{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}