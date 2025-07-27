package generator

import (
	"fmt"
	"strings"

	"teleflix/internal/config"
	"teleflix/internal/k8s"

	"gopkg.in/yaml.v3"
)

type Generator struct {
	config *config.Config
}

func New(cfg *config.Config) *Generator {
	return &Generator{config: cfg}
}

func (g *Generator) GenerateAll() (map[string]string, error) {
	manifests := make(map[string]string)

	// Générer le namespace
	ns := g.generateNamespace()
	manifests["00-namespace"] = ns

	// Générer les PVC
	pvcManifests, err := g.generatePVCs()
	if err != nil {
		return nil, err
	}
	manifests["01-storage"] = pvcManifests

	// Générer cert-manager resources si activé
	if g.config.CertManager.Enabled {
		certManagerManifests, err := g.generateCertManager()
		if err != nil {
			return nil, err
		}
		manifests["02-cert-manager"] = certManagerManifests
	}

	// Générer les services
	services := []struct {
		name   string
		config config.ServiceConfig
	}{
		{"jellyfin", g.config.Services.Jellyfin},
		{"sonarr", g.config.Services.Sonarr},
		{"radarr", g.config.Services.Radarr},
		{"jackett", g.config.Services.Jackett},
		{"qbittorrent", g.config.Services.QBittorrent},
	}

	serviceIndex := 3
	for _, svc := range services {
		if !svc.config.Enabled {
			continue
		}

		manifest, err := g.generateService(svc.name, svc.config)
		if err != nil {
			return nil, err
		}
		manifests[fmt.Sprintf("%02d-%s", serviceIndex, svc.name)] = manifest
		serviceIndex++
	}

	// Générer l'ingress seulement s'il y a des services exposés
	if g.config.Ingress.Enabled {
		ingress, err := g.generateIngress()
		if err != nil {
			return nil, err
		}
		// Ne pas ajouter l'ingress s'il est vide (aucun service exposé)
		if ingress != "" {
			manifests["99-ingress"] = ingress
		}
	}

	return manifests, nil
}

func (g *Generator) generateNamespace() string {
	ns := &k8s.Namespace{
		TypeMeta: k8s.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: k8s.ObjectMeta{
			Name: g.config.Namespace,
		},
	}

	data, _ := yaml.Marshal(ns)
	return string(data)
}

func (g *Generator) generatePVCs() (string, error) {
	var manifests []string

	// PVC Media
	mediaPVC := &k8s.PersistentVolumeClaim{
		TypeMeta: k8s.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: k8s.ObjectMeta{
			Name:      "media-pvc",
			Namespace: g.config.Namespace,
		},
		Spec: k8s.PersistentVolumeClaimSpec{
			AccessModes: g.config.Storage.Media.AccessModes,
			Resources: k8s.ResourceRequirements{
				Requests: map[string]string{
					"storage": g.config.Storage.Media.Size,
				},
			},
			StorageClassName: &g.config.StorageClass,
		},
	}

	// PVC Downloads
	downloadsPVC := &k8s.PersistentVolumeClaim{
		TypeMeta: k8s.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: k8s.ObjectMeta{
			Name:      "downloads-pvc",
			Namespace: g.config.Namespace,
		},
		Spec: k8s.PersistentVolumeClaimSpec{
			AccessModes: g.config.Storage.Downloads.AccessModes,
			Resources: k8s.ResourceRequirements{
				Requests: map[string]string{
					"storage": g.config.Storage.Downloads.Size,
				},
			},
			StorageClassName: &g.config.StorageClass,
		},
	}

	mediaData, _ := yaml.Marshal(mediaPVC)
	downloadsData, _ := yaml.Marshal(downloadsPVC)

	manifests = append(manifests, string(mediaData), "---", string(downloadsData))

	return strings.Join(manifests, "\n"), nil
}

