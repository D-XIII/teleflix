package k8s

// Types Kubernetes de base
type TypeMeta struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
}

type ObjectMeta struct {
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// Namespace
type Namespace struct {
	TypeMeta   `yaml:",inline"`
	ObjectMeta `yaml:"metadata"`
}

// PersistentVolumeClaim
type PersistentVolumeClaim struct {
	TypeMeta   `yaml:",inline"`
	ObjectMeta `yaml:"metadata"`
	Spec       PersistentVolumeClaimSpec `yaml:"spec"`
}

type PersistentVolumeClaimSpec struct {
	AccessModes      []string             `yaml:"accessModes"`
	Resources        ResourceRequirements `yaml:"resources"`
	StorageClassName *string              `yaml:"storageClassName,omitempty"`
}

// Deployment
type Deployment struct {
	TypeMeta   `yaml:",inline"`
	ObjectMeta `yaml:"metadata"`
	Spec       DeploymentSpec `yaml:"spec"`
}

type DeploymentSpec struct {
	Replicas *int32          `yaml:"replicas,omitempty"`
	Selector *LabelSelector  `yaml:"selector"`
	Template PodTemplateSpec `yaml:"template"`
}

type LabelSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels,omitempty"`
}

type PodTemplateSpec struct {
	ObjectMeta `yaml:"metadata"`
	Spec       PodSpec `yaml:"spec"`
}

type PodSpec struct {
	Containers []Container `yaml:"containers"`
	Volumes    []Volume    `yaml:"volumes,omitempty"`
}

type Container struct {
	Name         string               `yaml:"name"`
	Image        string               `yaml:"image"`
	Ports        []ContainerPort      `yaml:"ports,omitempty"`
	Env          []EnvVar             `yaml:"env,omitempty"`
	VolumeMounts []VolumeMount        `yaml:"volumeMounts,omitempty"`
	Resources    ResourceRequirements `yaml:"resources,omitempty"`
}

type ContainerPort struct {
	ContainerPort int32  `yaml:"containerPort"`
	Protocol      string `yaml:"protocol,omitempty"`
}

type EnvVar struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type VolumeMount struct {
	Name      string `yaml:"name"`
	MountPath string `yaml:"mountPath"`
	ReadOnly  bool   `yaml:"readOnly,omitempty"`
}

type Volume struct {
	Name         string `yaml:"name"`
	VolumeSource `yaml:",inline"`
}

type VolumeSource struct {
	PersistentVolumeClaim *PersistentVolumeClaimVolumeSource `yaml:"persistentVolumeClaim,omitempty"`
}

type PersistentVolumeClaimVolumeSource struct {
	ClaimName string `yaml:"claimName"`
}

type ResourceRequirements struct {
	Requests map[string]string `yaml:"requests,omitempty"`
	Limits   map[string]string `yaml:"limits,omitempty"`
}

// Service
type Service struct {
	TypeMeta   `yaml:",inline"`
	ObjectMeta `yaml:"metadata"`
	Spec       ServiceSpec `yaml:"spec"`
}

type ServiceSpec struct {
	Selector map[string]string `yaml:"selector"`
	Ports    []ServicePort     `yaml:"ports"`
}

type ServicePort struct {
	Port       int32  `yaml:"port"`
	TargetPort int32  `yaml:"targetPort"`
	Protocol   string `yaml:"protocol,omitempty"`
}

// Ingress
type Ingress struct {
	TypeMeta   `yaml:",inline"`
	ObjectMeta `yaml:"metadata"`
	Spec       IngressSpec `yaml:"spec"`
}

type IngressSpec struct {
	IngressClassName *string       `yaml:"ingressClassName,omitempty"`
	TLS              []IngressTLS  `yaml:"tls,omitempty"`
	Rules            []IngressRule `yaml:"rules"`
}

type IngressTLS struct {
	Hosts      []string `yaml:"hosts"`
	SecretName string   `yaml:"secretName"`
}

type IngressRule struct {
	Host             string `yaml:"host"`
	IngressRuleValue `yaml:",inline"`
}

type IngressRuleValue struct {
	HTTP *HTTPIngressRuleValue `yaml:"http,omitempty"`
}

type HTTPIngressRuleValue struct {
	Paths []HTTPIngressPath `yaml:"paths"`
}

type HTTPIngressPath struct {
	Path     string         `yaml:"path"`
	PathType *string        `yaml:"pathType,omitempty"`
	Backend  IngressBackend `yaml:"backend"`
}

type IngressBackend struct {
	Service *IngressServiceBackend `yaml:"service,omitempty"`
}

type IngressServiceBackend struct {
	Name string             `yaml:"name"`
	Port ServiceBackendPort `yaml:"port"`
}

type ServiceBackendPort struct {
	Number int32 `yaml:"number"`
}

// Cert-Manager Types
type ClusterIssuer struct {
	TypeMeta   `yaml:",inline"`
	ObjectMeta `yaml:"metadata"`
	Spec       ClusterIssuerSpec `yaml:"spec"`
}

type ClusterIssuerSpec struct {
	ACME       *ACMEIssuer       `yaml:"acme,omitempty"`
	SelfSigned *SelfSignedIssuer `yaml:"selfSigned,omitempty"`
}

type ACMEIssuer struct {
	Server              string       `yaml:"server"`
	Email               string       `yaml:"email"`
	PrivateKeySecretRef SecretKeyRef `yaml:"privateKeySecretRef"`
	Solvers             []ACMESolver `yaml:"solvers"`
}

type SelfSignedIssuer struct{}

type SecretKeyRef struct {
	Name string `yaml:"name"`
}

type ACMESolver struct {
	HTTP01 *HTTP01Solver `yaml:"http01,omitempty"`
}

type HTTP01Solver struct {
	Ingress *HTTP01IngressSolver `yaml:"ingress,omitempty"`
}

type HTTP01IngressSolver struct {
	Class string `yaml:"class,omitempty"`
}

type Certificate struct {
	TypeMeta   `yaml:",inline"`
	ObjectMeta `yaml:"metadata"`
	Spec       CertificateSpec `yaml:"spec"`
}

type CertificateSpec struct {
	SecretName string    `yaml:"secretName"`
	IssuerRef  IssuerRef `yaml:"issuerRef"`
	DNSNames   []string  `yaml:"dnsNames"`
}

type IssuerRef struct {
	Name string `yaml:"name"`
	Kind string `yaml:"kind"`
}
