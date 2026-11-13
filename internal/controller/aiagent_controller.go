package controller

import (
	"context"
	"fmt"

	aiv1alpha1 "github.com/gnishanth0728/kubernetes-ai-agent-operator/api/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// +kubebuilder:rbac:groups=ai.example.com,resources=aiagents,verbs=get;list;watch
// +kubebuilder:rbac:groups=ai.example.com,resources=aiagents/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

type AIAgentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *AIAgentReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (ctrl.Result, error) {

	logger := log.FromContext(ctx)

	// 1. Fetch AIAgent custom resource
	agent := &aiv1alpha1.AIAgent{}

	if err := r.Get(ctx, req.NamespacedName, agent); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	logger.Info("Reconciling AI Agent", "name", agent.Name)

	replicas := int32(1)
	if agent.Spec.Replicas != nil {
		replicas = *agent.Spec.Replicas
	}

	if replicas < 0 {
		return ctrl.Result{}, fmt.Errorf("replicas cannot be negative")
	}

	labels := map[string]string{
		"app":                          agent.Name,
		"app.kubernetes.io/managed-by": "ai-agent-operator",
	}

	// 2. Create or update Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      agent.Name,
			Namespace: agent.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(
		ctx, r.Client, deployment,
		func() error {
			if err := controllerutil.SetControllerReference(
				agent, deployment, r.Scheme,
			); err != nil {
				return err
			}

			deployment.Spec.Replicas = &replicas

			deployment.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: labels,
			}

			deployment.Spec.Template.ObjectMeta.Labels = labels

			deployment.Spec.Template.Spec.Containers = []corev1.Container{
				{
					Name:            "ai-agent",
					Image:           agent.Spec.Image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: 8000,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					Resources: agent.Spec.Resources,
					Env: []corev1.EnvVar{
						{
							Name:  "AGENT_ROLE",
							Value: agent.Spec.Agent.Role,
						},
						{
							Name:  "LLM_PROVIDER",
							Value: agent.Spec.Model.Provider,
						},
						{
							Name:  "LLM_MODEL",
							Value: agent.Spec.Model.Name,
						},
					},
				},
			}

			return nil
		},
	)

	if err != nil {
		return ctrl.Result{}, err
	}

	// 3. Create or update Service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      agent.Name,
			Namespace: agent.Namespace,
		},
	}

	_, err = controllerutil.CreateOrUpdate(
		ctx, r.Client, service,
		func() error {
			if err := controllerutil.SetControllerReference(
				agent, service, r.Scheme,
			); err != nil {
				return err
			}

			service.Spec.Selector = labels
			service.Spec.Ports = []corev1.ServicePort{
				{
					Name:       "http",
					Port:       8000,
					TargetPort: intstr.FromInt(8000),
					Protocol:   corev1.ProtocolTCP,
				},
			}

			service.Spec.Type = corev1.ServiceTypeClusterIP

			return nil
		},
	)

	if err != nil {
		return ctrl.Result{}, err
	}

	// 4. Update AIAgent status
	ready := deployment.Status.ReadyReplicas

	phase := "Provisioning"
	if ready == replicas {
		phase = "Running"
	}

	if agent.Status.Phase != phase ||
		agent.Status.ReadyReplicas != ready {

		agent.Status.Phase = phase
		agent.Status.ReadyReplicas = ready
		agent.Status.Message = "Agent Deployment reconciled"

		if err := r.Status().Update(ctx, agent); err != nil {
			return ctrl.Result{}, err
		}
	}

	logger.Info(
		"AI Agent reconciliation completed",
		"phase", phase,
		"readyReplicas", ready,
	)

	return ctrl.Result{}, nil
}

func (r *AIAgentReconciler) SetupWithManager(
	mgr ctrl.Manager,
) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&aiv1alpha1.AIAgent{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
