package util

import (
	"reflect"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	appsv1 "github.com/openshift/api/apps/v1"
	appstest "github.com/openshift/origin/pkg/apps/util/test"
)

func TestPodName(t *testing.T) {
	deployment := &corev1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testName",
		},
	}
	expected := "testName-deploy"
	actual := DeployerPodNameForDeployment(deployment.Name)
	if expected != actual {
		t.Errorf("Unexpected pod name for deployment. Expected: %s Got: %s", expected, actual)
	}
}

func TestCanTransitionPhase(t *testing.T) {
	tests := []struct {
		name          string
		current, next DeploymentStatus
		expected      bool
	}{
		{
			name:     "New->New",
			current:  DeploymentStatusNew,
			next:     DeploymentStatusNew,
			expected: false,
		},
		{
			name:     "New->Pending",
			current:  DeploymentStatusNew,
			next:     DeploymentStatusPending,
			expected: true,
		},
		{
			name:     "New->Running",
			current:  DeploymentStatusNew,
			next:     DeploymentStatusRunning,
			expected: true,
		},
		{
			name:     "New->Complete",
			current:  DeploymentStatusNew,
			next:     DeploymentStatusComplete,
			expected: true,
		},
		{
			name:     "New->Failed",
			current:  DeploymentStatusNew,
			next:     DeploymentStatusFailed,
			expected: true,
		},
		{
			name:     "Pending->New",
			current:  DeploymentStatusPending,
			next:     DeploymentStatusNew,
			expected: false,
		},
		{
			name:     "Pending->Pending",
			current:  DeploymentStatusPending,
			next:     DeploymentStatusPending,
			expected: false,
		},
		{
			name:     "Pending->Running",
			current:  DeploymentStatusPending,
			next:     DeploymentStatusRunning,
			expected: true,
		},
		{
			name:     "Pending->Failed",
			current:  DeploymentStatusPending,
			next:     DeploymentStatusFailed,
			expected: true,
		},
		{
			name:     "Pending->Complete",
			current:  DeploymentStatusPending,
			next:     DeploymentStatusComplete,
			expected: true,
		},
		{
			name:     "Running->New",
			current:  DeploymentStatusRunning,
			next:     DeploymentStatusNew,
			expected: false,
		},
		{
			name:     "Running->Pending",
			current:  DeploymentStatusRunning,
			next:     DeploymentStatusPending,
			expected: false,
		},
		{
			name:     "Running->Running",
			current:  DeploymentStatusRunning,
			next:     DeploymentStatusRunning,
			expected: false,
		},
		{
			name:     "Running->Failed",
			current:  DeploymentStatusRunning,
			next:     DeploymentStatusFailed,
			expected: true,
		},
		{
			name:     "Running->Complete",
			current:  DeploymentStatusRunning,
			next:     DeploymentStatusComplete,
			expected: true,
		},
		{
			name:     "Complete->New",
			current:  DeploymentStatusComplete,
			next:     DeploymentStatusNew,
			expected: false,
		},
		{
			name:     "Complete->Pending",
			current:  DeploymentStatusComplete,
			next:     DeploymentStatusPending,
			expected: false,
		},
		{
			name:     "Complete->Running",
			current:  DeploymentStatusComplete,
			next:     DeploymentStatusRunning,
			expected: false,
		},
		{
			name:     "Complete->Failed",
			current:  DeploymentStatusComplete,
			next:     DeploymentStatusFailed,
			expected: false,
		},
		{
			name:     "Complete->Complete",
			current:  DeploymentStatusComplete,
			next:     DeploymentStatusComplete,
			expected: false,
		},
		{
			name:     "Failed->New",
			current:  DeploymentStatusFailed,
			next:     DeploymentStatusNew,
			expected: false,
		},
		{
			name:     "Failed->Pending",
			current:  DeploymentStatusFailed,
			next:     DeploymentStatusPending,
			expected: false,
		},
		{
			name:     "Failed->Running",
			current:  DeploymentStatusFailed,
			next:     DeploymentStatusRunning,
			expected: false,
		},
		{
			name:     "Failed->Complete",
			current:  DeploymentStatusFailed,
			next:     DeploymentStatusComplete,
			expected: false,
		},
		{
			name:     "Failed->Failed",
			current:  DeploymentStatusFailed,
			next:     DeploymentStatusFailed,
			expected: false,
		},
	}

	for _, test := range tests {
		got := CanTransitionPhase(test.current, test.next)
		if got != test.expected {
			t.Errorf("%s: expected %t, got %t", test.name, test.expected, got)
		}
	}
}