func (g *Generator) generateCertManager() (string, error) {
	var manifests []string

	// Générer le ClusterIssuer
	var clusterIssuer *k8s.ClusterIssuer

	if g.config.CertManager.Issuer.Type == "letsencrypt" {
		if g.config.CertManager.Issuer.Email == "" {
			return "", fmt.Errorf("email requis pour Let's Encrypt")
		}

		clusterIssuer = &k8s.ClusterIssuer{
			TypeMeta: k8s.TypeMeta{
				APIVersion: "cert-manager.io/v1",
				Kind:       "ClusterIssuer",
			},
			ObjectMeta: k8s.ObjectMeta{
				Name: g.config.CertManager.Issuer.Name,
			},
			Spec: k8s.ClusterIssuerSpec{
				ACME: &k8s.ACMEIssuer{
					Server: "https://acme-v02.api.letsencrypt.org/directory",
					Email:  g.config.CertManager.Issuer.Email,
					PrivateKeySecretRef: k8s.SecretKeyRef{
						Name: g.config.CertManager.Issuer.Name + "-key",
					},
					Solvers: []k8s.ACMESolver{
						{
							HTTP01: &k8s.HTTP01Solver{
								Ingress: &k8s.HTTP01IngressSolver{
									Class: g.config.Ingress.ClassName,
								},
							},
						},
					},
				},
			},
		}
	} else if g.config.CertManager.Issuer.Type == "selfsigned" {
		clusterIssuer = &k8s.ClusterIssuer{
			TypeMeta: k8s.TypeMeta{
				APIVersion: "cert-manager.io/v1",
				Kind:       "ClusterIssuer",
			},
			ObjectMeta: k8s.ObjectMeta{
				Name: g.config.CertManager.Issuer.Name,
			},
			Spec: k8s.ClusterIssuerSpec{
				SelfSigned: &k8s.SelfSignedIssuer{},
			},
		}
	} else {
		return "", fmt.Errorf("type d'issuer non supporté: %s", g.config.CertManager.Issuer.Type)
	}

	clusterIssuerData, err := yaml.Marshal(clusterIssuer)
	if err != nil {
		return "", err
	}
	manifests = append(manifests, string(clusterIssuerData))

	// Générer le Certificate si TLS est activé
	if g.config.Ingress.TLS.Enabled {
		dnsNames := []string{}

		// Ajouter Jellyfin si activé ET exposé
		if g.config.Services.Jellyfin.Enabled && g.config.Services.Jellyfin.Exposed {
			dnsNames = append(dnsNames, fmt.Sprintf("jellyfin.%s", g.config.Domain))
		}

		// Ajouter les autres services s'ils sont activés ET exposés
		services := map[string]config.ServiceConfig{
			"sonarr":      g.config.Services.Sonarr,
			"radarr":      g.config.Services.Radarr,
			"jackett":     g.config.Services.Jackett,
			"qbittorrent": g.config.Services.QBittorrent,
		}

		for serviceName, cfg := range services {
			if cfg.Enabled && cfg.Exposed {
				dnsNames = append(dnsNames, fmt.Sprintf("%s.%s", serviceName, g.config.Domain))
			}
		}

		// Ne créer le certificat que s'il y a des domaines à couvrir
		if len(dnsNames) > 0 {
			certificate := &k8s.Certificate{
				TypeMeta: k8s.TypeMeta{
					APIVersion: "cert-manager.io/v1",
					Kind:       "Certificate",
				},
				ObjectMeta: k8s.ObjectMeta{
					Name:      "teleflix-certificate", // Nom unique du certificat
					Namespace: g.config.Namespace,
				},
				Spec: k8s.CertificateSpec{
					SecretName: g.config.Ingress.TLS.SecretName, // Utilise le nom du secret configuré
					IssuerRef: k8s.IssuerRef{
						Name: g.config.CertManager.Issuer.Name,
						Kind: "ClusterIssuer",
					},
					DNSNames: dnsNames,
				},
			}

			certificateData, err := yaml.Marshal(certificate)
			if err != nil {
				return "", err
			}
			manifests = append(manifests, "---", string(certificateData))
		}
	}

	return strings.Join(manifests, "\n"), nil
}

