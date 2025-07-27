package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Namespace    string `yaml:"namespace"`
	StorageClass string `yaml:"storageClass"`
	Domain       string `yaml:"domain"`

	Services struct {
		Jellyfin    ServiceConfig `yaml:"jellyfin"`
		Sonarr      ServiceConfig `yaml:"sonarr"`
		Radarr      ServiceConfig `yaml:"radarr"`
		Jackett     ServiceConfig `yaml:"jackett"`
		QBittorrent ServiceConfig `yaml:"qbittorrent"`
	} `yaml:"services"`

	Storage     StorageConfig     `yaml:"storage"`
	Ingress     IngressConfig     `yaml:"ingress"`
	CertManager CertManagerConfig `yaml:"certManager"`
}

type ServiceConfig struct {
	Enabled     bool              `yaml:"enabled"`
	Exposed     bool              `yaml:"exposed"` // Nouveau : contrôle l'exposition via ingress
	Image       string            `yaml:"image"`
	Tag         string            `yaml:"tag"`
	Port        int32             `yaml:"port"`
	Resources   ResourcesConfig   `yaml:"resources"`
	Environment map[string]string `yaml:"environment"`
	Volumes     []VolumeConfig    `yaml:"volumes"`
}

type ResourcesConfig struct {
	Requests struct {
		CPU    string `yaml:"cpu"`
		Memory string `yaml:"memory"`
	} `yaml:"requests"`
	Limits struct {
		CPU    string `yaml:"cpu"`
		Memory string `yaml:"memory"`
	} `yaml:"limits"`
}

type VolumeConfig struct {
	Name      string `yaml:"name"`
	MountPath string `yaml:"mountPath"`
	Size      string `yaml:"size"`
	ReadOnly  bool   `yaml:"readOnly"`
}

type StorageConfig struct {
	Media struct {
		Size        string   `yaml:"size"`
		AccessModes []string `yaml:"accessModes"`
	} `yaml:"media"`
	Downloads struct {
		Size        string   `yaml:"size"`
		AccessModes []string `yaml:"accessModes"`
	} `yaml:"downloads"`
}

type IngressConfig struct {
	Enabled     bool              `yaml:"enabled"`
	ClassName   string            `yaml:"className"`
	Annotations map[string]string `yaml:"annotations"`
	TLS         struct {
		Enabled    bool   `yaml:"enabled"`
		SecretName string `yaml:"secretName"`
	} `yaml:"tls"`
}

type CertManagerConfig struct {
	Enabled bool `yaml:"enabled"`
	Issuer  struct {
		Name  string `yaml:"name"`
		Type  string `yaml:"type"` // "letsencrypt" ou "selfsigned"
		Email string `yaml:"email,omitempty"`
	} `yaml:"issuer"`
}