var (
	now     = metav1.Now()
	later   = metav1.Time{Time: now.Add(time.Minute)}
	earlier = metav1.Time{Time: now.Add(-time.Minute)}

	condProgressing = func() appsv1.DeploymentCondition {
		return appsv1.DeploymentCondition{
			Type:               appsv1.DeploymentProgressing,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: now,
		}
	}

	condProgressingDifferentTime = func() appsv1.DeploymentCondition {
		return appsv1.DeploymentCondition{
			Type:               appsv1.DeploymentProgressing,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: later,
		}
	}

	condProgressingDifferentReason = func() appsv1.DeploymentCondition {
		return appsv1.DeploymentCondition{
			Type:               appsv1.DeploymentProgressing,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: later,
			Reason:             NewReplicationControllerReason,
		}
	}

	condNotProgressing = func() appsv1.DeploymentCondition {
		return appsv1.DeploymentCondition{
			Type:               appsv1.DeploymentProgressing,
			Status:             corev1.ConditionFalse,
			LastUpdateTime:     earlier,
			LastTransitionTime: earlier,
		}
	}

	condAvailable = func() appsv1.DeploymentCondition {
		return appsv1.DeploymentCondition{
			Type:   appsv1.DeploymentAvailable,
			Status: corev1.ConditionTrue,
		}
	}
)

func TestGetCondition(t *testing.T) {
	exampleStatus := func() appsv1.DeploymentConfigStatus {
		return appsv1.DeploymentConfigStatus{
			Conditions: []appsv1.DeploymentCondition{condProgressing(), condAvailable()},
		}
	}

	tests := []struct {
		name string

		status     appsv1.DeploymentConfigStatus
		condType   appsv1.DeploymentConditionType
		condStatus corev1.ConditionStatus

		expected bool
	}{
		{
			name: "condition exists",

			status:   exampleStatus(),
			condType: appsv1.DeploymentAvailable,

			expected: true,
		},
		{
			name: "condition does not exist",

			status:   exampleStatus(),
			condType: appsv1.DeploymentReplicaFailure,

			expected: false,
		},
	}

	for _, test := range tests {
		cond := GetDeploymentCondition(test.status, test.condType)
		exists := cond != nil
		if exists != test.expected {
			t.Errorf("%s: expected condition to exist: %t, got: %t", test.name, test.expected, exists)
		}
	}
}

func TestSetCondition(t *testing.T) {
	tests := []struct {
		name string

		status *appsv1.DeploymentConfigStatus
		cond   appsv1.DeploymentCondition

		expectedStatus *appsv1.DeploymentConfigStatus
	}{
		{
			name: "set for the first time",

			status: &appsv1.DeploymentConfigStatus{},
			cond:   condAvailable(),

			expectedStatus: &appsv1.DeploymentConfigStatus{
				Conditions: []appsv1.DeploymentCondition{
					condAvailable(),
				},
			},
		},
		{
			name: "simple set",

			status: &appsv1.DeploymentConfigStatus{
				Conditions: []appsv1.DeploymentCondition{
					condProgressing(),
				},
			},
			cond: condAvailable(),

			expectedStatus: &appsv1.DeploymentConfigStatus{
				Conditions: []appsv1.DeploymentCondition{
					condProgressing(), condAvailable(),
				},
			},
		},
		{
			name: "replace if status changes",

			status: &appsv1.DeploymentConfigStatus{
				Conditions: []appsv1.DeploymentCondition{
					condNotProgressing(),
				},
			},
			cond: condProgressing(),

			expectedStatus: &appsv1.DeploymentConfigStatus{Conditions: []appsv1.DeploymentCondition{condProgressing()}},
		},
		{
			name: "replace if reason changes",

			status: &appsv1.DeploymentConfigStatus{
				Conditions: []appsv1.DeploymentCondition{
					condProgressing(),
				},
			},
			cond: condProgressingDifferentReason(),

			expectedStatus: &appsv1.DeploymentConfigStatus{
				Conditions: []appsv1.DeploymentCondition{
					{
						Type:   appsv1.DeploymentProgressing,
						Status: corev1.ConditionTrue,
						// Note that LastTransitionTime stays the same.
						LastTransitionTime: now,
						// Only the reason changes.
						Reason: NewReplicationControllerReason,
					},
				},
			},
		},
		{
			name: "don't replace if status and reason don't change",

			status: &appsv1.DeploymentConfigStatus{
				Conditions: []appsv1.DeploymentCondition{
					condProgressing(),
				},
			},
			cond: condProgressingDifferentTime(),

			expectedStatus: &appsv1.DeploymentConfigStatus{Conditions: []appsv1.DeploymentCondition{condProgressing()}},
		},
	}

	for _, test := range tests {
		t.Logf("running test %q", test.name)
		SetDeploymentCondition(test.status, test.cond)
		if !reflect.DeepEqual(test.status, test.expectedStatus) {
			t.Errorf("expected status: %v, got: %v", test.expectedStatus, test.status)
		}
	}
}