func (g *Generator) generateService(name string, cfg config.ServiceConfig) (string, error) {
	var manifests []string

	// Générer les PVC pour les volumes spécifiques au service
	for _, vol := range cfg.Volumes {
		if vol.Size != "" {
			pvc := &k8s.PersistentVolumeClaim{
				TypeMeta: k8s.TypeMeta{
					APIVersion: "v1",
					Kind:       "PersistentVolumeClaim",
				},
				ObjectMeta: k8s.ObjectMeta{
					Name:      fmt.Sprintf("%s-%s-pvc", name, vol.Name),
					Namespace: g.config.Namespace,
				},
				Spec: k8s.PersistentVolumeClaimSpec{
					AccessModes: []string{"ReadWriteOnce"},
					Resources: k8s.ResourceRequirements{
						Requests: map[string]string{
							"storage": vol.Size,
						},
					},
					StorageClassName: &g.config.StorageClass,
				},
			}

			data, _ := yaml.Marshal(pvc)
			manifests = append(manifests, string(data), "---")
		}
	}

	// Deployment
	deployment := g.createDeployment(name, cfg)
	deploymentData, _ := yaml.Marshal(deployment)
	manifests = append(manifests, string(deploymentData), "---")

	// Service
	service := g.createService(name, cfg)
	serviceData, _ := yaml.Marshal(service)
	manifests = append(manifests, string(serviceData))

	return strings.Join(manifests, "\n"), nil
}

func (g *Generator) createDeployment(name string, cfg config.ServiceConfig) *k8s.Deployment {
	labels := map[string]string{
		"app":       name,
		"component": "teleflix",
	}

	// Construire les variables d'environnement
	var envVars []k8s.EnvVar
	for key, value := range cfg.Environment {
		envVars = append(envVars, k8s.EnvVar{
			Name:  key,
			Value: value,
		})
	}

	// Construire les volumes et volume mounts
	var volumes []k8s.Volume
	var volumeMounts []k8s.VolumeMount

	for _, vol := range cfg.Volumes {
		volumeMounts = append(volumeMounts, k8s.VolumeMount{
			Name:      vol.Name,
			MountPath: vol.MountPath,
			ReadOnly:  vol.ReadOnly,
		})

		// Déterminer le nom du PVC
		var pvcName string
		if vol.Name == "media" {
			pvcName = "media-pvc"
		} else if vol.Name == "downloads" {
			pvcName = "downloads-pvc"
		} else {
			pvcName = fmt.Sprintf("%s-%s-pvc", name, vol.Name)
		}

		volumes = append(volumes, k8s.Volume{
			Name: vol.Name,
			VolumeSource: k8s.VolumeSource{
				PersistentVolumeClaim: &k8s.PersistentVolumeClaimVolumeSource{
					ClaimName: pvcName,
				},
			},
		})
	}

	return &k8s.Deployment{
		TypeMeta: k8s.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: k8s.ObjectMeta{
			Name:      name,
			Namespace: g.config.Namespace,
			Labels:    labels,
		},
		Spec: k8s.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &k8s.LabelSelector{
				MatchLabels: labels,
			},
			Template: k8s.PodTemplateSpec{
				ObjectMeta: k8s.ObjectMeta{
					Labels: labels,
				},
				Spec: k8s.PodSpec{
					Containers: []k8s.Container{
						{
							Name:  name,
							Image: fmt.Sprintf("%s:%s", cfg.Image, cfg.Tag),
							Ports: []k8s.ContainerPort{
								{
									ContainerPort: cfg.Port,
									Protocol:      "TCP",
								},
							},
							Env:          envVars,
							VolumeMounts: volumeMounts,
							Resources: k8s.ResourceRequirements{
								Requests: map[string]string{
									"cpu":    cfg.Resources.Requests.CPU,
									"memory": cfg.Resources.Requests.Memory,
								},
								Limits: map[string]string{
									"cpu":    cfg.Resources.Limits.CPU,
									"memory": cfg.Resources.Limits.Memory,
								},
							},
						},
					},
					Volumes: volumes,
				},
			},
		},
	}
}

func (g *Generator) createService(name string, cfg config.ServiceConfig) *k8s.Service {
	labels := map[string]string{
		"app":       name,
		"component": "teleflix",
	}

	return &k8s.Service{
		TypeMeta: k8s.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: k8s.ObjectMeta{
			Name:      name,
			Namespace: g.config.Namespace,
			Labels:    labels,
		},
		Spec: k8s.ServiceSpec{
			Selector: labels,
			Ports: []k8s.ServicePort{
				{
					Port:       cfg.Port,
					TargetPort: int32(cfg.Port),
					Protocol:   "TCP",
				},
			},
		},
	}
}

