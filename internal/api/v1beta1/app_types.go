/*


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

package v1beta1

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/shipa-corp/ketch/internal/templates"
)

func init() {
	SchemeBuilder.Register(&App{}, &AppList{})
}

const (
	ShipaCloudDomain     = "shipa.cloud"
	DefaultNumberOfUnits = 1
)

// Env represents an environment variable present in an application.
type Env struct {
	// +kubebuilder:validation:MinLength=1
	// Name of the environment variable. Must be a C_IDENTIFIER.
	Name string `json:"name"`

	// Value of the environment variable.
	Value string `json:"value"`
}

// Label represents an environment variable present in an application.
type Label struct {
	// +kubebuilder:validation:MinLength=1
	// Name of the label.
	Name string `json:"name"`

	// Value of the label.
	Value string `json:"value"`
}

// CnameList is a list of an app's CNAMEs.
type CnameList []string

// RoutingSettings contains a weight of the current deployment used to route incoming traffic.
// If an application has two deployments with corresponding weights of 30 and 70,
// then 3 of 10 incoming requests will be sent to the first deployment (approximately).
type RoutingSettings struct {
	Weight int `json:"weight"`
}

// ProcessSpec is a specification of the desired behavior of a process.
type ProcessSpec struct {
	// +kubebuilder:validation:MinLength=1
	// Name of the process.
	Name string `json:"name"`

	// Units is a number of replicas of the process.
	Units *int `json:"units,omitempty"`

	// Env is a list of environment variables to set in pods created for the process.
	Env []Env `json:"env,omitempty"`

	// Commands executed on startup.
	Cmd []string `json:"cmd"`

	// Security options the process should run with.
	SecurityContext *v1.SecurityContext `json:"securityContext,omitempty"`
}

type DeploymentVersion int

func (v DeploymentVersion) String() string {
	return fmt.Sprintf("%d", v)
}

type AppDeploymentSpec struct {
	Image           string            `json:"image"`
	Version         DeploymentVersion `json:"version"`
	Processes       []ProcessSpec     `json:"processes,omitempty"`
	KetchYaml       *KetchYamlData    `json:"ketchYaml,omitempty"`
	Labels          []Label           `json:"labels,omitempty"`
	RoutingSettings RoutingSettings   `json:"routingSettings,omitempty"`
	ExposedPorts    []ExposedPort     `json:"exposedPorts,omitempty"`
}

// IngressSpec configures entrypoints to access an application.
type IngressSpec struct {

	// GenerateDefaultCname if set the application will have a default cname <app-name>.<ServiceEndpoint>.shipa.cloud.
	GenerateDefaultCname bool `json:"generateDefaultCname"`

	// Cnames is a list of additional cnames.
	Cnames CnameList `json:"cnames,omitempty"`
}

// DockerRegistrySpec contains docker registry configuration of an application.
type DockerRegistrySpec struct {

	// SecretName is added to the "imagePullSecrets" list of each application pod.
	SecretName string `json:"secretName,omitempty"`
}

// AppPhase is a label for the condition of an application at the current time.
type AppPhase string

const (
	// AppPending means the app has been accepted by the system, but has not been started.
	AppCreated AppPhase = "Created"

	// AppError means the app CRD is broken in some way and ketch controller can't render and install a new helm chart.
	AppError AppPhase = "Error"

	// AppRunning means that ketch controller has rendered a helm chart of the application and installed it to a cluster.
	AppRunning AppPhase = "Running"
)

type AppConditionType string

// These are valid conditions of app.
const (

	// AppScheduled indicates whether the has been processed by ketch-controller.
	AppScheduled AppConditionType = "Scheduled"
)

// AppCondition contains details for the current condition of this app.
type AppCondition struct {

	// Type of the condition.
	Type AppConditionType `json:"type"`

	// Status of the condition.
	Status v1.ConditionStatus `json:"status"`

	// LastTransitionTime is the timestamp corresponding to the last status.
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// A human readable message indicating details about why the application is in this condition.
	Message string `json:"message,omitempty"`
}

// AppStatus represents information about the status of an application.
type AppStatus struct {

	// Conditions of App resource.
	Conditions []AppCondition `json:"conditions,omitempty"`

	Pool *v1.ObjectReference `json:"pool,omitempty"`
}

// AppSpec defines the desired state of App.
type AppSpec struct {
	Version *string `json:"version,omitempty"`

	// +kubebuilder:validation:MaxLength=140
	Description string `json:"description,omitempty"`

	// Deployments is a list of running deployments.
	Deployments []AppDeploymentSpec `json:"deployments"`

	// DeploymentsCount is incremented every time a new deployment is added to Deployments and used as a version for new deployments.
	DeploymentsCount int `json:"deploymentsCount,omitempty"`

	// List of environment variables of the application.
	Env []Env `json:"env,omitempty"`

	// Pool is a name of a pool used to run the application.
	// +kubebuilder:validation:MinLength=1
	Pool string `json:"pool"`

	// Ingress contains configuration of entrypoints to access the application.
	Ingress IngressSpec `json:"ingress"`

	// DockerRegistry contains docker registry configuration of the application.
	DockerRegistry DockerRegistrySpec `json:"dockerRegisty,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Pool",type=string,JSONPath=`.spec.pool`
// +kubebuilder:printcolumn:name="Description",type=string,JSONPath=`.spec.description`

// App is the Schema for the apps API.
type App struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppSpec   `json:"spec,omitempty"`
	Status AppStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AppList contains a list of App.
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []App `json:"items"`
}

func (s *AppDeploymentSpec) setUnits(process string, units int) error {
	for i, processSpec := range s.Processes {
		if processSpec.Name == process {
			s.Processes[i].Units = &units
			return nil
		}
	}
	return ErrProcessNotFound
}

func (s *AppDeploymentSpec) setUnitsForAllProcess(units int) {
	for i := range s.Processes {
		s.Processes[i].Units = &units
	}
}

func (s *AppDeploymentSpec) addUnits(process string, units int) error {
	for i, processSpec := range s.Processes {
		if processSpec.Name == process {
			currentUnits := DefaultNumberOfUnits
			if processSpec.Units != nil {
				currentUnits = *processSpec.Units
			}
			newUnits := currentUnits + units
			if newUnits < 0 {
				newUnits = 0
			}
			s.Processes[i].Units = &newUnits
			return nil
		}
	}
	return ErrProcessNotFound
}

func (s *AppDeploymentSpec) addUnitsForAllProcess(units int) {
	for _, processSpec := range s.Processes {
		_ = s.addUnits(processSpec.Name, units)
	}
}

// SetUnits set quantity of units of the specified processes.
func (app *App) SetUnits(selector Selector, units int) error {
	deploymentFound := false
	for _, deploymentSpec := range app.Spec.Deployments {
		if selector.DeploymentVersion != nil && *selector.DeploymentVersion != deploymentSpec.Version {
			continue
		}
		if selector.Process != nil {
			if err := deploymentSpec.setUnits(*selector.Process, units); err != nil {
				return err
			}
		} else {
			deploymentSpec.setUnitsForAllProcess(units)
		}
		deploymentFound = true
	}
	if selector.DeploymentVersion != nil && !deploymentFound {
		return ErrDeploymentNotFound
	}
	return nil
}

// AddUnits add units to the specified processes.
func (app *App) AddUnits(selector Selector, units int) error {
	deploymentFound := false
	for _, deploymentSpec := range app.Spec.Deployments {
		if selector.DeploymentVersion != nil && *selector.DeploymentVersion != deploymentSpec.Version {
			continue
		}
		if selector.Process != nil {
			if err := deploymentSpec.addUnits(*selector.Process, units); err != nil {
				return err
			}
		} else {
			deploymentSpec.addUnitsForAllProcess(units)
		}
		deploymentFound = true
	}
	if selector.DeploymentVersion != nil && !deploymentFound {
		return ErrDeploymentNotFound
	}
	return nil
}

// SetEnvs extends the current list of environment variables with the provided list.
// If the current list has an env variable from the provided list, the env variable will be updated with a new value.
func (app *App) SetEnvs(envs []Env) {
	names := make(map[string]Env, len(envs))
	for _, env := range envs {
		names[env.Name] = env
	}
	newEnvs := make([]Env, 0, len(envs))
	for _, env := range app.Spec.Env {
		if newEnv, hasNewValue := names[env.Name]; hasNewValue {
			newEnvs = append(newEnvs, newEnv)
			delete(names, env.Name)
			continue
		}
		newEnvs = append(newEnvs, env)
	}
	for _, env := range names {
		newEnvs = append(newEnvs, env)
	}
	app.Spec.Env = newEnvs
}

// Envs returns values of the asked env variables.
func (app *App) Envs(names []string) map[string]string {
	namesMap := make(map[string]struct{}, len(names))
	for _, name := range names {
		namesMap[name] = struct{}{}
	}

	envs := make(map[string]string)
	for _, env := range app.Spec.Env {
		if len(names) == 0 {
			envs[env.Name] = env.Value
			continue
		}
		if _, ok := namesMap[env.Name]; ok {
			envs[env.Name] = env.Value
		}
	}
	return envs
}

// UnsetEnvs unsets environment values.
func (app *App) UnsetEnvs(envs []string) {
	names := make(map[string]struct{}, len(envs))
	for _, name := range envs {
		names[name] = struct{}{}
	}
	var newEnvs []Env
	for _, env := range app.Spec.Env {
		if _, remove := names[env.Name]; !remove {
			newEnvs = append(newEnvs, env)
		}
	}
	app.Spec.Env = newEnvs
}

// Stop stops processes specified by the selector.
func (app *App) Stop(selector Selector) error {
	return app.SetUnits(selector, 0)
}

// Start starts processes specified by the selector.
// We start a process by setting its unit quantity to 1.
// If a process has running units, nothing will be changed.
func (app *App) Start(selector Selector) error {
	deploymentFound := false
	units := 1
	for _, deploymentSpec := range app.Spec.Deployments {
		if selector.DeploymentVersion != nil && *selector.DeploymentVersion != deploymentSpec.Version {
			continue
		}
		if selector.Process != nil {
			for i, processSpec := range deploymentSpec.Processes {
				if processSpec.Name == *selector.Process && (processSpec.Units == nil || *processSpec.Units == 0) {
					deploymentSpec.Processes[i].Units = &units
				}
			}
		} else {
			for i, processSpec := range deploymentSpec.Processes {
				if processSpec.Units != nil && *processSpec.Units > 1 {
					continue
				}
				deploymentSpec.Processes[i].Units = &units
			}
		}
		deploymentFound = true
	}
	if selector.DeploymentVersion != nil && !deploymentFound {
		return ErrDeploymentNotFound
	}
	return nil
}

// CNames returns all CNAMEs to access the application including a default cname.
func (app *App) CNames(pool *Pool) []string {
	scheme := "http"
	if len(pool.Spec.IngressController.ClusterIssuer) > 0 {
		scheme = "https"
	}
	cnames := []string{}
	defaultCname := app.DefaultCname(pool)
	if defaultCname != nil {
		cnames = append(cnames, fmt.Sprintf("http://%s", *defaultCname))
	}
	for _, cname := range app.Spec.Ingress.Cnames {
		cnames = append(cnames, fmt.Sprintf("%s://%s", scheme, cname))
	}
	return cnames
}

// DefaultCname returns a default cname to access the application.
// A default cname uses the following format: <app name>.<pool's ServiceEndpoint>.shipa.cloud.
func (app *App) DefaultCname(pool *Pool) *string {
	if pool == nil {
		return nil
	}
	if !app.Spec.Ingress.GenerateDefaultCname {
		return nil
	}
	if len(pool.Spec.IngressController.ServiceEndpoint) == 0 {
		return nil
	}
	url := fmt.Sprintf("%s.%s.%s", app.Name, pool.Spec.IngressController.ServiceEndpoint, ShipaCloudDomain)
	return &url
}

// TemplatesConfigMapName returns a name of a configmap that contains templates used to render a helm chart.
func (app *App) TemplatesConfigMapName(ingressControllerType IngressControllerType) string {
	return templates.IngressConfigMapName(ingressControllerType.String())
}

// Units returns a total number units.
func (app *App) Units() int {
	units := 0
	for _, deploymentSpec := range app.Spec.Deployments {
		for _, processSpec := range deploymentSpec.Processes {
			if processSpec.Units == nil {
				units += 1
			} else {
				units += *processSpec.Units
			}
		}
	}
	return units
}

// ExposedPorts returns ports exposed by an image of each deployment.
func (app *App) ExposedPorts() map[DeploymentVersion][]ExposedPort {
	ports := make(map[DeploymentVersion][]ExposedPort, len(app.Spec.Deployments))
	for _, deployment := range app.Spec.Deployments {
		ports[deployment.Version] = deployment.ExposedPorts
	}
	return ports
}

// SetCondition sets Status and Message fields of the given type of condition to the provided values.
func (app *App) SetCondition(t AppConditionType, status v1.ConditionStatus, message string, time metav1.Time) {
	c := AppCondition{
		Type:               t,
		Status:             status,
		LastTransitionTime: &time,
		Message:            message,
	}
	for i, cond := range app.Status.Conditions {
		if cond.Type == t {
			if cond.Status == c.Status && cond.Message == c.Message {
				return
			}
			app.Status.Conditions[i] = c
			return
		}
	}
	app.Status.Conditions = append(app.Status.Conditions, c)
}

// Phase return a simple, high-level summary of where the application is in its lifecycle.
func (app *App) Phase() AppPhase {
	for _, cond := range app.Status.Conditions {
		if cond.Status == v1.ConditionFalse {
			return AppError
		}
	}
	if app.Units() == 0 {
		return AppCreated
	}
	return AppRunning
}

// Condition looks for a condition with the provided type in the condition list and returns it.
func (s AppStatus) Condition(t AppConditionType) *AppCondition {
	for _, c := range s.Conditions {
		if c.Type == t {
			return &c
		}
	}
	return nil
}