func TestRemoveCondition(t *testing.T) {
	exampleStatus := func() *appsv1.DeploymentConfigStatus {
		return &appsv1.DeploymentConfigStatus{
			Conditions: []appsv1.DeploymentCondition{condProgressing(), condAvailable()},
		}
	}

	tests := []struct {
		name string

		status   *appsv1.DeploymentConfigStatus
		condType appsv1.DeploymentConditionType

		expectedStatus *appsv1.DeploymentConfigStatus
	}{
		{
			name: "remove from empty status",

			status:   &appsv1.DeploymentConfigStatus{},
			condType: appsv1.DeploymentProgressing,

			expectedStatus: &appsv1.DeploymentConfigStatus{},
		},
		{
			name: "simple remove",

			status:   &appsv1.DeploymentConfigStatus{Conditions: []appsv1.DeploymentCondition{condProgressing()}},
			condType: appsv1.DeploymentProgressing,

			expectedStatus: &appsv1.DeploymentConfigStatus{},
		},
		{
			name: "doesn't remove anything",

			status:   exampleStatus(),
			condType: appsv1.DeploymentReplicaFailure,

			expectedStatus: exampleStatus(),
		},
	}

	for _, test := range tests {
		RemoveDeploymentCondition(test.status, test.condType)
		if !reflect.DeepEqual(test.status, test.expectedStatus) {
			t.Errorf("%s: expected status: %v, got: %v", test.name, test.expectedStatus, test.status)
		}
	}
}

func TestRolloutExceededTimeoutSeconds(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name                   string
		config                 *appsv1.DeploymentConfig
		deploymentCreationTime time.Time
		expectTimeout          bool
	}{
		// Recreate strategy with deployment running for 20s (exceeding 10s timeout)
		{
			name: "recreate timeout",
			config: func(timeoutSeconds int64) *appsv1.DeploymentConfig {
				config := appstest.OkDeploymentConfig(1)
				config.Spec.Strategy.RecreateParams.TimeoutSeconds = &timeoutSeconds
				return config
			}(int64(10)),
			deploymentCreationTime: now.Add(-20 * time.Second),
			expectTimeout:          true,
		},
		// Recreate strategy with no timeout
		{
			name: "recreate no timeout",
			config: func(timeoutSeconds int64) *appsv1.DeploymentConfig {
				config := appstest.OkDeploymentConfig(1)
				config.Spec.Strategy.RecreateParams.TimeoutSeconds = &timeoutSeconds
				return config
			}(int64(0)),
			deploymentCreationTime: now.Add(-700 * time.Second),
			expectTimeout:          false,
		},

		// Rolling strategy with deployment running for 20s (exceeding 10s timeout)
		{
			name: "rolling timeout",
			config: func(timeoutSeconds int64) *appsv1.DeploymentConfig {
				config := appstest.OkDeploymentConfig(1)
				config.Spec.Strategy = appstest.OkRollingStrategy()
				config.Spec.Strategy.RollingParams.TimeoutSeconds = &timeoutSeconds
				return config
			}(int64(10)),
			deploymentCreationTime: now.Add(-20 * time.Second),
			expectTimeout:          true,
		},
		// Rolling strategy with deployment with no timeout specified.
		{
			name: "rolling using default timeout",
			config: func(timeoutSeconds int64) *appsv1.DeploymentConfig {
				config := appstest.OkDeploymentConfig(1)
				config.Spec.Strategy = appstest.OkRollingStrategy()
				config.Spec.Strategy.RollingParams.TimeoutSeconds = nil
				return config
			}(0),
			deploymentCreationTime: now.Add(-20 * time.Second),
			expectTimeout:          false,
		},
		// Recreate strategy with deployment with no timeout specified.
		{
			name: "recreate using default timeout",
			config: func(timeoutSeconds int64) *appsv1.DeploymentConfig {
				config := appstest.OkDeploymentConfig(1)
				config.Spec.Strategy.RecreateParams.TimeoutSeconds = nil
				return config
			}(0),
			deploymentCreationTime: now.Add(-20 * time.Second),
			expectTimeout:          false,
		},
		// Custom strategy with deployment with no timeout specified.
		{
			name: "custom using default timeout",
			config: func(timeoutSeconds int64) *appsv1.DeploymentConfig {
				config := appstest.OkDeploymentConfig(1)
				config.Spec.Strategy = appstest.OkCustomStrategy()
				return config
			}(0),
			deploymentCreationTime: now.Add(-20 * time.Second),
			expectTimeout:          false,
		},
		// Custom strategy use default timeout exceeding it.
		{
			name: "custom using default timeout timing out",
			config: func(timeoutSeconds int64) *appsv1.DeploymentConfig {
				config := appstest.OkDeploymentConfig(1)
				config.Spec.Strategy = appstest.OkCustomStrategy()
				return config
			}(0),
			deploymentCreationTime: now.Add(-700 * time.Second),
			expectTimeout:          true,
		},
	}

	for _, tc := range tests {
		config := tc.config
		deployment, err := MakeDeployment(config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		deployment.ObjectMeta.CreationTimestamp = metav1.Time{Time: tc.deploymentCreationTime}
		gotTimeout := RolloutExceededTimeoutSeconds(config, deployment)
		if tc.expectTimeout && !gotTimeout {
			t.Errorf("[%s]: expected timeout, but got no timeout", tc.name)
		}
		if !tc.expectTimeout && gotTimeout {
			t.Errorf("[%s]: expected no timeout, but got timeout", tc.name)
		}

	}
}