func (g *Generator) generateIngress() (string, error) {
	rules := []k8s.IngressRule{}

	// Ajouter Jellyfin si activé ET exposé
	if g.config.Services.Jellyfin.Enabled && g.config.Services.Jellyfin.Exposed {
		rules = append(rules, k8s.IngressRule{
			Host: fmt.Sprintf("jellyfin.%s", g.config.Domain),
			IngressRuleValue: k8s.IngressRuleValue{
				HTTP: &k8s.HTTPIngressRuleValue{
					Paths: []k8s.HTTPIngressPath{
						{
							Path:     "/",
							PathType: pathTypePtr("Prefix"),
							Backend: k8s.IngressBackend{
								Service: &k8s.IngressServiceBackend{
									Name: "jellyfin",
									Port: k8s.ServiceBackendPort{
										Number: g.config.Services.Jellyfin.Port,
									},
								},
							},
						},
					},
				},
			},
		})
	}

	// Ajouter les autres services s'ils sont activés ET exposés
	services := map[string]config.ServiceConfig{
		"sonarr":      g.config.Services.Sonarr,
		"radarr":      g.config.Services.Radarr,
		"jackett":     g.config.Services.Jackett,
		"qbittorrent": g.config.Services.QBittorrent,
	}

	for serviceName, cfg := range services {
		if cfg.Enabled && cfg.Exposed {
			rules = append(rules, k8s.IngressRule{
				Host: fmt.Sprintf("%s.%s", serviceName, g.config.Domain),
				IngressRuleValue: k8s.IngressRuleValue{
					HTTP: &k8s.HTTPIngressRuleValue{
						Paths: []k8s.HTTPIngressPath{
							{
								Path:     "/",
								PathType: pathTypePtr("Prefix"),
								Backend: k8s.IngressBackend{
									Service: &k8s.IngressServiceBackend{
										Name: serviceName,
										Port: k8s.ServiceBackendPort{
											Number: cfg.Port,
										},
									},
								},
							},
						},
					},
				},
			})
		}
	}

	// Si aucun service n'est exposé, ne pas créer d'ingress
	if len(rules) == 0 {
		return "", nil
	}

	// Préparer les annotations avec celles par défaut
	annotations := make(map[string]string)
	for k, v := range g.config.Ingress.Annotations {
		annotations[k] = v
	}

	// Ajouter les annotations TLS si activé (mais PAS cert-manager.io/cluster-issuer)
	if g.config.Ingress.TLS.Enabled && g.config.CertManager.Enabled {
		// Ne PAS ajouter cert-manager.io/cluster-issuer car on gère le Certificate explicitement

		// Ajouter redirection HTTPS selon l'ingress controller
		if g.config.Ingress.ClassName == "traefik" {
			if _, exists := annotations["traefik.ingress.kubernetes.io/redirect-to-https"]; !exists {
				annotations["traefik.ingress.kubernetes.io/redirect-to-https"] = "true"
			}
		} else if g.config.Ingress.ClassName == "nginx" {
			if _, exists := annotations["nginx.ingress.kubernetes.io/ssl-redirect"]; !exists {
				annotations["nginx.ingress.kubernetes.io/ssl-redirect"] = "true"
			}
			if _, exists := annotations["nginx.ingress.kubernetes.io/force-ssl-redirect"]; !exists {
				annotations["nginx.ingress.kubernetes.io/force-ssl-redirect"] = "true"
			}
		}
	}

	ingress := &k8s.Ingress{
		TypeMeta: k8s.TypeMeta{
			APIVersion: "networking.k8s.io/v1",
			Kind:       "Ingress",
		},
		ObjectMeta: k8s.ObjectMeta{
			Name:        "teleflix-ingress",
			Namespace:   g.config.Namespace,
			Annotations: annotations,
		},
		Spec: k8s.IngressSpec{
			IngressClassName: &g.config.Ingress.ClassName,
			Rules:            rules,
		},
	}

	// Ajouter TLS si activé
	if g.config.Ingress.TLS.Enabled {
		var hosts []string
		for _, rule := range rules {
			hosts = append(hosts, rule.Host)
		}

		ingress.Spec.TLS = []k8s.IngressTLS{
			{
				Hosts:      hosts,
				SecretName: g.config.Ingress.TLS.SecretName,
			},
		}
	}

	data, err := yaml.Marshal(ingress)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func int32Ptr(i int32) *int32 {
	return &i
}

func pathTypePtr(pt string) *string {
	return &pt
}