func Load(filename string) (*Config, error) {
	// Configuration par défaut
	cfg := getDefaultConfig()

	// Charger depuis le fichier s'il existe
	if _, err := os.Stat(filename); err == nil {
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func getDefaultConfig() *Config {
	return &Config{
		Namespace:    "teleflix",
		StorageClass: "default",
		Domain:       "teleflix.local",
		Services: struct {
			Jellyfin    ServiceConfig `yaml:"jellyfin"`
			Sonarr      ServiceConfig `yaml:"sonarr"`
			Radarr      ServiceConfig `yaml:"radarr"`
			Jackett     ServiceConfig `yaml:"jackett"`
			QBittorrent ServiceConfig `yaml:"qbittorrent"`
		}{
			Jellyfin: ServiceConfig{
				Enabled: true,
				Exposed: true, // Jellyfin exposé par défaut
				Image:   "jellyfin/jellyfin",
				Tag:     "latest",
				Port:    8096,
				Resources: ResourcesConfig{
					Requests: struct {
						CPU    string `yaml:"cpu"`
						Memory string `yaml:"memory"`
					}{CPU: "500m", Memory: "512Mi"},
					Limits: struct {
						CPU    string `yaml:"cpu"`
						Memory string `yaml:"memory"`
					}{CPU: "2", Memory: "2Gi"},
				},
				Volumes: []VolumeConfig{
					{Name: "media", MountPath: "/media", ReadOnly: true},
					{Name: "config", MountPath: "/config", Size: "1Gi"},
				},
			},
			Sonarr: ServiceConfig{
				Enabled: true,
				Exposed: false, // Sonarr non exposé par défaut (sensible)
				Image:   "linuxserver/sonarr",
				Tag:     "latest",
				Port:    8989,
				Resources: ResourcesConfig{
					Requests: struct {
						CPU    string `yaml:"cpu"`
						Memory string `yaml:"memory"`
					}{CPU: "100m", Memory: "256Mi"},
					Limits: struct {
						CPU    string `yaml:"cpu"`
						Memory string `yaml:"memory"`
					}{CPU: "500m", Memory: "512Mi"},
				},
				Environment: map[string]string{
					"PUID": "1000",
					"PGID": "1000",
					"TZ":   "Europe/Paris",
				},
				Volumes: []VolumeConfig{
					{Name: "config", MountPath: "/config", Size: "1Gi"},
					{Name: "downloads", MountPath: "/downloads"},
					{Name: "media", MountPath: "/tv"},
				},
			},
			Radarr: ServiceConfig{
				Enabled: true,
				Exposed: false, // Radarr non exposé par défaut (sensible)
				Image:   "linuxserver/radarr",
				Tag:     "latest",
				Port:    7878,
				Resources: ResourcesConfig{
					Requests: struct {
						CPU    string `yaml:"cpu"`
						Memory string `yaml:"memory"`
					}{CPU: "100m", Memory: "256Mi"},
					Limits: struct {
						CPU    string `yaml:"cpu"`
						Memory string `yaml:"memory"`
					}{CPU: "500m", Memory: "512Mi"},
				},
				Environment: map[string]string{
					"PUID": "1000",
					"PGID": "1000",
					"TZ":   "Europe/Paris",
				},
				Volumes: []VolumeConfig{
					{Name: "config", MountPath: "/config", Size: "1Gi"},
					{Name: "downloads", MountPath: "/downloads"},
					{Name: "media", MountPath: "/movies"},
				},
			},
			Jackett: ServiceConfig{
				Enabled: true,
				Exposed: false, // Jackett non exposé par défaut (très sensible)
				Image:   "linuxserver/jackett",
				Tag:     "latest",
				Port:    9117,
				Resources: ResourcesConfig{
					Requests: struct {
						CPU    string `yaml:"cpu"`
						Memory string `yaml:"memory"`
					}{CPU: "100m", Memory: "128Mi"},
					Limits: struct {
						CPU    string `yaml:"cpu"`
						Memory string `yaml:"memory"`
					}{CPU: "200m", Memory: "256Mi"},
				},
				Environment: map[string]string{
					"PUID": "1000",
					"PGID": "1000",
					"TZ":   "Europe/Paris",
				},
				Volumes: []VolumeConfig{
					{Name: "config", MountPath: "/config", Size: "500Mi"},
				},
			},
			QBittorrent: ServiceConfig{
				Enabled: true,
				Exposed: false, // qBittorrent non exposé par défaut (risque sécurité)
				Image:   "linuxserver/qbittorrent",
				Tag:     "latest",
				Port:    8080,
				Resources: ResourcesConfig{
					Requests: struct {
						CPU    string `yaml:"cpu"`
						Memory string `yaml:"memory"`
					}{CPU: "200m", Memory: "512Mi"},
					Limits: struct {
						CPU    string `yaml:"cpu"`
						Memory string `yaml:"memory"`
					}{CPU: "1", Memory: "1Gi"},
				},
				Environment: map[string]string{
					"PUID":       "1000",
					"PGID":       "1000",
					"TZ":         "Europe/Paris",
					"WEBUI_PORT": "8080",
				},
				Volumes: []VolumeConfig{
					{Name: "config", MountPath: "/config", Size: "1Gi"},
					{Name: "downloads", MountPath: "/downloads"},
				},
			},
		},
		Storage: StorageConfig{
			Media: struct {
				Size        string   `yaml:"size"`
				AccessModes []string `yaml:"accessModes"`
			}{
				Size:        "100Gi",
				AccessModes: []string{"ReadWriteMany"},
			},
			Downloads: struct {
				Size        string   `yaml:"size"`
				AccessModes []string `yaml:"accessModes"`
			}{
				Size:        "50Gi",
				AccessModes: []string{"ReadWriteOnce"},
			},
		},
		Ingress: IngressConfig{
			Enabled:   true,
			ClassName: "traefik", // Par défaut pour k3s/Rancher
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target": "/",
			},
			TLS: struct {
				Enabled    bool   `yaml:"enabled"`
				SecretName string `yaml:"secretName"`
			}{
				Enabled:    false, // Désactivé par défaut, à activer manuellement
				SecretName: "teleflix-tls",
			},
		},
		CertManager: CertManagerConfig{
			Enabled: false, // Désactivé par défaut
			Issuer: struct {
				Name  string `yaml:"name"`
				Type  string `yaml:"type"`
				Email string `yaml:"email,omitempty"`
			}{
				Name:  "teleflix-issuer",
				Type:  "letsencrypt", // "letsencrypt" ou "selfsigned"
				Email: "",            // À remplir par l'utilisateur
			},
		},
	}
}
